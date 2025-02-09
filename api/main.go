package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yaikob92/url_shorten/routes"
)

func setupRoutes(app *gin.Engine) {
	// end point function
	app.POST("/api/v1", routes.ShortenURL)
	app.GET("/:url", routes.ResolveURL)
}

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println(err)
	}

	app := gin.Default() // new gin instance with default logger

	setupRoutes(app)

	log.Fatal(app.Run(os.Getenv("APP_PORT")))
}
