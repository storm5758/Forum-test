package service

import (
	"github.com/moguchev/BD-Forum/pkg/messages"
	. "github.com/moguchev/BD-Forum/pkg/models"
)

func (s Service) CreateForum(nf NewForum) (Forum, error) {
	err := s.Repository.CreateForum(nf)

	f := Forum{}
	if err != nil {
		switch err.Error() {
		case messages.UserNotFound:
			return f, err
		case messages.ForumAlreadyExists:
			f, _ := s.Repository.GetForum(nf.Slug)
			return f, err
		}
	}

	f, _ = s.Repository.GetForum(nf.Slug)

	return f, nil
}

func (s Service) GetForum(slug string) (Forum, error) {
	forum, err := s.Repository.GetForum(slug)
	return forum, err
}

func (s Service) CreateThread(t Thread) (Thread, error) {
	thread, err := s.Repository.CreateThread(t)

	if err != nil {
		switch err.Error() {
		case messages.UserNotFound:
			break
		case messages.ForumNotFound:
			break
		case messages.ThreadAlreadyExists:
			thread, _ = s.Repository.GetThreadBySlug(t.Slug)
			break
		}
	}
	return thread, err
}

func (s Service) GetThreads(forum string, limit int64, since string, desc bool) ([]Thread, error) {
	ths, err := s.Repository.GetThreads(forum, limit, since, desc)
	return ths, err
}

func (s Service) GetUsersByForum(forum string, limit int64, since string, desc bool) ([]User, error) {
	us, err := s.Repository.GetUsersByForum(forum, limit, since, desc)
	return us, err
}
