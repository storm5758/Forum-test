package repository

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx"
	"github.com/moguchev/BD-Forum/pkg/codes"
	"github.com/moguchev/BD-Forum/pkg/messages"
	. "github.com/moguchev/BD-Forum/pkg/models"
	"github.com/moguchev/BD-Forum/pkg/sql_queries"
)

var getForumUsersTemplate *template.Template

type Sync struct {
	mutex              sync.Mutex
	numberOfNewUpdates int32
}

var accessToForumPosts Sync

func init() {
	accessToForumPosts.mutex = sync.Mutex{}
}

func (r *Repository) CreateForum(nf NewForum) error {
	_, err := r.DbConn.Exec(sql_queries.InsertForum,
		nf.Slug, nf.Title, nf.User)

	if err != nil {
		if e, ok := err.(pgx.PgError); ok {
			switch e.Code {
			case codes.NotNullViolation, codes.ForeignKeyViolation:
				err = errors.New(messages.UserNotFound)
				break
			case codes.UniqueViolation:
				err = errors.New(messages.ForumAlreadyExists)
				break
			default:
				log.Println(e.Code)
			}
		}
		return err
	}
	if err == nil {
		_, err = r.DbConn.Exec(`INSERT INTO ForumPosts(forum, posts) VALUES ($1, 0)`, nf.Slug)
	}
	return nil
}

func (r *Repository) GetForum(slug string) (Forum, error) {
	var forum Forum

	if err := r.loadForumPosts(); err != nil {
		return forum, err
	}
	row := r.DbConn.QueryRowx(sql_queries.SelectForum, slug)

	err := row.StructScan(&forum)
	if err != nil {
		fmt.Println(err)
	}

	return forum, err
}

func (r *Repository) CreateThread(t Thread) (Thread, error) {
	var slug pgtype.Text
	time, _ := time.Parse(time.RFC3339Nano, t.Created)
	err := r.DbConn.QueryRow(sql_queries.InsertThread,
		t.Author, t.Forum, t.Message, time, t.Title, t.Slug).
		Scan(&t.Id, &slug, &t.Votes, &t.Forum)

	t.Slug = slug.String

	if err != nil {
		if e, ok := err.(pgx.PgError); ok {
			switch e.Code {
			case codes.ForeignKeyViolation, codes.NotNullViolation:
				if e.ConstraintName == "thread_author_fkey" {
					err = errors.New(messages.UserNotFound)
				} else {
					err = errors.New(messages.ForumNotFound)
				}
				break
			case codes.UniqueViolation:
				err = errors.New(messages.ThreadAlreadyExists)
				break
			default:
				log.Println(e.Code)
				err = errors.New(messages.ForumNotFound)
			}
		} else {
			err = errors.New(messages.ForumNotFound)
		}
	}

	return t, err
}

func (r *Repository) GetThreads(forum string, limit int64, since string, desc bool) ([]Thread, error) {
	threads := make([]Thread, 0)

	query := sql_queries.SelectThreadsByForum

	var err error

	params := make([]interface{}, 0, 2)
	params = append(params, forum)
	var placeholderSince, placeholderDesc, placeholderLimit string

	if since != "" {
		params = append(params, since)
		placeholderSince = `AND created>=$` + strconv.Itoa(len(params))
		if desc {
			placeholderSince = `AND created<=$` + strconv.Itoa(len(params))
		}
	}
	if desc {
		placeholderDesc = `DESC`
	}
	if limit != 0 {
		params = append(params, limit)
		placeholderLimit = `LIMIT $` + strconv.Itoa(len(params))
	}

	query = fmt.Sprintf(query, placeholderSince, placeholderDesc, placeholderLimit)
	log.Printf(query)

	log.Printf("FORUM: %s", forum)
	log.Printf("LIMIT: %d", limit)
	log.Printf("SINCE: %s", since)
	rows, err := r.DbConn.Queryx(query, params...)

	if err != nil {
		return threads, err
	}
	defer rows.Close()
	for rows.Next() {
		t := Thread{}

		slug := sql.NullString{}

		err = rows.Scan(&t.Author, &t.Forum, &t.Created, &t.Id, &t.Message, &slug, &t.Title, &t.Votes)

		if err != nil {
			log.Println(err)
			return threads, err
		}

		if slug.Valid {
			t.Slug = slug.String
		}

		threads = append(threads, t)
	}
	if len(threads) == 0 {
		_, e := r.GetForum(forum)
		if e != nil {
			return threads, errors.New(messages.ForumNotFound)
		}
	}
	return threads, nil
}

func (r *Repository) GetUsersByForum(forum string, limit int64, since string, desc bool) ([]User, error) {
	users := make([]User, 0)
	templateArgs := struct {
		Since string
		Limit string
		Desc  string
	}{}

	paramsCount := 1
	params := make([]interface{}, 0)
	params = append(params, forum)
	if desc {
		if since != "" {
			templateArgs.Since = `AND nickname<$2`
			paramsCount++
			params = append(params, since)
		}
		templateArgs.Desc = "DESC"
	} else {
		if since != "" {
			templateArgs.Since = `AND nickname>$2`
			paramsCount++
			params = append(params, since)
		}
		templateArgs.Desc = "ASC"
	}
	if limit != 0 {
		paramsCount++
		templateArgs.Limit = fmt.Sprintf(`LIMIT $%d`, paramsCount)
		params = append(params, limit)
	}
	queryBuf := &bytes.Buffer{}

	getForumUsersTemplate, _ = template.New("getForumUsers").Parse(sql_queries.QueryTemplateGetForumUsers)
	err := getForumUsersTemplate.Execute(queryBuf, templateArgs)
	if err != nil {
		return users, err
	}

	query := queryBuf.String()
	rows, err := r.DbConn.Queryx(query, params...)

	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		user := User{}
		err := rows.StructScan(&user)

		if err != nil {
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *Repository) loadForumPosts() error {
	if accessToForumPosts.numberOfNewUpdates == 0 {
		return nil
	}
	tx, err := r.DbConn.Begin()
	if err != nil {
		return err
	}

	accessToForumPosts.mutex.Lock()
	defer accessToForumPosts.mutex.Unlock()

	if accessToForumPosts.numberOfNewUpdates == 0 {
		return nil
	}
	err = func() error {
		_, err := tx.Exec(`UPDATE forums SET posts = forums.posts + temp.posts FROM ForumPosts as temp WHERE temp.forum = forums.slug`)

		if err != nil {
			return err
		}

		_, err = tx.Exec(`UPDATE ForumPosts SET posts = 0`)
		return err
	}()
	if err != nil {
		return tx.Rollback()
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	accessToForumPosts.numberOfNewUpdates = 0
	return nil
}
