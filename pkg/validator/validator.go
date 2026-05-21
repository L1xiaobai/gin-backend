package validator

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func BindJSON(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return formatValidateError(err)
	}
	return nil
}

func formatValidateError(err error) error {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldErr := range validationErrors {
			return errors.New(formatFieldError(fieldErr))
		}
	}

	return errors.New("参数格式错误")
}

func fieldName(field string) string {
	switch field {
	case "Username":
		return "用户名"
	case "Password":
		return "密码"
	default:
		return field
	}
}

func formatFieldError(err validator.FieldError) string {
	field := fieldName(err.Field())

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "min":
		return fmt.Sprintf("%s长度不能小于%s", field, err.Param())
	case "max":
		return fmt.Sprintf("%s长度不能大于%s", field, err.Param())
	case "email":
		return fmt.Sprintf("%s格式不正确", field)
	default:
		return fmt.Sprintf("%s参数不合法", field)
	}
}