package api

import (
	"context"
	"time"
	"strconv"

	"go-test/internal/dto"
	"go-test/internal/service"
	"go-test/internal/model"
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
		response.Error(c, err)
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

	if err := h.userService.UpdateUser(ctx, user); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
    // 获取 GET 请求的 query 参数
    pageStr := c.DefaultQuery("page", "1")          // 默认页码 1
    pageSizeStr := c.DefaultQuery("page_size", "10") // 默认每页 10 条

    // 转换为整数
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        response.Fail(c, code.InvalidParam, "page 参数无效")
        return
    }

    pageSize, err := strconv.Atoi(pageSizeStr)
    if err != nil || pageSize < 1 {
        response.Fail(c, code.InvalidParam, "page_size 参数无效")
        return
    }

    // 调用 service 获取用户列表
    users, err := h.userService.ListUsers(c.Request.Context(), page, pageSize)
    if err != nil {
        response.Error(c, err)
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
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}