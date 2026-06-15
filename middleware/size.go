package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequestSizeLimiter(limit int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.ContentLength > limit {
			ctx.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"code":        413,
				"description": "request too large",
				"detail":      "request body exceeds size limit",
			})
			return
		}
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, limit)
		ctx.Next()
	}
}
