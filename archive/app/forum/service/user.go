package service

import (
	"errors"

	"github.com/storm5758/Forum-test/pkg/messages"
)

func (s Service) CreateUser(u User) ([]User, error) {
	existedUsers, _ := s.Repository.FindUsers(u.Nickname, u.Email)

	if len(existedUsers) > 0 {
		return existedUsers, nil
	}

	err := s.Repository.CreateUser(u)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s Service) UpdateUser(u User) (User, error) {
	if u.About == "" && u.Email == "" && u.Fullname == "" {
		user, err := s.Repository.GetUserByNickname(u.Nickname)
		if err != nil {
			err = errors.New(messages.UserNotFound)
		}
		return user, err
	}
	return s.Repository.UpdateUser(u)
}

func (s Service) GetUser(nickname string) (User, error) {
	user, err := s.Repository.GetUserByNickname(nickname)
	return user, err
}
