package middleware

import "github.com/gin-gonic/gin"

type sonolusVersionDelWriter struct {
	gin.ResponseWriter
}

func (w *sonolusVersionDelWriter) del() {
	w.Header().Del("Sonolus-Version")
}

func (w *sonolusVersionDelWriter) WriteHeader(code int) {
	w.del()
	w.ResponseWriter.WriteHeader(code)
}

func (w *sonolusVersionDelWriter) WriteHeaderNow() {
	w.del()
	w.ResponseWriter.WriteHeaderNow()
}

func (w *sonolusVersionDelWriter) Write(data []byte) (int, error) {
	w.del()
	return w.ResponseWriter.Write(data)
}

func (w *sonolusVersionDelWriter) WriteString(s string) (int, error) {
	w.del()
	return w.ResponseWriter.WriteString(s)
}

func RemoveSonolusVersionHeader() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer = &sonolusVersionDelWriter{ResponseWriter: ctx.Writer}
		ctx.Next()
	}
}
