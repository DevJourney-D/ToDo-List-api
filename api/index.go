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
	corsConfig.AllowOrigins = []string{
		"https://daily-palette.vercel.app",
		"http://localhost:3000",
		"http://localhost:5173",
		"http://127.0.0.1:5500",
		"http://localhost:5500",
		"*", // Allow all origins as fallback
	}
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
<head>
    <title>ToDo List API</title>
    <style>
        body { font-family: Arial; text-align: center; padding: 50px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; min-height: 100vh; margin: 0; }
        .container { background: rgba(255,255,255,0.1); padding: 2rem; border-radius: 15px; max-width: 600px; margin: 0 auto; }
        .btn { background: #4ade80; color: #065f46; padding: 10px 20px; text-decoration: none; border-radius: 5px; margin: 5px; display: inline-block; }
        .btn:hover { transform: translateY(-2px); transition: transform 0.2s; }
        .frontend-link { background: #3b82f6; color: white; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 ToDo List API</h1>
        <p>เซิร์ฟเวอร์กำลังทำงานบน Vercel Serverless</p>
        <p>🌐 Connected Frontend: <strong>Daily Palette</strong></p>
        
        <div style="margin: 20px 0;">
            <a href="/health" class="btn">🔍 Check Health</a>
            <a href="/test" class="btn">🧪 Test API</a>
            <a href="/api/v1/status" class="btn">📊 API Status</a>
        </div>
        
        <div style="margin: 20px 0;">
            <a href="https://daily-palette.vercel.app" class="btn frontend-link" target="_blank">🎨 Open Frontend App</a>
        </div>
        
        <div style="margin-top: 30px; font-size: 0.9em; opacity: 0.8;">
            <p>⚡ CORS enabled for: daily-palette.vercel.app</p>
            <p>🔗 API Base URL: <code>${window.location.origin}</code></p>
        </div>
    </div>
    
    <script>
        // Auto-update CORS info
        document.addEventListener('DOMContentLoaded', function() {
            const baseUrl = document.querySelector('code');
            if (baseUrl) {
                baseUrl.textContent = window.location.origin;
            }
        });
    </script>
</body>
</html>`
		c.Data(200, "text/html; charset=utf-8", []byte(html))
	})

	// Basic API routes
	app.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"api":     "ToDo List API",
			"status":  "running",
			"version": "1.0.0",
			"message": "API is working properly on Vercel",
		})
	})

	// Test route for debugging
	app.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":      "Test route working",
			"path":         "/test",
			"origin":       c.GetHeader("Origin"),
			"cors_enabled": "✅",
		})
	})

	// CORS test endpoint
	app.GET("/api/v1/cors-test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":      "CORS test successful",
			"frontend_url": "https://daily-palette.vercel.app",
			"origin":       c.GetHeader("Origin"),
			"user_agent":   c.GetHeader("User-Agent"),
			"method":       c.Request.Method,
			"timestamp":    time.Now().Format(time.RFC3339),
		})
	})

	// Authentication endpoints (basic implementation)
	app.POST("/api/v1/register", func(c *gin.Context) {
		var request struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}
		
		// TODO: Implement actual registration logic with database
		c.JSON(200, gin.H{
			"message": "Registration endpoint (demo)",
			"status":  "success",
			"user": gin.H{
				"username": request.Username,
				"email":    request.Email,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.POST("/api/v1/login", func(c *gin.Context) {
		var request struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}
		
		// TODO: Implement actual login logic with database and JWT
		c.JSON(200, gin.H{
			"message": "Login endpoint (demo)",
			"status":  "success",
			"user": gin.H{
				"username": request.Username,
				"id":       1,
			},
			"token": "demo-jwt-token-placeholder",
			"note":   "This is a demo response. JWT authentication needed.",
		})
	})

	app.GET("/api/v1/check-username", func(c *gin.Context) {
		username := c.Query("username")
		if username == "" {
			c.JSON(400, gin.H{
				"error": "Username parameter required",
			})
			return
		}
		
		// TODO: Check username availability in database
		c.JSON(200, gin.H{
			"username":  username,
			"available": true, // Demo response
			"message":   "Username check endpoint (demo)",
		})
	})

	// Catch all route for debugging
	app.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":        "Route not found",
			"path":         c.Request.URL.Path,
			"method":       c.Request.Method,
			"message":      "The requested route does not exist",
			"frontend_url": "https://daily-palette.vercel.app",
			"available_routes": []string{
				"GET /",
				"GET /health",
				"GET /test",
				"GET /api/v1/status",
				"GET /api/v1/cors-test",
				"POST /api/v1/register",
				"POST /api/v1/login",
				"GET /api/v1/check-username",
			},
		})
	})
}

// Handler is the main entry point for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
