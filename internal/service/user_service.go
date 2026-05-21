package service

import (
	"context"
	"time"
	"fmt"

	"go-test/pkg/code"
	appErrors "go-test/pkg/errors"
	"go-test/pkg/db"
	"go-test/internal/model"
	"go-test/internal/repository"
	"go-test/pkg/redis"
	
	"gorm.io/gorm"
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
    cacheKey := fmt.Sprintf("user:%d", user.ID)

	return db.WithTransactionAndCache(ctx, func(tx *gorm.DB) error {
		return s.userRepo.CreateUser(ctx, tx, user)
	}, cacheKey)
}


// redis缓存查询
func (s *UserService) GetUserFromCache(ctx context.Context, userID uint) (*model.User, error) {
    key := fmt.Sprintf("user:%d", userID)
    var user model.User
    if err := redis.GetStruct(ctx, key, &user); err != nil {
        return nil, err
    }

    if user.ID == 0 {
        dbUser, err := s.userRepo.FindByID(ctx, userID)
        if err != nil {
            return nil, err
        }
        _ = redis.SetStruct(ctx, key, dbUser, 5*time.Minute)
        return dbUser, nil
    }

    return &user, nil
}


func (s *UserService) UpdateUser(ctx context.Context, user *model.User) error {
    cacheKey := fmt.Sprintf("user:%d", user.ID)

    return db.WithTransactionAndCache(ctx, func(tx *gorm.DB) error {
        return s.userRepo.UpdateUser(ctx, tx, user)
    }, cacheKey)
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.ListUsers(ctx, offset, pageSize)
}


func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	cacheKey := fmt.Sprintf("user:%d", id)

	err := db.WithTransactionAndCache(ctx, func(tx *gorm.DB) error {
		return s.userRepo.DeleteUser(ctx, tx, id)
	}, cacheKey)
	
	return err
}