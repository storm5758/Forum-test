package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/moguchev/BD-Forum/internal/app/models"
)

func (r *Repository) GetUsersByNicknameOrEmail(ctx context.Context, nickname, email string) ([]models.User, error) {
	query, args, err := squirrel.Select("nickname, email, full_name, about").
		From("users").
		Where(squirrel.Or{
			squirrel.Eq{
				"nickname": strings.ToLower(nickname),
			},
			squirrel.Eq{
				"email": email,
			},
		}).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("Repository.GetUsersByNiknameOrEmail: to sql: %w", err)
	}

	var users []models.User
	if err := pgxscan.Select(ctx, r.pool, &users, query, args...); err != nil {
		return nil, fmt.Errorf("Repository.GetUsersByNiknameOrEmail: select: %w", err)
	}

	return users, nil
}

func (r *Repository) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	query, args, err := squirrel.Insert("users").
		Columns("nickname, email, full_name, about").
		Values(strings.ToLower(user.Nickname), user.Email, user.Fullname, user.About).
		Suffix("RETURNING nickname").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("Repository.UserCreate: to sql: %w", err)
	}

	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&user.Nickname); err != nil {
		return models.User{}, fmt.Errorf("Repository.UserCreate: insert: %w", err)
	}

	return user, nil
}

// Создание нового пользователя
//
// Создание нового пользователя в базе данных.
// func (s *userService) UserCreate(context.Context, *api.UserCreateRequest) (*models.User, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method UserCreate not implemented")
// }
