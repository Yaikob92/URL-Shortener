package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaikob92/url_shorten/database"
	"github.com/yaikob92/url_shorten/helpers"
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

func ShortenURL(c *fiber.Ctx) error {
	body := new(Request)

	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	// ---------------- RATE LIMIT ----------------
	r2 := database.CreateClient(1)
	defer r2.Close()

	val, err := r2.Get(database.Ctx, c.IP()).Result()

	if err == redis.Nil {
		_ = r2.Set(
			database.Ctx,
			c.IP(),
			os.Getenv("API_QUOTA"),
			30*time.Minute,
		).Err()
		val = os.Getenv("API_QUOTA")
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "redis error"})
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		valInt = 0
	}

	if valInt <= 0 {
		ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":           "rate limit exceeded",
			"rate_limit_rest": ttl.Minutes(),
		})
	}

	// ---------------- VALIDATION ----------------
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid URL",
		})
	}

	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "you cannot shorten your own domain",
		})
	}

	body.URL = helpers.EnforceHTTP(body.URL)

	// ---------------- SHORT ID ----------------
	id := body.CustomShort
	if id == "" {
		id = uuid.New().String()[:6]
	}

	// ---------------- SAVE TO REDIS ----------------
	r := database.CreateClient(0)
	defer r.Close()

	existing, err := r.Get(database.Ctx, id).Result()
	if err != nil && err != redis.Nil {
		return c.Status(500).JSON(fiber.Map{"error": "redis error"})
	}

	if existing != "" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "short URL already exists",
		})
	}

	expiry := body.Expiry
	if expiry == 0 {
		expiry = 24
	}

	err = r.Set(
		database.Ctx,
		id,
		body.URL,
		time.Duration(expiry)*time.Hour,
	).Err()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to save URL",
		})
	}

	// ---------------- UPDATE RATE LIMIT ----------------
	r2.Decr(database.Ctx, c.IP())

	remaining, _ := r2.Get(database.Ctx, c.IP()).Result()
	remainingInt, _ := strconv.Atoi(remaining)

	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()

	// ---------------- RESPONSE ----------------
	return c.JSON(Response{
		URL:            body.URL,
		CustomShort:    os.Getenv("DOMAIN") + "/" + id,
		Expiry:         expiry,
		XRateRemaining: remainingInt,
		XRateLimitRest: ttl,
	})
}
