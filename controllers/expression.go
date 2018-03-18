package controllers

import (
	"net/http"

	"github.com/TrendCatcher/database"
	"github.com/TrendCatcher/models"
	"github.com/labstack/echo"
)

// ListExpressions all menu titles
func ListExpressions(c echo.Context) error {
	query := database.Query{}
	query["deleted_at"] = nil

	expressions := &models.Expressions{}
	paginationParams := database.PaginationParamsForContext(c.QueryParam("page"),
		c.QueryParam("limit"), c.QueryParam("sort_by"))

	expressions, err := models.ListExpressions(query, paginationParams)
	if err != nil {
		return err
	}

	json, _ := models.NewExpressionSerializer().TransformArray(*expressions)
	return c.JSON(http.StatusOK, json)
}
