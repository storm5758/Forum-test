package service

import (
	"context"
	"log"

	"github.com/moguchev/BD-Forum/internal/app/models"
	"github.com/moguchev/BD-Forum/internal/app/repository"
	"github.com/moguchev/BD-Forum/pkg/api"
	api_models "github.com/moguchev/BD-Forum/pkg/api/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userService struct {
	api.UnimplementedUserServer
	Deps
}

type Deps struct {
	UserRepository repository.User // либо как потребитель объевляем свой интерфейс в пакете и реализации должны удовлетворять ему
}

func NewUserService(d Deps) api.UserServer {
	return &userService{
		Deps: d,
	}
}

// Создание нового пользователя
//
// Создание нового пользователя в базе данных.
func (s *userService) UserCreate(ctx context.Context, req *api.UserCreateRequest) (*api_models.User, error) {
	nikname := req.GetNickname()
	if len(nikname) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty nickname")
	}
	profile := req.GetProfile()
	if profile == nil {
		return nil, status.Error(codes.InvalidArgument, "empty profile")
	}
	email := profile.GetEmail()
	if len(email) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty email")
	}

	existedUsers, err := s.UserRepository.GetUsersByNicknameOrEmail(ctx, nikname, email)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	if len(existedUsers) > 0 {
		return nil, status.Error(codes.AlreadyExists, codes.AlreadyExists.String())
	}

	user := models.User{
		Nickname: nikname,
		Email:    email,
		Fullname: profile.GetFullname(),
		About:    profile.GetAbout(),
	}

	createdUser, err := s.UserRepository.CreateUser(ctx, user)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	return &api_models.User{
		About:    createdUser.About,
		Email:    createdUser.Email,
		Fullname: createdUser.Fullname,
		Nickname: createdUser.Nickname,
	}, nil
}

// Получение информации о пользователе
//
// Получение информации о пользователе форума по его имени.
func (s *userService) UserGetOne(context.Context, *api.UserGetOneRequest) (*api_models.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserGetOne not implemented")
}

// Изменение данных о пользователе
//
// Изменение информации в профиле пользователя.
func (s *userService) UserUpdate(context.Context, *api.UserUpdateRequest) (*api_models.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserUpdate not implemented")
}
