package repository

import (
	"fmt"
	"io/ioutil"

	"github.com/jmoiron/sqlx"
	"github.com/storm5758/Forum-test/pkg/config"
)

type Repository struct {
	DbConn *sqlx.DB
}

func (Rep *Repository) InitDBSQL() error {
	if Rep.DbConn == nil {
		return fmt.Errorf("Dead connection")
	}

	content, err := ioutil.ReadFile(config.DBSchema)
	if err != nil {
		return err
	}

	tx, err := Rep.DbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(string(content)); err != nil {
		return err
	}
	tx.Commit()
	return nil
}
