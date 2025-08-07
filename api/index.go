package handler

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
	app.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	app.Use(cors.New(corsConfig))

	setupRoutes()
}

func setupRoutes() {
	app.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "ToDo List API is running on Vercel",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	app.GET("/", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html>
<head><title>ToDo List API</title></head>
<body style="font-family: Arial; text-align: center; padding: 50px;">
<h1>🚀 ToDo List API</h1>
<p>เซิร์ฟเวอร์กำลังทำงานบน Vercel Serverless</p>
<a href="/health" style="background: #4ade80; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Check Health</a>
</body>
</html>`
		c.Data(200, "text/html; charset=utf-8", []byte(html))
	})

	// Basic API route
	app.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"api":     "ToDo List API",
			"status":  "running",
			"version": "1.0.0",
		})
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
