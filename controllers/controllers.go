package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo-backend/middleware"
	"todo-backend/models"
	"todo-backend/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := ctrl.authService.Register(req.Username, req.Password)
	if err != nil {
		if err.Error() == "username already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := ctrl.authService.Login(req.Username, req.Password)
	if err != nil {
		if err.Error() == "invalid username or password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (ctrl *AuthController) GetUserInfo(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userInfo, err := ctrl.authService.GetUserInfo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

func (ctrl *AuthController) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.authService.UpdateProfile(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password requirements
	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 6 characters long"})
		return
	}

	err := ctrl.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if err.Error() == "invalid current password" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

func (ctrl *AuthController) CheckUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username parameter is required"})
		return
	}

	// Basic validation
	if len(username) < 3 || len(username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be between 3 and 50 characters"})
		return
	}

	exists, err := ctrl.authService.CheckUsernameExists(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username availability"})
		return
	}

	c.JSON(http.StatusOK, models.CheckUsernameResponse{
		Exists: exists,
	})
}

// Task Controller
type TaskController struct {
	taskService services.TaskService
}

func NewTaskController(taskService services.TaskService) *TaskController {
	return &TaskController{
		taskService: taskService,
	}
}

func (ctrl *TaskController) CreateTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := ctrl.taskService.CreateTask(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (ctrl *TaskController) GetTasks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check for filters
	category := c.Query("category")
	priorityStr := c.Query("priority")

	var tasks []*models.Task
	var err error

	if category != "" {
		tasks, err = ctrl.taskService.GetTasksByCategory(userID, category)
	} else if priorityStr != "" {
		priority, parseErr := strconv.ParseInt(priorityStr, 10, 16)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority value"})
			return
		}
		tasks, err = ctrl.taskService.GetTasksByPriority(userID, int16(priority))
	} else {
		tasks, err = ctrl.taskService.GetUserTasks(userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	response := models.TaskResponse{
		Tasks: tasks,
		Total: len(tasks),
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *TaskController) GetTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := ctrl.taskService.GetTaskByID(taskID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (ctrl *TaskController) UpdateTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := ctrl.taskService.UpdateTask(taskID, userID, &req)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (ctrl *TaskController) DeleteTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	err = ctrl.taskService.DeleteTask(taskID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func (ctrl *TaskController) MarkTaskCompleted(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req struct {
		IsCompleted bool `json:"is_completed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ctrl.taskService.MarkTaskCompleted(taskID, userID, req.IsCompleted)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task status"})
		return
	}

	status := "incomplete"
	if req.IsCompleted {
		status = "completed"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Task marked as " + status,
		"is_completed": req.IsCompleted,
	})
}

func (ctrl *TaskController) ExportTasks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	format := c.Query("format")
	if format == "" {
		format = "json" // default format
	}

	if format != "json" && format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supported formats: json, csv"})
		return
	}

	data, err := ctrl.taskService.ExportTasks(userID, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export tasks"})
		return
	}

	contentType := "application/json"
	filename := "tasks.json"
	if format == "csv" {
		contentType = "text/csv"
		filename = "tasks.csv"
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

func (ctrl *TaskController) ImportTasks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ctrl.taskService.ImportTasks(userID, []byte(req.Data), req.Format)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tasks imported successfully"})
}

// Enhanced Task Management Methods
func (ctrl *TaskController) GetTasksByStatus(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	status := c.Param("status")

	// For now, map to existing completed/pending logic
	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	var filteredTasks []*models.Task
	for _, task := range tasks {
		switch status {
		case "completed":
			if task.IsCompleted {
				filteredTasks = append(filteredTasks, task)
			}
		case "pending":
			if !task.IsCompleted {
				filteredTasks = append(filteredTasks, task)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":  filteredTasks,
		"count":  len(filteredTasks),
		"status": status,
	})
}

func (ctrl *TaskController) GetTasksDueToday(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	today := time.Now().Format("2006-01-02")
	var dueTodayTasks []*models.Task

	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.Format("2006-01-02") == today {
			dueTodayTasks = append(dueTodayTasks, task)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": dueTodayTasks,
		"count": len(dueTodayTasks),
		"date":  today,
	})
}

func (ctrl *TaskController) GetTasksDueThisWeek(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	var weekTasks []*models.Task
	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.After(weekStart) && task.DueDate.Before(weekEnd) {
			weekTasks = append(weekTasks, task)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":      weekTasks,
		"count":      len(weekTasks),
		"week_start": weekStart.Format("2006-01-02"),
		"week_end":   weekEnd.Format("2006-01-02"),
	})
}

func (ctrl *TaskController) GetOverdueTasks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	now := time.Now()
	var overdueTasks []*models.Task

	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.Before(now) && !task.IsCompleted {
			overdueTasks = append(overdueTasks, task)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":        overdueTasks,
		"count":        len(overdueTasks),
		"current_time": now.Format("2006-01-02 15:04:05"),
	})
}

func (ctrl *TaskController) SearchTasks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	query := c.Query("q")
	category := c.Query("category")
	priority := c.Query("priority")
	status := c.Query("status")

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search tasks"})
		return
	}

	var filteredTasks []*models.Task
	for _, task := range tasks {
		// Text search in task name and description
		if query != "" {
			if !strings.Contains(strings.ToLower(task.TaskName), strings.ToLower(query)) {
				if task.Description == nil || !strings.Contains(strings.ToLower(*task.Description), strings.ToLower(query)) {
					continue
				}
			}
		}

		// Category filter
		if category != "" && (task.Category == nil || *task.Category != category) {
			continue
		}

		// Priority filter
		if priority != "" {
			if p, err := strconv.ParseInt(priority, 10, 16); err == nil {
				if task.Priority != int16(p) {
					continue
				}
			}
		}

		// Status filter
		if status != "" {
			switch status {
			case "completed":
				if !task.IsCompleted {
					continue
				}
			case "pending":
				if task.IsCompleted {
					continue
				}
			}
		}

		filteredTasks = append(filteredTasks, task)
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": filteredTasks,
		"count": len(filteredTasks),
		"query": query,
		"filters": gin.H{
			"category": category,
			"priority": priority,
			"status":   status,
		},
	})
}

func (ctrl *TaskController) UpdateTaskStatus(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req models.UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For now, map to existing completion logic
	isCompleted := req.Status == models.TaskStatusCompleted
	err = ctrl.taskService.MarkTaskCompleted(taskID, userID, isCompleted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task status updated successfully",
		"task_id": taskID,
		"status":  req.Status,
	})
}

func (ctrl *TaskController) DuplicateTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get original task
	originalTask, err := ctrl.taskService.GetTaskByID(taskID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Create duplicate
	duplicateReq := &models.CreateTaskRequest{
		TaskName:           originalTask.TaskName + " (Copy)",
		Description:        originalTask.Description,
		Category:           originalTask.Category,
		Priority:           originalTask.Priority,
		DueDate:            originalTask.DueDate,
		IsRecurring:        originalTask.IsRecurring,
		RecurringFrequency: originalTask.RecurringFrequency,
	}

	newTask, err := ctrl.taskService.CreateTask(userID, duplicateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to duplicate task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":          "Task duplicated successfully",
		"original_task_id": taskID,
		"new_task":         newTask,
	})
}

// Calendar & Views Methods
func (ctrl *TaskController) GetTasksByDate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	dateStr := c.Param("date")
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	var dayTasks []*models.Task
	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.Format("2006-01-02") == dateStr {
			dayTasks = append(dayTasks, task)
		}
	}

	calendarView := &models.CalendarView{
		Date:  dateStr,
		Tasks: dayTasks,
		Count: len(dayTasks),
	}

	c.JSON(http.StatusOK, calendarView)
}

func (ctrl *TaskController) GetTasksForWeek(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// Get start of week (Monday)
	weekStart := date.AddDate(0, 0, -int(date.Weekday()-time.Monday))
	weekEnd := weekStart.AddDate(0, 0, 7)

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	weekData := make(map[string]*models.CalendarView)

	// Initialize each day of the week
	for i := 0; i < 7; i++ {
		currentDate := weekStart.AddDate(0, 0, i)
		dateKey := currentDate.Format("2006-01-02")
		weekData[dateKey] = &models.CalendarView{
			Date:  dateKey,
			Tasks: []*models.Task{},
			Count: 0,
		}
	}

	// Assign tasks to appropriate days
	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.After(weekStart) && task.DueDate.Before(weekEnd) {
			dateKey := task.DueDate.Format("2006-01-02")
			if dayData, exists := weekData[dateKey]; exists {
				dayData.Tasks = append(dayData.Tasks, task)
				dayData.Count++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"week_start": weekStart.Format("2006-01-02"),
		"week_end":   weekEnd.Format("2006-01-02"),
		"days":       weekData,
	})
}

func (ctrl *TaskController) GetTasksForMonth(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// Get first and last day of month
	year, month, _ := date.Date()
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, date.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	monthData := make(map[string]*models.CalendarView)

	// Group tasks by date
	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.After(monthStart) && task.DueDate.Before(monthEnd) {
			dateKey := task.DueDate.Format("2006-01-02")
			if _, exists := monthData[dateKey]; !exists {
				monthData[dateKey] = &models.CalendarView{
					Date:  dateKey,
					Tasks: []*models.Task{},
					Count: 0,
				}
			}
			monthData[dateKey].Tasks = append(monthData[dateKey].Tasks, task)
			monthData[dateKey].Count++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"month":       month.String(),
		"year":        year,
		"month_start": monthStart.Format("2006-01-02"),
		"month_end":   monthEnd.Format("2006-01-02"),
		"days":        monthData,
	})
}

func (ctrl *TaskController) RescheduleTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req models.RescheduleTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update task with new due date
	updateReq := &models.UpdateTaskRequest{
		DueDate: &req.NewDueDate,
	}

	updatedTask, err := ctrl.taskService.UpdateTask(taskID, userID, updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reschedule task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task rescheduled successfully",
		"task":    updatedTask,
		"reason":  req.Reason,
	})
}

// Dashboard Methods
func (ctrl *TaskController) GetDashboardSummary(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard summary"})
		return
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	var summary models.DashboardSummary
	summary.TotalTasks = int64(len(tasks))

	for _, task := range tasks {
		if task.IsCompleted {
			if task.CreatedAt.Format("2006-01-02") == today {
				summary.CompletedToday++
			}
		} else {
			if task.DueDate != nil {
				if task.DueDate.Format("2006-01-02") == today {
					summary.DueToday++
				}
				if task.DueDate.Before(now) {
					summary.Overdue++
				}
			}
		}
	}

	if summary.TotalTasks > 0 {
		completedTasks := summary.TotalTasks - (summary.DueToday + summary.Overdue)
		summary.CompletionRate = float64(completedTasks) / float64(summary.TotalTasks) * 100
	}

	summary.ProductivityTrend = "stable" // This would be calculated based on historical data

	c.JSON(http.StatusOK, summary)
}

func (ctrl *TaskController) GetUpcomingTasks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limit := 10 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get upcoming tasks"})
		return
	}

	now := time.Now()
	var upcomingTasks []*models.UpcomingTask

	for _, task := range tasks {
		if !task.IsCompleted && task.DueDate != nil && task.DueDate.After(now) {
			daysLeft := int(task.DueDate.Sub(now).Hours() / 24)
			upcomingTask := &models.UpcomingTask{
				ID:       task.ID,
				TaskName: task.TaskName,
				Category: task.Category,
				Priority: task.Priority,
				DueDate:  task.DueDate,
				DaysLeft: daysLeft,
				IsUrgent: daysLeft <= 3, // Tasks due within 3 days are urgent
			}
			upcomingTasks = append(upcomingTasks, upcomingTask)
		}
	}

	// Limit results
	if len(upcomingTasks) > limit {
		upcomingTasks = upcomingTasks[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"upcoming_tasks": upcomingTasks,
		"count":          len(upcomingTasks),
		"limit":          limit,
	})
}

func (ctrl *TaskController) GetRecentActivity(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limit := 20 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent activity"})
		return
	}

	var activities []*models.RecentActivity

	// Add task creation activities
	for _, task := range tasks {
		activity := &models.RecentActivity{
			ID:          task.ID,
			Type:        "task_created",
			Description: "Created task: " + task.TaskName,
			TaskName:    &task.TaskName,
			CreatedAt:   task.CreatedAt,
		}
		activities = append(activities, activity)
	}

	// Sort by creation time (newest first) - simple approach
	// In a real implementation, you'd query logs table for proper activity tracking

	// Limit results
	if len(activities) > limit {
		activities = activities[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"recent_activities": activities,
		"count":             len(activities),
		"limit":             limit,
	})
}

func (ctrl *TaskController) GetTasksByCategory(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category parameter is required"})
		return
	}

	tasks, err := ctrl.taskService.GetTasksByCategory(userID, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks by category"})
		return
	}

	response := models.TaskResponse{
		Tasks: tasks,
		Total: len(tasks),
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *TaskController) GetTasksByPriority(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	priorityStr := c.Param("priority")
	priority, err := strconv.ParseInt(priorityStr, 10, 16)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority value"})
		return
	}

	tasks, err := ctrl.taskService.GetTasksByPriority(userID, int16(priority))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks by priority"})
		return
	}

	response := models.TaskResponse{
		Tasks: tasks,
		Total: len(tasks),
	}

	c.JSON(http.StatusOK, response)
}

// Log Controller
type LogController struct {
	logService services.LogService
}

func NewLogController(logService services.LogService) *LogController {
	return &LogController{
		logService: logService,
	}
}

func (ctrl *LogController) GetUserLogs(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Cap at 100
	}

	logs, err := ctrl.logService.GetUserLogs(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": len(logs),
	})
}

// Habit Controller
type HabitController struct {
	habitService services.HabitService
}

func NewHabitController(habitService services.HabitService) *HabitController {
	return &HabitController{
		habitService: habitService,
	}
}

func (ctrl *HabitController) CreateHabit(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	habit, err := ctrl.habitService.CreateHabit(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create habit"})
		return
	}

	c.JSON(http.StatusCreated, habit)
}

func (ctrl *HabitController) GetHabits(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habitType := c.Query("type")
	var habits []*models.Habit
	var err error

	if habitType != "" {
		habits, err = ctrl.habitService.GetHabitsByType(userID, habitType)
	} else {
		habits, err = ctrl.habitService.GetUserHabits(userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get habits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"habits": habits,
		"total":  len(habits),
	})
}

func (ctrl *HabitController) GetHabit(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habitIDStr := c.Param("id")
	habitID, err := strconv.ParseInt(habitIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid habit ID"})
		return
	}

	habit, err := ctrl.habitService.GetHabitByID(habitID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get habit"})
		return
	}

	c.JSON(http.StatusOK, habit)
}

func (ctrl *HabitController) UpdateHabit(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habitIDStr := c.Param("id")
	habitID, err := strconv.ParseInt(habitIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid habit ID"})
		return
	}

	var req models.UpdateHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	habit, err := ctrl.habitService.UpdateHabit(habitID, userID, &req)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update habit"})
		return
	}

	c.JSON(http.StatusOK, habit)
}

func (ctrl *HabitController) DeleteHabit(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habitIDStr := c.Param("id")
	habitID, err := strconv.ParseInt(habitIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid habit ID"})
		return
	}

	err = ctrl.habitService.DeleteHabit(habitID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete habit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Habit deleted successfully"})
}

func (ctrl *HabitController) TrackHabit(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habitIDStr := c.Param("id")
	habitID, err := strconv.ParseInt(habitIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid habit ID"})
		return
	}

	err = ctrl.habitService.TrackHabit(habitID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track habit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Habit tracked successfully"})
}

func (ctrl *HabitController) MarkHabitAchieved(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habitIDStr := c.Param("id")
	habitID, err := strconv.ParseInt(habitIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid habit ID"})
		return
	}

	var req struct {
		IsAchieved bool `json:"is_achieved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ctrl.habitService.MarkHabitAchieved(habitID, userID, req.IsAchieved)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update habit status"})
		return
	}

	status := "not achieved"
	if req.IsAchieved {
		status = "achieved"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Habit marked as " + status,
		"is_achieved": req.IsAchieved,
	})
}

func (ctrl *HabitController) ExportHabits(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	format := c.Query("format")
	if format == "" {
		format = "json"
	}

	if format != "json" && format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supported formats: json, csv"})
		return
	}

	data, err := ctrl.habitService.ExportHabits(userID, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export habits"})
		return
	}

	contentType := "application/json"
	filename := "habits.json"
	if format == "csv" {
		contentType = "text/csv"
		filename = "habits.csv"
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

func (ctrl *HabitController) ImportHabits(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ctrl.habitService.ImportHabits(userID, []byte(req.Data), req.Format)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Habits imported successfully"})
}

// Focus Tools - Pomodoro Methods
func (ctrl *HabitController) StartPomodoroSession(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.StartPomodoroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create pomodoro session (this would be implemented in service)
	session := &models.PomodoroSession{
		UserID:   userID,
		TaskID:   req.TaskID,
		Duration: req.Duration,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Pomodoro session started",
		"session":      session,
		"instructions": "Focus for " + strconv.Itoa(int(req.Duration)) + " minutes",
	})
}

func (ctrl *HabitController) CompletePomodoroSession(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Mark current session as completed
	c.JSON(http.StatusOK, gin.H{
		"message": "Pomodoro session completed",
		"reward":  "Great job! Take a 5-minute break.",
	})
}

func (ctrl *HabitController) GetPomodoroStats(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Return pomodoro statistics
	stats := gin.H{
		"total_sessions":         0,
		"completed_sessions":     0,
		"total_focus_time":       0,
		"average_session_length": 25,
		"productivity_score":     0,
	}

	c.JSON(http.StatusOK, stats)
}

// Goals Management Methods
func (ctrl *HabitController) CreateGoal(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal := &models.Goal{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		TargetValue: req.TargetValue,
		Unit:        req.Unit,
		DueDate:     req.DueDate,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Goal created successfully",
		"goal":    goal,
	})
}

