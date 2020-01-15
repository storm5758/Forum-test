package repository

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/moguchev/BD-Forum/pkg/codes"
	"github.com/moguchev/BD-Forum/pkg/messages"
	. "github.com/moguchev/BD-Forum/pkg/models"
	"github.com/moguchev/BD-Forum/pkg/sql_queries"
)

func (r *Repository) CreateUser(u User) error {
	_, err := r.DbConn.Exec(sql_queries.InsertUser,
		u.About, u.Email, u.Fullname, u.Nickname)

	if err != nil {
		return fmt.Errorf(messages.UserAlreadyExists)
	}

	return nil
}

func (r *Repository) UpdateUser(u User) (User, error) {
	query := `UPDATE Users SET %s WHERE lower(nickname)=lower($1) `
	postfix := "RETURNING about, email, fullname, nickname;"

	paramCount := 1
	set := []string{}
	var params []interface{}
	params = append(params, u.Nickname)

	if u.About != "" {
		paramCount++
		set = append(set, "about=$"+strconv.Itoa(paramCount))
		params = append(params, u.About)
	}
	if u.Email != "" {
		paramCount++
		set = append(set, "email=$"+strconv.Itoa(paramCount))
		params = append(params, u.Email)
	}
	if u.Fullname != "" {
		paramCount++
		set = append(set, "fullname=$"+strconv.Itoa(paramCount))
		params = append(params, u.Fullname)
	}

	if paramCount <= 1 {
		return u, nil
	}

	query = fmt.Sprintf(query, strings.Join(set, ", "))
	query += postfix
	log.Printf(query)
	row := r.DbConn.QueryRowx(query, params...)

	var user User
	err := row.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
	if err != nil {
		e, _ := err.(pgx.PgError)

		switch e.Code {
		case codes.NotNullViolation, codes.ForeignKeyViolation:
			err = errors.New(messages.UserNotFound)
			break
		case codes.UniqueViolation:
			err = errors.New(messages.ConflictsInUserUpdate)
			break
		default:
			err = errors.New(messages.UserNotFound)
			break
		}
	}

	return user, err
}

func (r *Repository) GetUserByNickname(nickname string) (User, error) {

	row := r.DbConn.QueryRowx(sql_queries.SelectUserByNickname, nickname)

	var user User
	err := row.StructScan(&user)
	if err != nil {
		fmt.Println(err)
	}
	return user, err
}

func (r *Repository) GetUserByEmail(email string) (User, error) {

	row := r.DbConn.QueryRowx(sql_queries.SelectUserByEmail, email)

	var user User
	err := row.StructScan(&user)
	if err != nil {
		fmt.Println(err)
		return user, err
	}

	return user, nil
}

func (r *Repository) FindUsers(nickname string, email string) ([]User, error) {
	var users []User

	rows, err := r.DbConn.Queryx(sql_queries.SelectUsersByNicknameAndEmail, nickname, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := User{}
		err = rows.StructScan(&user)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}

	return users, nil
}
