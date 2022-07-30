package repository

import (
	"context"

	"github.com/moguchev/BD-Forum/internal/app/models"
)

type User interface {
	GetUsersByNicknameOrEmail(ctx context.Context, nickname, email string) ([]models.User, error)
	CreateUser(ctx context.Context, u models.User) (models.User, error)
}
