package middleware

import (
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = c.GetHeader("Referer")
			if origin != "" {
				parsed, err := url.Parse(origin)
				if err == nil {
					origin = parsed.Scheme + "://" + parsed.Host
				}
			}
		}

		if origin == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Origin header required"})
			return
		}

		allowedOrigins := map[string]bool{
			"http://localhost:5173":      true,
			"https://scrapjobs.com.br":   true,
		}
		if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
			allowedOrigins[frontendURL] = true
		}

		if !allowedOrigins[origin] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid origin"})
			return
		}

		c.Next()
	}
}