func (ctrl *HabitController) GetGoals(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Return user's goals (placeholder)
	goals := []models.Goal{}

	c.JSON(http.StatusOK, gin.H{
		"goals": goals,
		"total": len(goals),
	})
}

func (ctrl *HabitController) GetGoal(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	// Get goal by ID (placeholder)
	goal := &models.Goal{
		ID:     goalID,
		UserID: userID,
		Title:  "Sample Goal",
	}

	c.JSON(http.StatusOK, goal)
}

func (ctrl *HabitController) UpdateGoal(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	var req models.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Goal updated successfully",
		"goal_id": goalID,
	})
}

func (ctrl *HabitController) DeleteGoal(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Goal deleted successfully",
		"goal_id": goalID,
	})
}

func (ctrl *HabitController) UpdateGoalProgress(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	var req models.UpdateGoalProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Goal progress updated",
		"goal_id":  goalID,
		"progress": req.Progress,
	})
}

// Habit Streaks & Consistency Methods
func (ctrl *HabitController) GetHabitStreak(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid habit ID"})
		return
	}

	habit, err := ctrl.habitService.GetHabitByID(habitID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get habit"})
		}
		return
	}

	streak := &models.HabitStreak{
		HabitID:       habit.ID,
		HabitName:     habit.Name,
		CurrentStreak: 0, // Calculate based on tracking history
		LongestStreak: 0, // Calculate based on tracking history
		LastTracked:   habit.LastTrackedDate,
	}

	c.JSON(http.StatusOK, streak)
}

