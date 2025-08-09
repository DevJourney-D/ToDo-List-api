package handler

import (
	"fmt"
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

	// Add request logging middleware
	app.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

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
		"https://to-do-list-web.vercel.app", // Add potential domain
		"http://localhost:3000",
		"http://localhost:5173",
		"http://127.0.0.1:5500",
		"http://localhost:5500",
		"*", // Allow all origins as fallback
	}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Access-Control-Allow-Origin"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.ExposeHeaders = []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers"}
	app.Use(cors.New(corsConfig))

	// Add debug middleware
	app.Use(func(c *gin.Context) {
		fmt.Printf("Request: %s %s from %s\n", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		fmt.Printf("Headers: %v\n", c.Request.Header)
		c.Next()
	})

	// Add rate limiting - 200 requests per minute per IP (increased for testing)
	app.Use(middleware.RateLimitMiddleware(200, time.Minute))

	setupRoutes()
}

func setupRoutes() {
	// Simple test endpoint (no middleware)
	app.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "API is working",
			"time":    time.Now().Format(time.RFC3339),
			"origin":  c.GetHeader("Origin"),
			"ip":      c.ClientIP(),
		})
	})

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
			<a href="/api/v1/ping" class="btn">🏓 Ping</a>
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

	// Health check endpoint (same as /health but under /api/v1)
	app.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "healthy",
			"message":     "ToDo List API is running on Vercel",
			"time":        time.Now().Format(time.RFC3339),
			"api_version": "v1",
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

	// API Info endpoint to help frontend understand the correct API URL
	app.GET("/api/v1/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"api_name":         "ToDo List API",
			"version":          "1.0.0",
			"frontend_url":     "https://daily-palette.vercel.app",
			"api_base_url":     c.Request.Header.Get("Host"),
			"full_api_url":     "https://" + c.Request.Header.Get("Host"),
			"message":          "Use this API URL in your frontend configuration",
			"example_endpoint": "https://" + c.Request.Header.Get("Host") + "/api/v1/tasks",
			"cors_enabled":     true,
			"auth_required":    true,
			"auth_header":      "Authorization: Bearer <your-jwt-token>",
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

	// Handle all OPTIONS requests
	app.OPTIONS("/api/v1/*path", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Status(200)
	})

	// Public endpoints for frontend initial setup
	app.GET("/api/v1/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "pong",
			"timestamp": time.Now().Format(time.RFC3339),
			"status":    "ok",
		})
	})

	// Authentication endpoints
	app.POST("/api/v1/register", authController.Register)
	app.POST("/api/v1/login", authController.Login)
	app.GET("/api/v1/check-username", authController.CheckUsername)

	// Protected routes (authentication required)
	protected := app.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())

	// User routes
	userRoutes := []struct {
		method, path string
		handler      gin.HandlerFunc
	}{
		{"GET", "/user/info", authController.GetUserInfo},
		{"PATCH", "/user/profile", authController.UpdateProfile},
		{"PATCH", "/user/change-password", authController.ChangePassword},
		{"GET", "/user/logs", logController.GetUserLogs},
	}
	for _, r := range userRoutes {
		protected.Handle(r.method, r.path, r.handler)
	}

	// Task CRUD routes
	taskRoutes := []struct {
		method, path string
		handler      gin.HandlerFunc
	}{
		{"POST", "/tasks", taskController.CreateTask},
		{"GET", "/tasks", taskController.GetTasks},
		{"GET", "/tasks/:id", taskController.GetTask},
		{"PATCH", "/tasks/:id", taskController.UpdateTask},
		{"PUT", "/tasks/:id", taskController.UpdateTask},
		{"DELETE", "/tasks/:id", taskController.DeleteTask},
		{"PATCH", "/tasks/:id/complete", taskController.MarkTaskCompleted},
		{"GET", "/tasks/category/:category", taskController.GetTasksByCategory},
		{"GET", "/tasks/priority/:priority", taskController.GetTasksByPriority},
		{"GET", "/tasks/categories", taskController.GetCategories},
		{"GET", "/tasks/status/:status", taskController.GetTasksByStatus},
		{"GET", "/tasks/due/today", taskController.GetTasksDueToday},
		{"GET", "/tasks/due/week", taskController.GetTasksDueThisWeek},
		{"GET", "/tasks/overdue", taskController.GetOverdueTasks},
		{"GET", "/tasks/search", taskController.SearchTasks},
		{"PATCH", "/tasks/:id/status", taskController.UpdateTaskStatus},
		{"POST", "/tasks/:id/duplicate", taskController.DuplicateTask},
		{"GET", "/tasks/calendar/day/:date", taskController.GetTasksByDate},
		{"GET", "/tasks/calendar/week/:date", taskController.GetTasksForWeek},
		{"GET", "/tasks/calendar/month/:date", taskController.GetTasksForMonth},
		{"PATCH", "/tasks/:id/reschedule", taskController.RescheduleTask},
		{"GET", "/dashboard/summary", taskController.GetDashboardSummary},
		{"GET", "/dashboard/upcoming", taskController.GetUpcomingTasks},
		{"GET", "/dashboard/recent", taskController.GetRecentActivity},
		{"GET", "/tasks/export", taskController.ExportTasks},
		{"POST", "/tasks/import", taskController.ImportTasks},
	}
	for _, r := range taskRoutes {
		protected.Handle(r.method, r.path, r.handler)
	}

	// Habit CRUD routes
	habitRoutes := []struct {
		method, path string
		handler      gin.HandlerFunc
	}{
		{"POST", "/habits", habitController.CreateHabit},
		{"GET", "/habits", habitController.GetHabits},
		{"GET", "/habits/:id", habitController.GetHabit},
		{"PATCH", "/habits/:id", habitController.UpdateHabit},
		{"DELETE", "/habits/:id", habitController.DeleteHabit},
		{"PATCH", "/habits/:id/track", habitController.TrackHabit},
		{"PATCH", "/habits/:id/achieve", habitController.MarkHabitAchieved},
		{"GET", "/habits/export", habitController.ExportHabits},
		{"POST", "/habits/import", habitController.ImportHabits},
		{"GET", "/habits/:id/streak", habitController.GetHabitStreak},
		{"GET", "/habits/consistency-report", habitController.GetHabitsConsistencyReport},
	}
	for _, r := range habitRoutes {
		protected.Handle(r.method, r.path, r.handler)
	}

	// Analytics routes
	analyticsRoutes := []struct {
		method, path string
		handler      gin.HandlerFunc
	}{
		{"GET", "/analytics/overview", analyticsController.GetOverview},
		{"GET", "/analytics/tasks", analyticsController.GetTaskAnalytics},
		{"GET", "/analytics/habits", analyticsController.GetHabitAnalytics},
		{"GET", "/analytics/performance/weekly", analyticsController.GetWeeklyPerformance},
		{"GET", "/analytics/performance/monthly", analyticsController.GetMonthlyPerformance},
		{"GET", "/analytics/time-allocation", analyticsController.GetTimeAllocation},
		{"GET", "/analytics/productivity-trends", analyticsController.GetProductivityTrends},
	}
	for _, r := range analyticsRoutes {
		protected.Handle(r.method, r.path, r.handler)
	}

	// Focus Tools routes
	focusRoutes := []struct {
		method, path string
		handler      gin.HandlerFunc
	}{
		{"POST", "/focus/pomodoro/start", habitController.StartPomodoroSession},
		{"POST", "/focus/pomodoro/complete", habitController.CompletePomodoroSession},
		{"GET", "/focus/pomodoro/stats", habitController.GetPomodoroStats},
	}
	for _, r := range focusRoutes {
		protected.Handle(r.method, r.path, r.handler)
	}

	// Goals & Personal Growth routes
	goalRoutes := []struct {
		method, path string
		handler      gin.HandlerFunc
	}{
		{"POST", "/goals", habitController.CreateGoal},
		{"GET", "/goals", habitController.GetGoals},
		{"GET", "/goals/:id", habitController.GetGoal},
		{"PATCH", "/goals/:id", habitController.UpdateGoal},
		{"DELETE", "/goals/:id", habitController.DeleteGoal},
		{"PATCH", "/goals/:id/progress", habitController.UpdateGoalProgress},
	}
	for _, r := range goalRoutes {
		protected.Handle(r.method, r.path, r.handler)
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
