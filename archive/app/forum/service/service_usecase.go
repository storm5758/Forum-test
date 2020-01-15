package service

import (
	. "github.com/moguchev/BD-Forum/pkg/models"
)

func (s Service) Clear() error {
	return s.Repository.Clear()
}

func (s Service) Status() (Status, error) {
	return s.Repository.Status()
}
