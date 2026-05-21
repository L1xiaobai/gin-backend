package response

import "github.com/gin-gonic/gin"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data any         `json:"data,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(200, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(400, Response{
		Code: code,
		Msg:  msg,
	})
}

func FailData(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  msg,
	}
}

func Error(c *gin.Context, err error) {
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(400, Response{
			Code: appErr.Code,
			Msg:  appErr.Msg,
		})
		return
	}

	c.JSON(500, Response{
		Code: code.InternalError,
		Msg:  "系统内部错误",
	})
}