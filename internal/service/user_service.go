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
	"go-test/internal/global"
	
	"github.com/redis/go-redis/v9"
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

	return s.userRepo.CreateUser(ctx, user)
}


// redis缓存查询
func (s *UserService) GetUserFromCache(ctx context.Context, username string) (*model.User, error) {
    key := fmt.Sprintf("user:%s", username)

    val, err := global.Redis.Get(ctx, key).Result()
    if err == redis.Nil {
        // 缓存未命中，查询数据库
        user, err := s.userRepo.FindByUsername(ctx, username)
        if err != nil {
            return nil, err
        }

        // 写回缓存，ttl 5 min
        global.Redis.Set(ctx, key, user.Username, time.Minute*5)
        return user, nil
    } else if err != nil {
        return nil, err
    }

    // 缓存命中，返回
    return &model.User{Username: val}, nil
}


func (s *UserService) UpdateUser(ctx context.Context, user *model.User) error {
    cacheKey := fmt.Sprintf("user:%s", user.Username)

    return db.WithTransactionAndCache(ctx, func(tx *gorm.DB) error {
        return s.userRepo.UpdateUser(ctx, user)
    }, cacheKey)
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.ListUsers(ctx, offset, pageSize)
}


func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	cacheKey := fmt.Sprintf("user:%d", id)

	err := db.WithTransactionAndCache(ctx, func(tx *gorm.DB) error {
		return s.userRepo.DeleteUser(ctx, id)
	}, cacheKey)

	return err
}