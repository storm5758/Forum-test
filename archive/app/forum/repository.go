package forum

import (
	"time"

	. "github.com/moguchev/BD-Forum/pkg/models"
)

type Repository interface {

	// user section
	CreateUser(User) error
	UpdateUser(User) (User, error)
	GetUserByNickname(string) (User, error)
	GetUserByEmail(string) (User, error)
	FindUsers(string, string) ([]User, error)

	// service section
	Clear() error
	Status() (Status, error)

	// forum section
	CreateForum(NewForum) error
	GetForum(string) (Forum, error)
	CreateThread(Thread) (Thread, error)
	GetThreadBySlug(string) (Thread, error)
	GetThreads(string, int64, string, bool) ([]Thread, error)
	GetUsersByForum(string, int64, string, bool) ([]User, error)

	// thread section
	GetThreadId(string) (int64, error)
	GetThreadForumSlug(int64) (string, error)
	CreatePostsByPacket(int64, string, []Post, time.Time) ([]Post, error)
	UpdateForumPosts(string, int) error
	InsertUsersToUsersInForum(map[string]bool, string) error
	GetThreadById(int64) (Thread, error)
	UpdateThread(Thread) (Thread, error)
	GetPosts(int64, int64, string, string, bool) ([]Post, error)
	CreateVote(int64, Vote) error

	// post
	GetPostById(int64) (Post, error)
	UpdatePost(int64, Post) (Post, error)

	InitDBSQL() error
}
