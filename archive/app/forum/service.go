package forum

import (
	. "github.com/moguchev/BD-Forum/pkg/models"
)

type ServiceInterface interface {
	// user section
	CreateUser(User) ([]User, error)
	GetUser(string) (User, error)
	UpdateUser(User) (User, error)

	// service section
	Clear() error
	Status() (Status, error)

	// forum section
	CreateForum(NewForum) (Forum, error)
	GetForum(string) (Forum, error)
	CreateThread(Thread) (Thread, error)
	GetThreads(string, int64, string, bool) ([]Thread, error)
	GetUsersByForum(string, int64, string, bool) ([]User, error)

	// threads sectio
	CreatePosts(string, []Post) ([]Post, error)
	GetThreadById(int64) (Thread, error)
	GetThreadBySlug(string) (Thread, error)
	UpdateThread(Thread) (Thread, error)
	GetPosts(string, int64, string, string, bool) ([]Post, error)
	CreateVote(string, Vote) (Thread, error)
	// post
	GetPostAccount(int64, []string) (PostAccount, error)
	UpdatePost(int64, Post) (Post, error)
}
