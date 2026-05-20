package api

import (
	stderrors "errors"

	"go-test/internal/service"
	appErrors "go-test/pkg/errors"
	"go-test/pkg/code"
	"go-test/pkg/response"

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
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, code.InvalidParam, "参数错误")
		return
	}

	ctx := c.Request.Context()

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