func (ctrl *HabitController) GetHabitsConsistencyReport(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habits, err := ctrl.habitService.GetUserHabits(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get habits"})
		return
	}

	var streaks []*models.HabitStreak
	for _, habit := range habits {
		streak := &models.HabitStreak{
			HabitID:       habit.ID,
			HabitName:     habit.Name,
			CurrentStreak: 0, // Calculate based on tracking history
			LongestStreak: 0, // Calculate based on tracking history
			LastTracked:   habit.LastTrackedDate,
		}
		streaks = append(streaks, streak)
	}

	report := gin.H{
		"habits_consistency":  streaks,
		"total_habits":        len(habits),
		"average_consistency": 0, // Calculate based on streaks
		"recommendations": []string{
			"Try to track habits daily for better consistency",
			"Set reminders for habit tracking",
			"Start with small, achievable habits",
		},
	}

	c.JSON(http.StatusOK, report)
}

// Analytics Controller
type AnalyticsController struct {
	authService  services.AuthService
	taskService  services.TaskService
	habitService services.HabitService
}

func NewAnalyticsController(authService services.AuthService, taskService services.TaskService, habitService services.HabitService) *AnalyticsController {
	return &AnalyticsController{
		authService:  authService,
		taskService:  taskService,
		habitService: habitService,
	}
}

