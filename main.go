package main

import (
	"fmt"
	"log"
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

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	config.InitDatabase(cfg.DatabaseURL)
	defer config.CloseDatabase()

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
	authController := controllers.NewAuthController(authService)
	taskController := controllers.NewTaskController(taskService)
	habitController := controllers.NewHabitController(habitService)
	logController := controllers.NewLogController(logService)
	analyticsController := controllers.NewAnalyticsController(authService, taskService, habitService)

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode) // Set to release mode for production
	r := gin.New()
	r.Use(gin.Recovery())

	// Add request logging for production
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
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

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{
		"https://daily-palette.vercel.app",
		"https://to-do-list-web.vercel.app", // Add potential domain
		"http://localhost:3000",
		"http://localhost:5173",
		"http://127.0.0.1:5500",
		"http://localhost:5500",
		"http://localhost:8080",
		"*", // Allow all origins as fallback
	}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Access-Control-Allow-Origin"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.ExposeHeaders = []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers"}
	r.Use(cors.New(corsConfig))

	// Add rate limiting - 200 requests per minute per IP (increased for testing)
	r.Use(middleware.RateLimitMiddleware(200, time.Minute))

	// Public routes (no authentication required)
	public := r.Group("/api/v1")
	{
		public.POST("/register", authController.Register)
		public.POST("/login", authController.Login)
		public.GET("/check-username", authController.CheckUsername)
	}

	// Protected routes (authentication required)
	protected := r.Group("/api/v1")
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

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "ToDo List Backend API is running",
		})
	})

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
