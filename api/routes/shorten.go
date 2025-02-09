package routes

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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

func ShortenURL(c *gin.Context) {
	var body Request

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "cannot parse JSON",
		})
		return
	}

	// ---------------- RATE LIMIT ----------------
	r2 := database.CreateClient(1)
	defer r2.Close()

	val, err := r2.Get(database.Ctx, c.ClientIP()).Result()

	if err == redis.Nil {
		_ = r2.Set(
			database.Ctx,
			c.ClientIP(),
			os.Getenv("API_QUOTA"),
			30*time.Minute,
		).Err()
		val = os.Getenv("API_QUOTA")
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "redis error"})
		return
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		valInt = 0
	}

	if valInt <= 0 {
		ttl, _ := r2.TTL(database.Ctx, c.ClientIP()).Result()
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":           "rate limit exceeded",
			"rate_limit_rest": ttl.Minutes(),
		})
		return
	}

	// ---------------- VALIDATION ----------------
	if !govalidator.IsURL(body.URL) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid URL",
		})
		return
	}

	if !helpers.RemoveDomainError(body.URL) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "you cannot shorten your own domain",
		})
		return
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "redis error"})
		return
	}

	if existing != "" {
		c.JSON(http.StatusConflict, gin.H{
			"error": "short URL already exists",
		})
		return
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save URL",
		})
		return
	}

	// ---------------- UPDATE RATE LIMIT ----------------
	r2.Decr(database.Ctx, c.ClientIP())

	remaining, _ := r2.Get(database.Ctx, c.ClientIP()).Result()
	remainingInt, _ := strconv.Atoi(remaining)

	ttl, _ := r2.TTL(database.Ctx, c.ClientIP()).Result()

	// ---------------- RESPONSE ----------------
	c.JSON(http.StatusOK, Response{
		URL:            body.URL,
		CustomShort:    os.Getenv("DOMAIN") + "/" + id,
		Expiry:         expiry,
		XRateRemaining: remainingInt,
		XRateLimitRest: ttl,
	})
}
