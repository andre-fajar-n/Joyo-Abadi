package main

import (
	"net/http"

	"github.com/andre-fajar-n/Joyo-Abadi/backend/config"
	"github.com/gin-gonic/gin"
)

func main() {
	db := config.ConnectDB()
	config.MigrateDB(db)

	r := setupRouter()

	r.Run(":8000")
}

func setupRouter() *gin.Engine {
	gin.ForceConsoleColor()

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to Joyo")
	})

	return r
}
