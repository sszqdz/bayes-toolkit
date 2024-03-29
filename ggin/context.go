package ggin

import "github.com/gin-gonic/gin"

func Get[T any](gtx *gin.Context, key string) (T, bool) {
	var (
		result T
		exist  bool
	)
	if val, ok := gtx.Get(key); ok && val != nil {
		result, exist = val.(T)
	}

	return result, exist
}
