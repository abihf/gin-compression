package compression

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Gin(configFns ...Configuration) gin.HandlerFunc {
	config := initConfig(configFns)

	return func(c *gin.Context) {
		c.Header("Vary", "Accept-Encoding")
		origWriter := c.Writer
		proxyWriter, ok := createResponseWriter(config, origWriter, c.Request)
		if !ok {
			return
		}
		defer proxyWriter.end()

		c.Writer = proxyWriter
		c.Next()
		// should we revert c.Writer value?
	}
}

func Wrap(h http.Handler, configFns ...Configuration) http.Handler {
	config := initConfig(configFns)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")

		proxyWriter, ok := createResponseWriter(config, w, r)
		if !ok {
			h.ServeHTTP(w, r)
			return
		}
		defer proxyWriter.end()

		h.ServeHTTP(proxyWriter, r)
	})
}
