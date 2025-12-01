package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yaikob92/URL-Shortener/api/database"
	"github.com/yaikob92/shorten-url-fiber-redis-yt/database"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")

	r := database.CreateClienet(0)
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short not found"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	rIncr := database.CreateClienet(1)
	defer rIncr.Close()

	r.Incr(database.Ctx, "counter")
	return c.Redirect(value, fiber.StatusTemporaryRedirect)
}
