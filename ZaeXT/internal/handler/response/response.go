package response

import (
	"ai-qa-backend/internal/pkg/e"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Result(c *gin.Context, httpCode int, resp Response) {
	c.JSON(httpCode, resp)
}

func Success(c *gin.Context, data interface{}) {
	Result(c, http.StatusOK, Response{
		Code: e.Success,
		Msg:  e.GetMsg(e.Success),
		Data: data,
	})
}

func Fail(c *gin.Context, code int, msg string) {
	if msg == "" {
		msg = e.GetMsg(code)
	}
	httpCode := MapErrorCodeToHTTPStatus(code)
	Result(c, httpCode, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func MapErrorCodeToHTTPStatus(code int) int {
	switch code {
	case e.Success:
		return http.StatusOK
	case e.InvalidParams:
		return http.StatusBadRequest // 400
	case e.Unauthorized:
		return http.StatusUnauthorized // 401
	case e.PermissionDenied:
		return http.StatusForbidden // 403
	case e.NotFound:
		return http.StatusNotFound // 404
	case e.TooManyRequests:
		return http.StatusTooManyRequests // 429
	case e.Error:
		// 500
		fallthrough
	default:
		return http.StatusInternalServerError // 500
	}
}
