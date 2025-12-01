package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/yaikob92/url_shorten/routes"
)

func setupRoutes(app *fiber.App) {
	// end point function
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v4", routes.ShortenURL)
}

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println(err)
	}

	app := fiber.New() // new fiber instances

	app.Use(logger.New()) // log all incoming requests (middleware)

	setupRoutes(app)

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
