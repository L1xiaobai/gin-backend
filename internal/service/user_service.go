package service

import (
	"context"
	"go-test/pkg/code"
	appErrors "go-test/pkg/errors"
	"go-test/internal/model"
	"go-test/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Register(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return appErrors.New(code.InvalidParam, "用户名和密码不能为空")
	}

	user := &model.User{
		Username: username,
		Password: password,
	}

	return s.userRepo.Create(ctx, user)
}