func (ctrl *AnalyticsController) GetOverview(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userInfo, err := ctrl.authService.GetUserInfo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get overview"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"overview": userInfo.Stats,
		"user":     userInfo.User,
	})
}

func (ctrl *AnalyticsController) GetTaskAnalytics(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task analytics"})
		return
	}

	// Calculate analytics
	analytics := make(map[string]interface{})
	categoryCount := make(map[string]int)
	priorityCount := make(map[int16]int)
	completedCount := 0
	totalCount := len(tasks)

	for _, task := range tasks {
		if task.IsCompleted {
			completedCount++
		}

		if task.Category != nil {
			categoryCount[*task.Category]++
		}

		priorityCount[task.Priority]++
	}

	analytics["total_tasks"] = totalCount
	analytics["completed_tasks"] = completedCount
	analytics["pending_tasks"] = totalCount - completedCount
	analytics["category_breakdown"] = categoryCount
	analytics["priority_breakdown"] = priorityCount

	if totalCount > 0 {
		analytics["completion_rate"] = float64(completedCount) / float64(totalCount) * 100
	} else {
		analytics["completion_rate"] = 0
	}

	c.JSON(http.StatusOK, analytics)
}

func (ctrl *AnalyticsController) GetHabitAnalytics(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habits, err := ctrl.habitService.GetUserHabits(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get habit analytics"})
		return
	}

	// Calculate analytics
	analytics := make(map[string]interface{})
	typeCount := make(map[string]int)
	achievedCount := 0
	totalCount := len(habits)

	for _, habit := range habits {
		if habit.IsAchieved {
			achievedCount++
		}

		typeCount[habit.Type]++
	}

	analytics["total_habits"] = totalCount
	analytics["achieved_habits"] = achievedCount
	analytics["pending_habits"] = totalCount - achievedCount
	analytics["type_breakdown"] = typeCount

	if totalCount > 0 {
		analytics["achievement_rate"] = float64(achievedCount) / float64(totalCount) * 100
	} else {
		analytics["achievement_rate"] = 0
	}

	c.JSON(http.StatusOK, analytics)
}

