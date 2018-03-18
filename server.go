package main

import (
	"net/http"

	"github.com/TrendCatcher/controllers"
	"github.com/TrendCatcher/database"
	"github.com/TrendCatcher/utils"
	"github.com/labstack/echo"
	echomiddleware "github.com/labstack/echo/middleware"
	"github.com/thoas/stats"

	"github.com/tuvistavie/structomap"
)

func init() {
	utils.ConfigRuntime()
	database.Connect()

	// Use snake case in all serializers
	structomap.SetDefaultCase(structomap.SnakeCase)
}

func main() {

	stats := stats.New()
	router := echo.New()
	router.Debug = true
	router.Pre(echomiddleware.RemoveTrailingSlash())
	router.Use(echomiddleware.Logger())
	router.Use(echomiddleware.Recover())

	router.Use(
		echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
			AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
			AllowOrigins:     []string{"*"},
			AllowHeaders:     []string{"Origin", "Accept", "Content-Type", "Authorization"},
			AllowCredentials: true,
		}))

	router.GET("all", controllers.ListExpressions)
	// Stats
	router.GET("/stats/system", func(c echo.Context) error {
		return c.JSON(http.StatusOK, stats.Data())
	})

	// Start listening
	router.Logger.Fatal(router.Start(":" + utils.GetEnvOrDefault("PORT", "3000")))
}
