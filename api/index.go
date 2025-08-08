package handler

import (
	"net/http"
	"time"
	"todo-backend/config"
	"todo-backend/controllers"
	"todo-backend/middleware"
	"todo-backend/repositories"
	"todo-backend/services"
	"todo-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	app                 *gin.Engine
	authController      *controllers.AuthController
	taskController      *controllers.TaskController
	habitController     *controllers.HabitController
	logController       *controllers.LogController
	analyticsController *controllers.AnalyticsController
)

func init() {
	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
	app.Use(gin.Recovery())

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	config.InitDatabase(cfg.DatabaseURL)

	// Set JWT secret
	utils.SetJWTSecret(cfg.JWTSecret)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(config.DB)
	taskRepo := repositories.NewTaskRepository(config.DB)
	habitRepo := repositories.NewHabitRepository(config.DB)
	logRepo := repositories.NewLogRepository(config.DB)
	pomodoroRepo := repositories.NewPomodoroRepository(config.DB)
	goalRepo := repositories.NewGoalRepository(config.DB)

	// Initialize services
	authService := services.NewAuthService(userRepo, logRepo)
	taskService := services.NewTaskService(taskRepo, logRepo)
	habitService := services.NewHabitService(habitRepo, logRepo, pomodoroRepo, goalRepo)
	logService := services.NewLogService(logRepo)

	// Initialize controllers
	authController = controllers.NewAuthController(authService)
	taskController = controllers.NewTaskController(taskService)
	habitController = controllers.NewHabitController(habitService)
	logController = controllers.NewLogController(logService)
	analyticsController = controllers.NewAnalyticsController(authService, taskService, habitService)

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
        .status { background: #10b981; color: white; padding: 15px; border-radius: 10px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 ToDo List API</h1>
        <p>เซิร์ฟเวอร์กำลังทำงานบน Vercel Serverless</p>
        <p>🌐 Connected Frontend: <strong>Daily Palette</strong></p>
        
        <div class="status">
            <h3>✅ Real Database Integration</h3>
            <p>API ตอนนี้ใช้ข้อมูลจริงจากฐานข้อมูล PostgreSQL/Supabase</p>
            <p>🔐 JWT Authentication | 🗄️ Full CRUD Operations | 📊 Analytics</p>
        </div>
        
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
            <p>📋 Features: Tasks, Habits, Analytics, Pomodoro, Goals</p>
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

	// Debug endpoint for task creation
	app.POST("/api/v1/debug/tasks", func(c *gin.Context) {
		// Parse raw body for debugging
		var rawBody map[string]interface{}
		if err := c.ShouldBindJSON(&rawBody); err != nil {
			c.JSON(400, gin.H{
				"error":   "Invalid JSON",
				"details": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Debug endpoint - received request",
			"body":    rawBody,
			"headers": gin.H{
				"content-type":  c.GetHeader("Content-Type"),
				"authorization": c.GetHeader("Authorization"),
			},
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

	// Authentication endpoints
	app.POST("/api/v1/register", authController.Register)
	app.POST("/api/v1/login", authController.Login)
	app.GET("/api/v1/check-username", authController.CheckUsername)

	// Protected routes (authentication required)
	protected := app.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		// User routes
		protected.GET("/user/info", authController.GetUserInfo)
		protected.PATCH("/user/profile", authController.UpdateProfile)
		protected.PATCH("/user/change-password", authController.ChangePassword)
		protected.GET("/user/logs", logController.GetUserLogs)

		// Task CRUD routes
		protected.POST("/tasks", taskController.CreateTask)       // Create
		protected.GET("/tasks", taskController.GetTasks)          // Read (list)
		protected.GET("/tasks/:id", taskController.GetTask)       // Read (single)
		protected.PATCH("/tasks/:id", taskController.UpdateTask)  // Update (partial)
		protected.PUT("/tasks/:id", taskController.UpdateTask)    // Update (full)
		protected.DELETE("/tasks/:id", taskController.DeleteTask) // Delete

		// Task specific actions
		protected.PATCH("/tasks/:id/complete", taskController.MarkTaskCompleted)
		protected.GET("/tasks/category/:category", taskController.GetTasksByCategory)
		protected.GET("/tasks/priority/:priority", taskController.GetTasksByPriority)
		protected.GET("/tasks/categories", taskController.GetCategories) // Get user categories

		// Enhanced Task Management
		protected.GET("/tasks/status/:status", taskController.GetTasksByStatus)
		protected.GET("/tasks/due/today", taskController.GetTasksDueToday)
		protected.GET("/tasks/due/week", taskController.GetTasksDueThisWeek)
		protected.GET("/tasks/overdue", taskController.GetOverdueTasks)
		protected.GET("/tasks/search", taskController.SearchTasks)
		protected.PATCH("/tasks/:id/status", taskController.UpdateTaskStatus)
		protected.POST("/tasks/:id/duplicate", taskController.DuplicateTask)

		// Calendar & Views
		protected.GET("/tasks/calendar/day/:date", taskController.GetTasksByDate)
		protected.GET("/tasks/calendar/week/:date", taskController.GetTasksForWeek)
		protected.GET("/tasks/calendar/month/:date", taskController.GetTasksForMonth)
		protected.PATCH("/tasks/:id/reschedule", taskController.RescheduleTask)

		// Dashboard & Summary
		protected.GET("/dashboard/summary", taskController.GetDashboardSummary)
		protected.GET("/dashboard/upcoming", taskController.GetUpcomingTasks)
		protected.GET("/dashboard/recent", taskController.GetRecentActivity)

		// Habit CRUD routes
		protected.POST("/habits", habitController.CreateHabit)       // Create
		protected.GET("/habits", habitController.GetHabits)          // Read (list)
		protected.GET("/habits/:id", habitController.GetHabit)       // Read (single)
		protected.PATCH("/habits/:id", habitController.UpdateHabit)  // Update
		protected.DELETE("/habits/:id", habitController.DeleteHabit) // Delete

		// Habit specific actions
		protected.PATCH("/habits/:id/track", habitController.TrackHabit)
		protected.PATCH("/habits/:id/achieve", habitController.MarkHabitAchieved)

		// Export/Import routes
		protected.GET("/tasks/export", taskController.ExportTasks)
		protected.POST("/tasks/import", taskController.ImportTasks)
		protected.GET("/habits/export", habitController.ExportHabits)
		protected.POST("/habits/import", habitController.ImportHabits)

		// Analytics routes
		protected.GET("/analytics/overview", analyticsController.GetOverview)
		protected.GET("/analytics/tasks", analyticsController.GetTaskAnalytics)
		protected.GET("/analytics/habits", analyticsController.GetHabitAnalytics)

		// Performance & Insights routes
		protected.GET("/analytics/performance/weekly", analyticsController.GetWeeklyPerformance)
		protected.GET("/analytics/performance/monthly", analyticsController.GetMonthlyPerformance)
		protected.GET("/analytics/time-allocation", analyticsController.GetTimeAllocation)
		protected.GET("/analytics/productivity-trends", analyticsController.GetProductivityTrends)

		// Focus Tools routes
		protected.POST("/focus/pomodoro/start", habitController.StartPomodoroSession)
		protected.POST("/focus/pomodoro/complete", habitController.CompletePomodoroSession)
		protected.GET("/focus/pomodoro/stats", habitController.GetPomodoroStats)

		// Goals & Personal Growth routes
		protected.POST("/goals", habitController.CreateGoal)
		protected.GET("/goals", habitController.GetGoals)
		protected.GET("/goals/:id", habitController.GetGoal)
		protected.PATCH("/goals/:id", habitController.UpdateGoal)
		protected.DELETE("/goals/:id", habitController.DeleteGoal)
		protected.PATCH("/goals/:id/progress", habitController.UpdateGoalProgress)

		// Habit Streaks & Consistency
		protected.GET("/habits/:id/streak", habitController.GetHabitStreak)
		protected.GET("/habits/consistency-report", habitController.GetHabitsConsistencyReport)
	}

	// Catch all route for debugging
	app.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":        "Route not found",
			"path":         c.Request.URL.Path,
			"method":       c.Request.Method,
			"message":      "The requested route does not exist",
			"frontend_url": "https://daily-palette.vercel.app",
			"note":         "This API now uses real database and controllers",
		})
	})
}

// Handler is the main entry point for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
