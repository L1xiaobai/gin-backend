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
