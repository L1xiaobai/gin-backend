package api

import (
	stderrors "errors"
	"context"
	"time"

	"go-test/internal/dto"
	"go-test/internal/service"
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
	var req dto.RegisterRequest

	if err := validator.BindJSON(c, &req); err != nil {
		response.Fail(c, code.InvalidParam, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	user := &model.User{
		ID:	   	  req.ID,
		Username: req.Username,
		Password: req.Password,
	}

	if err := h.userService.UpdateUser(ctx, user); err != nil {
		var appErr *appErrors.AppError
		if stderrors.As(err, &appErr) {
			response.Fail(c, appErr.Code, appErr.Msg)
			return
		}

		response.Fail(c, code.InternalError, "系统内部错误")
		return
	}
	cacheKey := fmt.Sprintf("user:%s", user.Username)
	_ = global.Redis.Delete(ctx, cacheKey)
	
	response.Success(c, nil)
}