package routes

import (
	"time"

	"github.com/akhil/shorten-url-fiber-redis-yt/helpers"
	"github.com/asaskevich/govalidator"

	"github.com/gofiber/fiber/v2"
)

type Request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type Response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining int           `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_rest"`
}


func ShortenURL(c *fiber.Ctx){
	body := new(Request) // pointer to the Request struct
	if err := c.BodyParser(body); err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"can not parse JSON"})
	}
	
	// implement rate limiting

  // check if the input is an actual URL
	if !govalidator.IsURL(body.URL){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Invalid URL"})
	}

	// check for domain error
	if !helpers.RemoveDomainError(body.URL){
		c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error":""})
	}

	// enforce https, SSL
	body.URL = helpers.EnforceHTTP(body.URL)
}