package service

import (
	"context"

	"github.com/moguchev/BD-Forum/pkg/api"
	"github.com/moguchev/BD-Forum/pkg/api/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userService struct {
	api.UnimplementedUserServer
}

func NewUserService() api.UserServer {
	return &userService{}
}

// Создание нового пользователя
//
// Создание нового пользователя в базе данных.
func (s *userService) UserCreate(context.Context, *api.UserCreateRequest) (*models.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserCreate not implemented")
}

// Получение информации о пользователе
//
// Получение информации о пользователе форума по его имени.
func (s *userService) UserGetOne(context.Context, *api.UserGetOneRequest) (*models.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserGetOne not implemented")
}

// Изменение данных о пользователе
//
// Изменение информации в профиле пользователя.
func (s *userService) UserUpdate(context.Context, *api.UserUpdateRequest) (*models.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserUpdate not implemented")
}
