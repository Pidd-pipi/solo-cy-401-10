// Package handler implements HTTP handlers.
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/util"
)

// parseUintParam extracts a uint path parameter.
func parseUintParam(c *gin.Context, name string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		util.Fail(c, constants.NewAppError(constants.CodeBadRequest, "无效的 ID"))
		return 0, false
	}
	return uint(v), true
}
