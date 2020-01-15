package repository

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/jackc/pgtype"
	"github.com/moguchev/BD-Forum/pkg/messages"
	. "github.com/moguchev/BD-Forum/pkg/models"
	"github.com/moguchev/BD-Forum/pkg/mytools"
	"github.com/moguchev/BD-Forum/pkg/sql_queries"
)

var mutexMapMutex sync.Mutex
var getPostsTemplate *template.Template
var getPostsParentTreeTemplate *template.Template

func init() {
	mutexMapMutex = sync.Mutex{}
	var err error
	getPostsTemplate, err = template.New("getPosts").Parse(queryTemplateGetPostsSorted)
	if err != nil {
		fmt.Println("Error: cannot create getPostsTemplate template: ", err)
		panic(err)
	}

	getPostsParentTreeTemplate, err = template.New("parent_tree").Parse(queryTemplateGetPostsParentTree)
	if err != nil {
		fmt.Println("Error: cannot create getPostsParentTreeTemplate template: ", err)
		panic(err)
	}
}

const (
	queryTemplateGetPostsSorted = `SELECT author, forum, created, posts.id, isEdited, message, coalesce(parent, 0), thread 
				FROM posts
					{{.Condition}}
					ORDER BY {{.OrderBy}}
					{{.Limit}}`
	queryTemplateGetPostsParentTree = `JOIN (
						SELECT parents.id FROM posts AS parents
						WHERE parents.thread=$1 AND parents.parent IS NULL
							{{- if .Since}} AND {{.Since}}{{- end}}
						ORDER BY parents.path[1] {{.Desc}}
						{{.Limit}}
						) as p ON path[1]=p.id`
)

type queryArgs struct {
	Since string
	Desc  string
	Limit string
}

type comandArgs struct {
	Condition string
	OrderBy   string
	Limit     string
}

func (r *Repository) GetThreadById(id int64) (Thread, error) {
	var t Thread
	var created time.Time
	var slug pgtype.Text

	err := r.DbConn.QueryRowx(sql_queries.SelectThreadById, id).
		Scan(&slug, &t.Title, &t.Message, &t.Forum, &t.Author,
			&created, &t.Votes, &t.Id)

	t.Created = created.Format(time.RFC3339Nano)
	t.Slug = slug.String

	if err != nil {
		fmt.Println(err)
		err = errors.New(messages.ThreadNotFound)
	}

	return t, err
}

func (r *Repository) GetThreadBySlug(slug string) (Thread, error) {
	var t Thread

	var created time.Time
	var s pgtype.Text

	err := r.DbConn.QueryRowx(sql_queries.SelectThreadBySlug, slug).
		Scan(&s, &t.Title, &t.Message, &t.Forum, &t.Author,
			&created, &t.Votes, &t.Id)

	t.Created = created.Format(time.RFC3339Nano)
	t.Slug = s.String
	if err != nil {
		fmt.Println(err)
		err = errors.New(messages.ThreadNotFound)
	}

	return t, err
}

func (r *Repository) GetThreadId(slugOrId string) (int64, error) {
	var id int64 = 0
	_, err := strconv.ParseInt(slugOrId, 10, 64)
	if err != nil {
		// slug
		err := r.DbConn.QueryRowx(sql_queries.SelectThreadIdBySlug, slugOrId).Scan(&id)
		if err != nil {
			return id, errors.New(messages.ThreadNotFound)
		}
	} else {
		// id
		err := r.DbConn.QueryRowx(sql_queries.SelectThreadIdById, slugOrId).Scan(&id)
		if err != nil {
			return id, errors.New(messages.ThreadNotFound)
		}
	}
	return id, nil
}

func (r *Repository) GetThreadForumSlug(threadId int64) (string, error) {
	var forum string
	err := r.DbConn.QueryRowx(`SELECT forum FROM threads WHERE id=$1`, threadId).Scan(&forum)
	if err != nil {
		return forum, errors.New(messages.ThreadNotFound)
	}
	return forum, nil
}

func (r *Repository) CreatePostsByPacket(threadId int64, forumSLug string, posts []Post, created time.Time) ([]Post, error) {
	var params []interface{}
	for _, post := range posts {
		var parent sql.NullInt64
		parent.Int64 = post.Parent
		if post.Parent != 0 {
			parent.Valid = true
		}
		params = append(params, post.Author, post.Message, parent, threadId, created, forumSLug)
	}

	query := `INSERT INTO posts (author, message, parent, thread, created, forum) VALUES `
	postfix := `RETURNING forum, id, created`

	query = mytools.CreatePacketQuery(query, 6, len(posts), postfix)

	rows, err := r.DbConn.Queryx(query, params...)
	defer rows.Close()

	if err != nil || (rows != nil && rows.Err() != nil) {
		return posts, err
	}
	i := 0
	for rows.Next() {
		var created time.Time
		err := rows.Scan(&(posts[i].Forum), &(posts[i].Id), &(created))
		if err != nil {
			return posts, err
		}
		posts[i].Created = created.Format(time.RFC3339Nano)
		posts[i].IsEdited = false
		posts[i].Thread = int32(threadId)
		i++
	}

	var cnt int64
	if i == 0 && len(posts) > 0 {
		// looking for exact error
		if row := r.DbConn.QueryRowx(`SELECT count(id) from threads WHERE id=$1;`, threadId); row.Scan(&cnt) != nil || cnt == 0 {
			return posts, errors.New(messages.ThreadNotFound)
		} else if row := r.DbConn.QueryRowx(`SELECT COUNT(nickname) FROM users WHERE nickname=$1`, posts[0].Author); row.Scan(&cnt) != nil || cnt == 0 {
			return posts, errors.New(messages.ThreadNotFound)
		} else {
			return posts, errors.New(messages.ParentNotFound)
		}
	}
	return posts, nil
}

