package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Transfer money feature
	r.POST("/transfer", func(c *gin.Context) {
		xTimestampStr := c.GetHeader("X-Timestamp")
		if xTimestampStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "X-Timestamp header is required"})
			return
		}

		// Process transfer logic here
		xTimestamp, err := strconv.ParseInt(xTimestampStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid X-Timestamp header"})
			return
		}
		xTimestampSeconds := xTimestamp / 1000 // Convert milliseconds to seconds
		frontendTimestamp := time.Unix(xTimestampSeconds, 0)
		log.Printf("Received X-Timestamp: %s, Converted to time: %s", xTimestampStr, frontendTimestamp)

		currentTime := time.Now()
		diff := currentTime.Sub(frontendTimestamp)
		// Check if the difference is within a threshold (e.g., 5 seconds)
		threshold := time.Second * 5

		if diff < 0 {
			diff = -diff // Ensure diff is always positive
		}

		if diff <= threshold {
			c.JSON(http.StatusOK, gin.H{"status": "transfer successful"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer rejected"})
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8000")
}
