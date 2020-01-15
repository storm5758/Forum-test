package repository

import (
	"fmt"

	. "github.com/moguchev/BD-Forum/pkg/models"
	"github.com/moguchev/BD-Forum/pkg/sql_queries"
)

func (r *Repository) Clear() error {
	_, err := r.DbConn.Exec(sql_queries.Truncate)
	return err
}

func (r *Repository) Status() (Status, error) {
	row := r.DbConn.QueryRowx(sql_queries.SelectAll)

	var data Status
	err := row.StructScan(&data)
	if err != nil {
		fmt.Println(err)
	}
	return data, err
}
