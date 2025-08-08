package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

// Simple auth middleware for demo
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"error":   "Authorization header required",
				"message": "Please provide a valid token",
			})
			c.Abort()
			return
		}

		// Simple token validation (in real app, verify JWT)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, gin.H{
				"error":   "Invalid authorization format",
				"message": "Use Bearer token format",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" || token == "null" || token == "undefined" {
			c.JSON(401, gin.H{
				"error":   "Invalid token",
				"message": "Token is empty or invalid",
			})
			c.Abort()
			return
		}

		// In real app, validate JWT and set user context
		c.Set("user_id", 1)
		c.Set("username", "demo_user")
		c.Next()
	}
}

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
			"note":  "This is a demo response. JWT authentication needed.",
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

	// User info endpoint (requires auth)
	app.GET("/api/v1/user/info", authMiddleware(), func(c *gin.Context) {
		userID := c.GetInt("user_id")
		username := c.GetString("username")

		c.JSON(200, gin.H{
			"status": "success",
			"user": gin.H{
				"id":       userID,
				"username": username,
				"email":    "demo@example.com",
				"profile": gin.H{
					"first_name": "Demo",
					"last_name":  "User",
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	// Tasks endpoints (require auth)
	app.GET("/api/v1/tasks", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"tasks": []gin.H{
				{
					"id":          1,
					"title":       "Demo Task 1",
					"description": "This is a demo task",
					"status":      "pending",
					"priority":    "medium",
					"category":    "work",
				},
				{
					"id":          2,
					"title":       "Demo Task 2",
					"description": "Another demo task",
					"status":      "completed",
					"priority":    "high",
					"category":    "personal",
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.POST("/api/v1/tasks", authMiddleware(), func(c *gin.Context) {
		var request struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Priority    string `json:"priority"`
			Category    string `json:"category"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"status": "success",
			"task": gin.H{
				"id":          3,
				"title":       request.Title,
				"description": request.Description,
				"priority":    request.Priority,
				"category":    request.Category,
				"status":      "pending",
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	// Dashboard summary endpoint (requires auth)
	app.GET("/api/v1/dashboard/summary", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"summary": gin.H{
				"total_tasks":        10,
				"completed_tasks":    6,
				"pending_tasks":      4,
				"overdue_tasks":      1,
				"habits_tracked":     5,
				"productivity_score": 85,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	// User profile endpoints (require auth)
	app.PATCH("/api/v1/user/profile", authMiddleware(), func(c *gin.Context) {
		var request struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Profile updated successfully (demo)",
			"user": gin.H{
				"first_name": request.FirstName,
				"last_name":  request.LastName,
				"email":      request.Email,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/user/change-password", authMiddleware(), func(c *gin.Context) {
		var request struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Password changed successfully (demo)",
			"note":    "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/user/logs", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"logs": []gin.H{
				{
					"id":        1,
					"action":    "task_created",
					"message":   "Created new task: Demo Task",
					"timestamp": time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
				},
				{
					"id":        2,
					"action":    "task_completed",
					"message":   "Completed task: Another Demo Task",
					"timestamp": time.Now().AddDate(0, 0, -2).Format(time.RFC3339),
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	// Task specific endpoints (require auth)
	app.GET("/api/v1/tasks/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status": "success",
			"task": gin.H{
				"id":          id,
				"title":       "Demo Task " + id,
				"description": "This is a demo task",
				"status":      "pending",
				"priority":    "medium",
				"category":    "work",
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/tasks/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		var request struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
			Priority    string `json:"priority"`
			Category    string `json:"category"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "success",
			"task": gin.H{
				"id":          id,
				"title":       request.Title,
				"description": request.Description,
				"status":      request.Status,
				"priority":    request.Priority,
				"category":    request.Category,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.DELETE("/api/v1/tasks/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Task deleted successfully (demo)",
			"task_id": id,
			"note":    "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/tasks/:id/complete", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status": "success",
			"task": gin.H{
				"id":     id,
				"status": "completed",
			},
			"message": "Task marked as completed (demo)",
			"note":    "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/category/:category", authMiddleware(), func(c *gin.Context) {
		category := c.Param("category")
		c.JSON(200, gin.H{
			"status":   "success",
			"category": category,
			"tasks": []gin.H{
				{
					"id":       1,
					"title":    "Demo Task in " + category,
					"category": category,
					"status":   "pending",
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/priority/:priority", authMiddleware(), func(c *gin.Context) {
		priority := c.Param("priority")
		c.JSON(200, gin.H{
			"status":   "success",
			"priority": priority,
			"tasks": []gin.H{
				{
					"id":       1,
					"title":    "Demo Task with " + priority + " priority",
					"priority": priority,
					"status":   "pending",
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/categories", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"categories": []string{
				"work",
				"personal",
				"health",
				"finance",
				"learning",
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/status/:status", authMiddleware(), func(c *gin.Context) {
		status := c.Param("status")
		c.JSON(200, gin.H{
			"status_filter": status,
			"tasks": []gin.H{
				{
					"id":     1,
					"title":  "Demo Task with " + status + " status",
					"status": status,
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/due/today", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"tasks": []gin.H{
				{
					"id":       1,
					"title":    "Task due today",
					"due_date": time.Now().Format("2006-01-02"),
					"status":   "pending",
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/due/week", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"tasks": []gin.H{
				{
					"id":       1,
					"title":    "Task due this week",
					"due_date": time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
					"status":   "pending",
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/overdue", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"tasks": []gin.H{
				{
					"id":       1,
					"title":    "Overdue task",
					"due_date": time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
					"status":   "pending",
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/search", authMiddleware(), func(c *gin.Context) {
		query := c.Query("q")
		c.JSON(200, gin.H{
			"status": "success",
			"query":  query,
			"tasks": []gin.H{
				{
					"id":    1,
					"title": "Search result for: " + query,
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/tasks/:id/status", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		var request struct {
			Status string `json:"status"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "success",
			"task": gin.H{
				"id":     id,
				"status": request.Status,
			},
			"message": "Task status updated (demo)",
			"note":    "This is a demo response. Database integration needed.",
		})
	})

	app.POST("/api/v1/tasks/:id/duplicate", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(201, gin.H{
			"status":           "success",
			"original_task_id": id,
			"new_task": gin.H{
				"id":    "new_id",
				"title": "Copy of Demo Task " + id,
			},
			"message": "Task duplicated (demo)",
			"note":    "This is a demo response. Database integration needed.",
		})
	})

	// Calendar endpoints
	app.GET("/api/v1/tasks/calendar/day/:date", authMiddleware(), func(c *gin.Context) {
		date := c.Param("date")
		c.JSON(200, gin.H{
			"status": "success",
			"date":   date,
			"tasks": []gin.H{
				{
					"id":    1,
					"title": "Task for " + date,
					"date":  date,
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/calendar/week/:date", authMiddleware(), func(c *gin.Context) {
		date := c.Param("date")
		c.JSON(200, gin.H{
			"status":     "success",
			"week_start": date,
			"tasks": []gin.H{
				{
					"id":    1,
					"title": "Task for week starting " + date,
					"date":  date,
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/tasks/calendar/month/:date", authMiddleware(), func(c *gin.Context) {
		date := c.Param("date")
		c.JSON(200, gin.H{
			"status": "success",
			"month":  date,
			"tasks": []gin.H{
				{
					"id":    1,
					"title": "Task for month " + date,
					"date":  date,
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/tasks/:id/reschedule", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		var request struct {
			NewDate string `json:"new_date"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "success",
			"task": gin.H{
				"id":       id,
				"new_date": request.NewDate,
			},
			"message": "Task rescheduled (demo)",
			"note":    "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/dashboard/upcoming", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"upcoming_tasks": []gin.H{
				{
					"id":       1,
					"title":    "Upcoming task 1",
					"due_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
				},
				{
					"id":       2,
					"title":    "Upcoming task 2",
					"due_date": time.Now().AddDate(0, 0, 2).Format("2006-01-02"),
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/dashboard/recent", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"recent_activities": []gin.H{
				{
					"id":        1,
					"action":    "task_completed",
					"task_name": "Demo Task",
					"timestamp": time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
				},
				{
					"id":        2,
					"action":    "task_created",
					"task_name": "New Demo Task",
					"timestamp": time.Now().AddDate(0, 0, -2).Format(time.RFC3339),
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	// Analytics endpoints (require auth)
	app.GET("/api/v1/analytics/overview", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"analytics": gin.H{
				"tasks_completed_today":     3,
				"tasks_completed_this_week": 15,
				"habits_tracked_today":      2,
				"productivity_trend":        "improving",
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/analytics/tasks", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"task_analytics": gin.H{
				"total_tasks":             50,
				"completed_tasks":         35,
				"pending_tasks":           15,
				"completion_rate":         70,
				"average_completion_time": "2.5 days",
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/analytics/habits", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"habit_analytics": gin.H{
				"total_habits":    10,
				"active_habits":   8,
				"streak_average":  15,
				"completion_rate": 80,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/analytics/performance/weekly", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"weekly_performance": gin.H{
				"tasks_completed":    25,
				"habits_tracked":     35,
				"productivity_score": 85,
				"week_start":         time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/analytics/performance/monthly", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"monthly_performance": gin.H{
				"tasks_completed":    100,
				"habits_tracked":     150,
				"productivity_score": 82,
				"month_start":        time.Now().AddDate(0, -1, 0).Format("2006-01-02"),
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/analytics/time-allocation", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"time_allocation": gin.H{
				"work":     40,
				"personal": 30,
				"health":   20,
				"learning": 10,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/analytics/productivity-trends", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"productivity_trends": []gin.H{
				{"date": "2025-08-01", "score": 75},
				{"date": "2025-08-02", "score": 80},
				{"date": "2025-08-03", "score": 85},
				{"date": "2025-08-04", "score": 82},
				{"date": "2025-08-05", "score": 88},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	// Habit CRUD endpoints (require auth)
	app.POST("/api/v1/habits", authMiddleware(), func(c *gin.Context) {
		var request struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Frequency   string `json:"frequency"`
			Category    string `json:"category"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"status": "success",
			"habit": gin.H{
				"id":          1,
				"name":        request.Name,
				"description": request.Description,
				"frequency":   request.Frequency,
				"category":    request.Category,
				"streak":      0,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/habits", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"habits": []gin.H{
				{
					"id":          1,
					"name":        "Morning Exercise",
					"description": "30 minutes of exercise",
					"frequency":   "daily",
					"category":    "health",
					"streak":      15,
				},
				{
					"id":          2,
					"name":        "Read Books",
					"description": "Read for 30 minutes",
					"frequency":   "daily",
					"category":    "learning",
					"streak":      8,
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/habits/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status": "success",
			"habit": gin.H{
				"id":          id,
				"name":        "Demo Habit " + id,
				"description": "This is a demo habit",
				"frequency":   "daily",
				"category":    "health",
				"streak":      10,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/habits/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		var request struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Frequency   string `json:"frequency"`
			Category    string `json:"category"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "success",
			"habit": gin.H{
				"id":          id,
				"name":        request.Name,
				"description": request.Description,
				"frequency":   request.Frequency,
				"category":    request.Category,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.DELETE("/api/v1/habits/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status":   "success",
			"message":  "Habit deleted successfully (demo)",
			"habit_id": id,
			"note":     "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/habits/:id/track", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status": "success",
			"habit": gin.H{
				"id":     id,
				"streak": 16,
			},
			"message": "Habit tracked successfully (demo)",
			"note":    "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/habits/:id/achieve", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status": "success",
			"habit": gin.H{
				"id":       id,
				"achieved": true,
			},
			"message": "Habit marked as achieved (demo)",
			"note":    "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/habits/:id/streak", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status":   "success",
			"habit_id": id,
			"streak": gin.H{
				"current":    15,
				"longest":    25,
				"this_week":  7,
				"this_month": 28,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/habits/consistency-report", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"consistency_report": gin.H{
				"overall_consistency": 85,
				"habits": []gin.H{
					{
						"id":             1,
						"name":           "Morning Exercise",
						"consistency":    90,
						"current_streak": 15,
					},
					{
						"id":             2,
						"name":           "Read Books",
						"consistency":    80,
						"current_streak": 8,
					},
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	// Export/Import endpoints
	app.GET("/api/v1/tasks/export", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":     "success",
			"export_url": "https://example.com/export/tasks.json",
			"format":     "JSON",
			"note":       "This is a demo response. Export functionality needed.",
		})
	})

	app.POST("/api/v1/tasks/import", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":         "success",
			"imported_count": 5,
			"message":        "Tasks imported successfully (demo)",
			"note":           "This is a demo response. Import functionality needed.",
		})
	})

	app.GET("/api/v1/habits/export", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":     "success",
			"export_url": "https://example.com/export/habits.json",
			"format":     "JSON",
			"note":       "This is a demo response. Export functionality needed.",
		})
	})

	app.POST("/api/v1/habits/import", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":         "success",
			"imported_count": 3,
			"message":        "Habits imported successfully (demo)",
			"note":           "This is a demo response. Import functionality needed.",
		})
	})

	// Focus Tools (Pomodoro) endpoints
	app.POST("/api/v1/focus/pomodoro/start", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"session": gin.H{
				"id":         1,
				"duration":   25,
				"start_time": time.Now().Format(time.RFC3339),
				"type":       "work",
			},
			"message": "Pomodoro session started (demo)",
			"note":    "This is a demo response. Pomodoro functionality needed.",
		})
	})

	app.POST("/api/v1/focus/pomodoro/complete", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"session": gin.H{
				"id":              1,
				"completed":       true,
				"end_time":        time.Now().Format(time.RFC3339),
				"actual_duration": 25,
			},
			"message": "Pomodoro session completed (demo)",
			"note":    "This is a demo response. Pomodoro functionality needed.",
		})
	})

	app.GET("/api/v1/focus/pomodoro/stats", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"pomodoro_stats": gin.H{
				"today_sessions":   4,
				"week_sessions":    28,
				"month_sessions":   120,
				"total_focus_time": "100 hours",
				"average_session":  "24 minutes",
			},
			"note": "This is a demo response. Pomodoro functionality needed.",
		})
	})

	// Goals endpoints
	app.POST("/api/v1/goals", authMiddleware(), func(c *gin.Context) {
		var request struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			TargetDate  string `json:"target_date"`
			Category    string `json:"category"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"status": "success",
			"goal": gin.H{
				"id":          1,
				"title":       request.Title,
				"description": request.Description,
				"target_date": request.TargetDate,
				"category":    request.Category,
				"progress":    0,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/goals", authMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
			"goals": []gin.H{
				{
					"id":          1,
					"title":       "Learn Go Programming",
					"description": "Complete Go course and build projects",
					"target_date": "2025-12-31",
					"category":    "learning",
					"progress":    60,
				},
				{
					"id":          2,
					"title":       "Run Marathon",
					"description": "Train and complete a full marathon",
					"target_date": "2025-10-15",
					"category":    "health",
					"progress":    30,
				},
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.GET("/api/v1/goals/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status": "success",
			"goal": gin.H{
				"id":          id,
				"title":       "Demo Goal " + id,
				"description": "This is a demo goal",
				"target_date": "2025-12-31",
				"category":    "personal",
				"progress":    45,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/goals/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		var request struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			TargetDate  string `json:"target_date"`
			Category    string `json:"category"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "success",
			"goal": gin.H{
				"id":          id,
				"title":       request.Title,
				"description": request.Description,
				"target_date": request.TargetDate,
				"category":    request.Category,
			},
			"note": "This is a demo response. Database integration needed.",
		})
	})

	app.DELETE("/api/v1/goals/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Goal deleted successfully (demo)",
			"goal_id": id,
			"note":    "This is a demo response. Database integration needed.",
		})
	})

	app.PATCH("/api/v1/goals/:id/progress", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		var request struct {
			Progress int `json:"progress"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "success",
			"goal": gin.H{
				"id":       id,
				"progress": request.Progress,
			},
			"message": "Goal progress updated (demo)",
			"note":    "This is a demo response. Database integration needed.",
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
				"GET /api/v1/user/info",
				"PATCH /api/v1/user/profile",
				"PATCH /api/v1/user/change-password",
				"GET /api/v1/user/logs",
				"GET /api/v1/tasks",
				"POST /api/v1/tasks",
				"GET /api/v1/tasks/:id",
				"PATCH /api/v1/tasks/:id",
				"DELETE /api/v1/tasks/:id",
				"PATCH /api/v1/tasks/:id/complete",
				"GET /api/v1/tasks/category/:category",
				"GET /api/v1/tasks/priority/:priority",
				"GET /api/v1/tasks/categories",
				"GET /api/v1/tasks/status/:status",
				"GET /api/v1/tasks/due/today",
				"GET /api/v1/tasks/due/week",
				"GET /api/v1/tasks/overdue",
				"GET /api/v1/tasks/search",
				"PATCH /api/v1/tasks/:id/status",
				"POST /api/v1/tasks/:id/duplicate",
				"GET /api/v1/tasks/calendar/day/:date",
				"GET /api/v1/tasks/calendar/week/:date",
				"GET /api/v1/tasks/calendar/month/:date",
				"PATCH /api/v1/tasks/:id/reschedule",
				"GET /api/v1/dashboard/summary",
				"GET /api/v1/dashboard/upcoming",
				"GET /api/v1/dashboard/recent",
				"POST /api/v1/habits",
				"GET /api/v1/habits",
				"GET /api/v1/habits/:id",
				"PATCH /api/v1/habits/:id",
				"DELETE /api/v1/habits/:id",
				"PATCH /api/v1/habits/:id/track",
				"PATCH /api/v1/habits/:id/achieve",
				"GET /api/v1/habits/:id/streak",
				"GET /api/v1/habits/consistency-report",
				"GET /api/v1/tasks/export",
				"POST /api/v1/tasks/import",
				"GET /api/v1/habits/export",
				"POST /api/v1/habits/import",
				"GET /api/v1/analytics/overview",
				"GET /api/v1/analytics/tasks",
				"GET /api/v1/analytics/habits",
				"GET /api/v1/analytics/performance/weekly",
				"GET /api/v1/analytics/performance/monthly",
				"GET /api/v1/analytics/time-allocation",
				"GET /api/v1/analytics/productivity-trends",
				"POST /api/v1/focus/pomodoro/start",
				"POST /api/v1/focus/pomodoro/complete",
				"GET /api/v1/focus/pomodoro/stats",
				"POST /api/v1/goals",
				"GET /api/v1/goals",
				"GET /api/v1/goals/:id",
				"PATCH /api/v1/goals/:id",
				"DELETE /api/v1/goals/:id",
				"PATCH /api/v1/goals/:id/progress",
			},
		})
	})
}

// Handler is the main entry point for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
