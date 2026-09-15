// Package docs serves a hand-maintained OpenAPI document and a simple UI.
package docs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed index.html
var indexHTML []byte

//go:embed openapi.json
var openAPIJSON []byte

// IndexHandler serves the HTML API documentation page.
func IndexHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}
}

// OpenAPIHandler serves the OpenAPI JSON document.
func OpenAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", openAPIJSON)
	}
}
