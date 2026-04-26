package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Frontend UI usually makes requests from another origin or port (e.g., localhost:3000)
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.Any("/bookings/*path", func(c *gin.Context) {
			proxyReq(c, "http://booking:8080", "/bookings")
		})
		api.Any("/bookings", func(c *gin.Context) {
			proxyReq(c, "http://booking:8080", "/bookings")
		})
		api.Any("/passengers/*path", func(c *gin.Context) {
			proxyReq(c, "http://passenger:8080", "/passengers")
		})
		api.Any("/passengers", func(c *gin.Context) {
			proxyReq(c, "http://passenger:8080", "/passengers")
		})
		api.Any("/baggage/*path", func(c *gin.Context) {
			proxyReq(c, "http://baggage:8080", "/baggage")
		})
		api.Any("/baggage", func(c *gin.Context) {
			proxyReq(c, "http://baggage:8080", "/baggage")
		})
		api.Any("/incidents/*path", func(c *gin.Context) {
			proxyReq(c, "http://incident:8080", "/incidents")
		})
		api.Any("/incidents", func(c *gin.Context) {
			proxyReq(c, "http://incident:8080", "/incidents")
		})
		api.Any("/analytics/*path", func(c *gin.Context) {
			proxyReq(c, "http://analytics:8080", "/analytics")
		})
		api.Any("/analytics", func(c *gin.Context) {
			proxyReq(c, "http://analytics:8080", "/analytics")
		})
	}

	log.Println("Gateway listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start gateway: %v", err)
	}
}

func proxyReq(c *gin.Context, targetHost string, targetBase string) {
	remote, err := url.Parse(targetHost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid target host"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	// Update the path prefix if necessary. /api/v1/bookings -> /bookings
	c.Request.URL.Path = strings.Replace(c.Request.URL.Path, "/api/v1", "", 1)

	proxy.ServeHTTP(c.Writer, c.Request)
}
