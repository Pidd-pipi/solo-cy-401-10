package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/constants"
)

// parseQueryUint parses a required uint query parameter.
func parseQueryUint(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Query(name), 10, 64)
	if err != nil {
		return 0, constants.NewAppError(constants.CodeBadRequest, "缺少有效参数 "+name)
	}
	return uint(v), nil
}
