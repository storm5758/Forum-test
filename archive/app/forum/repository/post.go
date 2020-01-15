package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/moguchev/BD-Forum/pkg/models"
	"github.com/moguchev/BD-Forum/pkg/sql_queries"
)

func (r *Repository) GetPostById(id int64) (Post, error) {
	var p Post
	var created time.Time
	var parent sql.NullInt64
	err := r.DbConn.QueryRowx(sql_queries.SelectPostById, id).Scan(
		&p.Author, &created, &p.Forum, &p.Id, &p.Message, &p.Thread, &p.IsEdited, &parent,
	)
	if err != nil {
		return p, err
	}
	p.Created = created.Format(time.RFC3339Nano)
	p.Parent = 0
	if parent.Valid {
		p.Parent = parent.Int64
	}

	return p, err
}

func (r *Repository) UpdatePost(id int64, p Post) (Post, error) {
	query := `UPDATE posts SET %s WHERE id=$1`
	postfix := " RETURNING author, created, forum, id, isedited, message, parent, thread;"
	count := 1
	set := []string{}
	var params []interface{}
	params = append(params, id)
	if p.Message != "" {
		count++
		set = append(set, "message=$"+strconv.Itoa(count))
		params = append(params, p.Message)
	}
	if p.Parent != 0 {
		count++
		set = append(set, "parent=$"+strconv.Itoa(count))
		params = append(params, p.Parent)
	}
	if len(set) == 0 {
		return r.GetPostById(id)
	}
	query = fmt.Sprintf(query, strings.Join(set, ", "))
	query += postfix

	var post Post
	var created time.Time
	var parent sql.NullInt64
	row := r.DbConn.QueryRow(query, params...)
	err := row.Scan(&post.Author, &created, &post.Forum, &post.Id,
		&post.IsEdited, &post.Message, &parent, &post.Thread)
	if err != nil {
		return post, err
	}
	post.Created = created.Format(time.RFC3339Nano)
	post.Parent = 0
	if parent.Valid {
		post.Parent = parent.Int64
	}
	return post, err
}