// Performance Analytics Methods
func (ctrl *AnalyticsController) GetWeeklyPerformance(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get tasks from the last 7 days
	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weekly performance"})
		return
	}

	habits, err := ctrl.habitService.GetUserHabits(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weekly performance"})
		return
	}

	report := ctrl.calculatePerformanceReport(tasks, habits, "weekly")
	c.JSON(http.StatusOK, report)
}

func (ctrl *AnalyticsController) GetMonthlyPerformance(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get tasks from the last 30 days
	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get monthly performance"})
		return
	}

	habits, err := ctrl.habitService.GetUserHabits(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get monthly performance"})
		return
	}

	report := ctrl.calculatePerformanceReport(tasks, habits, "monthly")
	c.JSON(http.StatusOK, report)
}

func (ctrl *AnalyticsController) GetTimeAllocation(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get time allocation"})
		return
	}

	allocation := ctrl.calculateTimeAllocation(tasks)
	c.JSON(http.StatusOK, gin.H{"time_allocation": allocation})
}

func (ctrl *AnalyticsController) GetProductivityTrends(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	tasks, err := ctrl.taskService.GetUserTasks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get productivity trends"})
		return
	}

	trends := ctrl.calculateProductivityTrends(tasks)
	c.JSON(http.StatusOK, gin.H{"productivity_trends": trends})
}

