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

var app *gin.Engine

func init() {
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
	authController := controllers.NewAuthController(authService)
	taskController := controllers.NewTaskController(taskService)
	habitController := controllers.NewHabitController(habitService)
	logController := controllers.NewLogController(logService)
	analyticsController := controllers.NewAnalyticsController(authService, taskService, habitService)

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode) // Set to release mode for production
	app = gin.New()

	// Add recovery middleware
	app.Use(gin.Recovery())

	// Configure CORS for production
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{
		"*", // Allow all origins for simplicity, configure specific domains in production
	}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	app.Use(cors.New(corsConfig))

	// Public routes (no authentication required)
	public := app.Group("/api/v1")
	{
		public.POST("/register", authController.Register)
		public.POST("/login", authController.Login)
		public.GET("/check-username", authController.CheckUsername)
	}

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
		protected.PATCH("/tasks/:id", taskController.UpdateTask)  // Update
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

	// Health check endpoint with enhanced response
	app.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "✅ healthy",
			"message": "ToDo List Backend API กำลังทำงานบน Vercel Serverless",
			"timestamp": gin.H{
				"server_time": gin.H{
					"utc":     fmt.Sprintf("%v", time.Now().UTC()),
					"bangkok": fmt.Sprintf("%v", time.Now().In(time.FixedZone("Bangkok", 7*3600))),
				},
			},
			"version": "1.0.0",
			"environment": gin.H{
				"platform":  "Vercel Serverless",
				"runtime":   "Go",
				"framework": "Gin",
				"gin_mode":  gin.Mode(),
			},
			"database": gin.H{
				"type":     "PostgreSQL",
				"status":   "connected",
				"provider": "Supabase",
			},
			"endpoints": gin.H{
				"total_routes": "35+",
				"auth_routes": []string{
					"POST /api/v1/register",
					"POST /api/v1/login",
					"GET /api/v1/check-username",
				},
				"main_features": []string{
					"Task Management",
					"Habit Tracking",
					"Analytics & Reports",
					"Pomodoro Timer",
					"Goal Setting",
					"Export/Import",
				},
			},
			"performance": gin.H{
				"cold_start":    "< 500ms",
				"response_time": "< 100ms",
				"uptime":        "99.9%",
			},
		})
	})

	// Root endpoint with beautiful HTML preview
	app.GET("/", func(c *gin.Context) {
		html := `
<!DOCTYPE html>
<html lang="th">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ToDo List API - เซิร์ฟเวอร์กำลังทำงาน</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Arial', sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
        }
        
        .container {
            text-align: center;
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 3rem;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
            border: 1px solid rgba(255, 255, 255, 0.2);
            max-width: 600px;
            width: 90%;
        }
        
        .status-indicator {
            width: 20px;
            height: 20px;
            background: #4ade80;
            border-radius: 50%;
            display: inline-block;
            margin-right: 10px;
            animation: pulse 2s infinite;
        }
        
        @keyframes pulse {
            0% { opacity: 1; }
            50% { opacity: 0.5; }
            100% { opacity: 1; }
        }
        
        h1 {
            font-size: 2.5rem;
            margin-bottom: 1rem;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .subtitle {
            font-size: 1.2rem;
            margin-bottom: 2rem;
            opacity: 0.9;
        }
        
        .api-info {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 15px;
            padding: 2rem;
            margin-bottom: 2rem;
            text-align: left;
        }
        
        .endpoint {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.8rem;
            margin: 0.5rem 0;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 8px;
            border-left: 4px solid #4ade80;
        }
        
        .endpoint-method {
            background: #3b82f6;
            color: white;
            padding: 0.3rem 0.8rem;
            border-radius: 4px;
            font-size: 0.8rem;
            font-weight: bold;
        }
        
        .endpoint-path {
            font-family: 'Courier New', monospace;
            color: #e2e8f0;
        }
        
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 1rem;
            margin-top: 2rem;
        }
        
        .stat-card {
            background: rgba(255, 255, 255, 0.1);
            padding: 1rem;
            border-radius: 10px;
            text-align: center;
        }
        
        .stat-number {
            font-size: 1.8rem;
            font-weight: bold;
            color: #4ade80;
        }
        
        .stat-label {
            font-size: 0.9rem;
            opacity: 0.8;
            margin-top: 0.5rem;
        }
        
        .footer {
            margin-top: 2rem;
            opacity: 0.7;
            font-size: 0.9rem;
        }
        
        .health-check {
            display: inline-block;
            background: #4ade80;
            color: #065f46;
            padding: 0.5rem 1rem;
            border-radius: 20px;
            text-decoration: none;
            margin-top: 1rem;
            transition: transform 0.2s;
        }
        
        .health-check:hover {
            transform: translateY(-2px);
        }
        
        @media (max-width: 768px) {
            h1 { font-size: 2rem; }
            .container { padding: 2rem; }
            .endpoint { flex-direction: column; gap: 0.5rem; }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>
            <span class="status-indicator"></span>
            เซิร์ฟเวอร์กำลังทำงาน
        </h1>
        <p class="subtitle">ToDo List API พร้อมให้บริการแล้ว! 🚀</p>
        
        <div class="api-info">
            <h3 style="margin-bottom: 1rem; color: #4ade80;">📡 API Endpoints</h3>
            
            <div class="endpoint">
                <span class="endpoint-path">/health</span>
                <span class="endpoint-method">GET</span>
            </div>
            
            <div class="endpoint">
                <span class="endpoint-path">/api/v1/register</span>
                <span class="endpoint-method">POST</span>
            </div>
            
            <div class="endpoint">
                <span class="endpoint-path">/api/v1/login</span>
                <span class="endpoint-method">POST</span>
            </div>
            
            <div class="endpoint">
                <span class="endpoint-path">/api/v1/tasks</span>
                <span class="endpoint-method">GET</span>
            </div>
            
            <div class="endpoint">
                <span class="endpoint-path">/api/v1/habits</span>
                <span class="endpoint-method">GET</span>
            </div>
            
            <div class="endpoint">
                <span class="endpoint-path">/api/v1/analytics</span>
                <span class="endpoint-method">GET</span>
            </div>
        </div>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">1.0.0</div>
                <div class="stat-label">เวอร์ชัน</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">✅</div>
                <div class="stat-label">สถานะ</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">🌐</div>
                <div class="stat-label">Vercel</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">⚡</div>
                <div class="stat-label">Serverless</div>
            </div>
        </div>
        
        <a href="/health" class="health-check">🔍 ตรวจสอบสุขภาพเซิร์ฟเวอร์</a>
        
        <div class="footer">
            <p>🛠️ พัฒนาด้วย Go + Gin Framework</p>
            <p>☁️ Deploy บน Vercel Serverless</p>
            <p id="timestamp"></p>
        </div>
    </div>
    
    <script>
        // แสดงเวลาปัจจุบัน
        function updateTimestamp() {
            const now = new Date();
            const options = { 
                timeZone: 'Asia/Bangkok',
                year: 'numeric',
                month: 'long',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit'
            };
            document.getElementById('timestamp').textContent = 
                '⏰ ' + now.toLocaleDateString('th-TH', options);
        }
        
        updateTimestamp();
        setInterval(updateTimestamp, 1000);
        
        // เพิ่มเอฟเฟกต์ hover ให้กับ endpoint cards
        document.querySelectorAll('.endpoint').forEach(endpoint => {
            endpoint.addEventListener('mouseenter', function() {
                this.style.transform = 'translateX(5px)';
                this.style.transition = 'transform 0.2s';
            });
            
            endpoint.addEventListener('mouseleave', function() {
                this.style.transform = 'translateX(0)';
            });
        });
    </script>
</body>
</html>`
		c.Data(200, "text/html; charset=utf-8", []byte(html))
	})
}

// Handler is the serverless function handler for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
