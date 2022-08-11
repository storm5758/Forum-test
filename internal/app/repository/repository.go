package repository

import (
	"context"

	"github.com/storm5758/Forum-test/internal/app/models"
)

type User interface {
	GetUsersByNicknameOrEmail(ctx context.Context, nickname, email string) ([]models.User, error)
	CreateUser(ctx context.Context, u models.User) (models.User, error)
}
