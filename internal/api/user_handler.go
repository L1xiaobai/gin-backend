package api

import (
	stderrors "errors"
	"context"
	"time"

	"go-test/internal/dto"
	"go-test/internal/service"
	"go-test/internal/model"
	appErrors "go-test/pkg/errors"
	"go-test/pkg/code"
	"go-test/pkg/response"
	"go-test/pkg/validator"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	// 参数绑定和校验
	if err := validator.BindJSON(c, &req); err != nil {
		response.Fail(c, code.InvalidParam, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := h.userService.Register(ctx, req.Username, req.Password); err != nil {
		var appErr *appErrors.AppError
		if stderrors.As(err, &appErr) {
			response.Fail(c, appErr.Code, appErr.Msg)
			return
		}

		response.Fail(c, code.InternalError, "系统内部错误")
		return
	}

	response.Success(c, nil)
}


func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest

	if err := validator.BindJSON(c, &req); err != nil {
		response.Fail(c, code.InvalidParam, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	user := &model.User{
		ID:       req.ID,
		Username: req.Username,
		Password: req.Password,
	}

	// Service 层统一处理事务和 Redis 缓存删除
	if err := h.userService.UpdateUser(ctx, user); err != nil {
        response.Fail(c, code.InternalError, err.Error())
        return
	}

	response.Success(c, nil)
}


func (h *UserHandler) ListUsers(c *gin.Context) {
	var req dto.ListUsersRequest
	if err := validator.BindJSON(c, &req); err != nil {
		response.Fail(c, code.InvalidParam, err.Error())
		return
	}

	users, err := h.userService.ListUsers(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		response.Fail(c, code.InternalError, "获取用户列表失败")
		return
	}

	response.Success(c, users)
}


func (h *UserHandler) DeleteUser(c *gin.Context) {
	type Req struct {
		ID uint `json:"id" binding:"required"`
	}
	var req Req
	if err := validator.BindJSON(c, &req); err != nil {
		response.Fail(c, code.InvalidParam, err.Error())
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), req.ID); err != nil {
		response.Fail(c, code.InternalError, "删除用户失败")
		return
	}

	response.Success(c, nil)
}