// Helper methods for analytics calculations
func (ctrl *AnalyticsController) calculatePerformanceReport(tasks []*models.Task, habits []*models.Habit, period string) *models.PerformanceReport {
	var tasksCompleted, tasksCreated int64
	var habitsTracked int64

	for _, task := range tasks {
		if task.IsCompleted {
			tasksCompleted++
		}
		tasksCreated++
	}

	for _, habit := range habits {
		if habit.LastTrackedDate != nil {
			habitsTracked++
		}
	}

	var completionRate float64
	if tasksCreated > 0 {
		completionRate = float64(tasksCompleted) / float64(tasksCreated) * 100
	}

	// Simple productivity score calculation
	productivityScore := completionRate*0.7 + (float64(habitsTracked)/float64(len(habits)))*100*0.3

	return &models.PerformanceReport{
		Period:            period,
		TasksCompleted:    tasksCompleted,
		TasksCreated:      tasksCreated,
		CompletionRate:    completionRate,
		HabitsTracked:     habitsTracked,
		PomodoroSessions:  0, // Will be implemented with Pomodoro feature
		ProductivityScore: productivityScore,
	}
}

func (ctrl *AnalyticsController) calculateTimeAllocation(tasks []*models.Task) []*models.TimeAllocation {
	priorityCount := make(map[int16]int64)
	categoryCount := make(map[string]int64)
	total := int64(len(tasks))

	for _, task := range tasks {
		priorityCount[task.Priority]++
		if task.Category != nil {
			categoryCount[*task.Category]++
		} else {
			categoryCount["Uncategorized"]++
		}
	}

	var allocation []*models.TimeAllocation

	// Add priority-based allocation
	for priority, count := range priorityCount {
		percentage := float64(count) / float64(total) * 100
		allocation = append(allocation, &models.TimeAllocation{
			Category:   "Priority " + strconv.Itoa(int(priority)),
			Priority:   priority,
			TaskCount:  count,
			Percentage: percentage,
		})
	}

	return allocation
}

func (ctrl *AnalyticsController) calculateProductivityTrends(tasks []*models.Task) map[string]interface{} {
	trends := make(map[string]interface{})

	// Simple trend calculation - can be enhanced with more sophisticated analysis
	completed := 0
	total := len(tasks)

	for _, task := range tasks {
		if task.IsCompleted {
			completed++
		}
	}

	trends["current_completion_rate"] = float64(completed) / float64(total) * 100
	trends["trend_direction"] = "stable" // This would be calculated based on historical data
	trends["improvement_suggestions"] = []string{
		"Focus on high-priority tasks first",
		"Break large tasks into smaller ones",
		"Use Pomodoro technique for better focus",
	}

	return trends
}

// GetCategories returns all unique categories used by the user
func (ctrl *TaskController) GetCategories(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	categories, err := ctrl.taskService.GetUserCategories(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}

	// Add some default categories if none exist
	if len(categories) == 0 {
		categories = []string{"Work", "Personal", "Shopping", "Health", "Education", "Finance", "Travel", "Family", "Hobbies", "Study"}
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}
