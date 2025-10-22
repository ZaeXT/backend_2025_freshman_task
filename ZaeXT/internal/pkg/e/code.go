package e

const (
	Success          = 200
	Error            = 500
	InvalidParams    = 400
	Unauthorized     = 401
	PermissionDenied = 403
	NotFound         = 404
	RequestTimeout   = 408
	TooManyRequests  = 429
)

var msgFlags = map[int]string{
	Success:          "ok",
	Error:            "fail",
	InvalidParams:    "请求参数错误",
	Unauthorized:     "未授权，请先登录",
	PermissionDenied: "权限不足",
	NotFound:         "请求资源不存在",
	RequestTimeout:   "请求超时",
	TooManyRequests:  "请求过于频繁",
}

func GetMsg(code int) string {
	msg, ok := msgFlags[code]
	if ok {
		return msg
	}
	return msgFlags[Error]
}
