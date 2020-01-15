package service

import (
	"context"

	"github.com/moguchev/BD-Forum/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type adminService struct {
	api.UnimplementedAdminServer
}

func NewAdminService() api.AdminServer {
	return &adminService{}
}

// Очистка всех данных в базе
//
// Безвозвратное удаление всей пользовательской информации из базы данных.
func (s *adminService) Clear(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Clear not implemented")
}

// Получение инфомарции о базе данных
//
// Получение инфомарции о базе данных.
func (s *adminService) Status(context.Context, *emptypb.Empty) (*api.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
