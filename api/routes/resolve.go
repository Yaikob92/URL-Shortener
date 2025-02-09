package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/yaikob92/url_shorten/database"
)

func ResolveURL(c *gin.Context) {
	url := c.Param("url")

	r := database.CreateClient(0)
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "short not found",
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	rIncr := database.CreateClient(1)
	defer rIncr.Close()

	_ = rIncr.Incr(database.Ctx, "counter")
	c.Redirect(http.StatusTemporaryRedirect, value)
}