func (r *Repository) UpdateForumPosts(forum string, numPosts int) error {
	query := `UPDATE ForumPosts SET posts = posts + $2 WHERE forum = $1;`
	_, err := r.DbConn.Exec(query, forum, numPosts)
	if err == nil {
		atomic.AddInt32(&accessToForumPosts.numberOfNewUpdates, 1)
	}
	return err
}

func (r *Repository) InsertUsersToUsersInForum(users map[string]bool, forum string) error {
	prefix := `INSERT INTO UsersInForum(nickname, forum) VALUES `
	postfix := `ON CONFLICT DO NOTHING`
	query := mytools.CreatePacketQuery(prefix, 2, len(users), postfix)

	params := make([]interface{}, 0, len(users))
	for key := range users {
		params = append(params, key, forum)
	}

	mutexMapMutex.Lock()
	defer mutexMapMutex.Unlock()

	_, err := r.DbConn.Exec(query, params...)
	return err
}

func (r *Repository) GetPosts(threadID, limit int64, since string, sort string, desc bool) ([]Post, error) {
	posts := make([]Post, 0)

	temp := comandArgs{}

	params := make([]interface{}, 0, 2)
	params = append(params, threadID)

	var placeholderSince string
	placeholderDesc := "ASC"

	if desc {
		placeholderDesc = "DESC"
	}
	if limit != 0 {
		params = append(params, limit)
		temp.Limit = `LIMIT $` + strconv.Itoa(len(params))
	}
	if since != "" {
		params = append(params, since)
		compareSign := ">"
		if desc {
			compareSign = "<"
		}
		paramNum := len(params)
		queryGetPath := `SELECT %s FROM posts AS since WHERE since.id=%s`

		switch sort {
		case "flat":
			//	AND id > $n
			placeholderSince = fmt.Sprintf(`AND id%s$%d`, compareSign, paramNum)
		case "tree":
			//	AND path[&n] > (SELECT since.path from Posts AS since WHERE since.id=&n)
			placeholderSince = fmt.Sprintf(
				`AND path%s(%s)`,
				compareSign,
				fmt.Sprintf(queryGetPath, `since.path`, fmt.Sprintf(`$%d`, paramNum)),
			)
		case "parent_tree":
			//	AND parents[1] > (SELECT since.path[1] from Posts AS since WHERE since.id=&n)
			placeholderSince = fmt.Sprintf(
				`parents.path[1]%s(%s)`,
				compareSign,
				fmt.Sprintf(queryGetPath, `since.path[1]`, fmt.Sprintf(`$%d`, paramNum)),
			)
		}
	}

	switch sort {
	case "flat":
		temp.Condition = `WHERE thread=$1 ` + placeholderSince
		temp.OrderBy = fmt.Sprintf(`(created, id) %s`, placeholderDesc)
	case "tree":
		temp.Condition = `WHERE thread=$1 ` + placeholderSince
		temp.OrderBy = fmt.Sprintf(`(path, created) %s`, placeholderDesc)
	case "parent_tree":
		conditionBuffer := &bytes.Buffer{}
		err := getPostsParentTreeTemplate.Execute(conditionBuffer,
			queryArgs{Since: placeholderSince, Desc: placeholderDesc, Limit: temp.Limit})
		if err != nil {
			return posts, err
		}
		temp.Condition = conditionBuffer.String()
		temp.OrderBy = fmt.Sprintf(`path[1] %s, path`, placeholderDesc)
		temp.Limit = ""
	}

	queryBuffer := &bytes.Buffer{}
	err := getPostsTemplate.Execute(queryBuffer, temp)
	if err != nil {
		return posts, err
	}
	query := queryBuffer.String()

	rows, err := r.DbConn.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Author, &p.Forum, &p.Created, &p.Id, &p.IsEdited, &p.Message, &p.Parent, &p.Thread)
		if err != nil {
			return nil, err
		}

		posts = append(posts, p)
	}

	return posts, nil
}

func (r *Repository) CreateVote(id int64, v Vote) error {
	_, err := r.DbConn.Exec(sql_queries.InsertVote, id, v.Nickname, v.Voice)
	return err
}

func (r *Repository) UpdateThread(t Thread) (Thread, error) {
	var thread Thread

	query := `UPDATE threads SET %s WHERE %s=$1 `
	postfix := "RETURNING slug, title, message, forum, author, created, votes, id;"

	paramCount := 1
	set := []string{}
	var params []interface{}
	var key interface{}
	var keyName string

	if t.Id != 0 {
		key = t.Id
		keyName = "id"
	} else {
		key = t.Slug
		keyName = "slug"
	}
	params = append(params, key)

	if t.Message != "" {
		paramCount++
		set = append(set, "message=$"+strconv.Itoa(paramCount))
		params = append(params, t.Message)
	}
	if t.Title != "" {
		paramCount++
		set = append(set, "title=$"+strconv.Itoa(paramCount))
		params = append(params, t.Title)
	}

	if len(set) == 0 {
		if t.Id != 0 {
			th, err := r.GetThreadById(int64(t.Id))
			return th, err
		}
		th, err := r.GetThreadBySlug(t.Slug)
		return th, err
	}

	query = fmt.Sprintf(query, strings.Join(set, ", "), keyName)
	query += postfix

	row := r.DbConn.QueryRow(query, params...)

	var created time.Time
	var s sql.NullString
	err := row.Scan(&s, &thread.Title, &thread.Message, &thread.Forum,
		&thread.Author, &created, &thread.Votes, &thread.Id)
	if err != nil {
		return thread, err
	}
	if s.Valid {
		thread.Slug = s.String
	}
	thread.Created = created.Format(time.RFC3339Nano)

	return thread, err
}
