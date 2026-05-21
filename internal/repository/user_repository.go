package repository

import (
	"errors"
	"context"
	
	"go-test/pkg/code"
	"go-test/internal/model"
	appErrors "go-test/pkg/errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type UserRepository struct{
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return appErrors.New(code.UserExists, "用户名已存在")
		}

		return appErrors.Wrap(code.DatabaseError, "创建用户失败", err)
	}
	return nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(code.UserNotFound, "用户不存在")
		}
		return nil, appErrors.Wrap(code.DatabaseError, "查询用户失败", err)
	}
	return &user, nil
}


func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
    return r.db.WithContext(ctx).Model(&model.User{}).
        Where("id = ?", user.ID).
        Updates(map[string]interface{}{
            "username": user.Username,
            "password": user.Password,
        }).Error
}

func (r *UserRepository) ListUsers(ctx context.Context, offset, limit int) ([]*model.User, error) {
	var users []*model.User
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}


func (r *UserRepository) DeleteUser(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}