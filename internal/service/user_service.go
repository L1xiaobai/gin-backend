package service

import (
	"context"
	"time"

	"go-test/pkg/code"
	appErrors "go-test/pkg/errors"
	"go-test/internal/model"
	"go-test/internal/repository"
	"go-test/internal/global"
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
	// 更新数据库
	err := s.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	// 删除redis缓存，避免脏数据
	cacheKey := fmt.Sprintf("user:%s", user.Username)
	_ = global.Redis.Delete(ctx, cacheKey)

	return nil
}