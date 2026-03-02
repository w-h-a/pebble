package service

import "github.com/w-h-a/pebble/internal/client/repo"

type Service struct {
	repo   repo.Repo
	prefix string
}

func NewService(repo repo.Repo, prefix string) *Service {
	return &Service{
		repo:   repo,
		prefix: prefix,
	}
}
