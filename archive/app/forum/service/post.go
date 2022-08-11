package service

func (s Service) GetPostAccount(id int64, related []string) (PostAccount, error) {
	fullPost := PostAccount{}

	post, err := s.Repository.GetPostById(id)
	if err != nil {
		return fullPost, err
	}
	fullPost.Post = post

	for _, key := range related {
		switch key {
		case "user":
			u, err := s.Repository.GetUserByNickname(fullPost.Post.Author)
			if err != nil {
				return fullPost, err
			}
			fullPost.Author = &User{About: u.About, Email: u.Email, Fullname: u.Fullname, Nickname: u.Nickname}
			break
		case "forum":
			f, err := s.Repository.GetForum(fullPost.Post.Forum)
			if err != nil {
				return fullPost, err
			}
			fullPost.Forum = &Forum{Posts: f.Posts, Slug: f.Slug, Threads: f.Threads, Title: f.Title, User: f.User}
			break
		case "thread":
			th, err := s.Repository.GetThreadById(int64(fullPost.Post.Thread))
			if err != nil {
				return fullPost, err
			}
			fullPost.Thread = &Thread{Author: th.Author, Created: th.Created, Forum: th.Forum, Id: th.Id, Message: th.Message, Slug: th.Slug, Title: th.Title, Votes: th.Votes}
			break
		}
	}

	return fullPost, nil
}

func (s Service) UpdatePost(id int64, p Post) (Post, error) {
	return s.Repository.UpdatePost(id, p)
}
