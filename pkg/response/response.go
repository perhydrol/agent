package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(c *gin.Context, data any) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

func Error(c *gin.Context, err error) {
	var appErr *errno.AppErro
	if errors.As(err, &appErr) {
		c.JSON(http.StatusOK, Response{
			Code: appErr.Code,
			Msg:  appErr.Msg,
			Data: nil,
		})
		return
	}

	// 非 AppError 视为服务器内部错误
	c.JSON(http.StatusInternalServerError, Response{
		Code: errno.ErrInternalServer.Code,
		Msg:  errno.ErrInternalServer.Msg,
		Data: nil,
	})
}

func Page(c *gin.Context, data any) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}
