package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"todo-backend/models"
	"todo-backend/repositories"
	"todo-backend/utils"
)

type AuthService interface {
	Register(username, password string) (*models.User, string, error)
	Login(username, password string) (*models.User, string, error)
	GetUserInfo(userID int64) (*models.UserInfo, error)
	CheckUsernameExists(username string) (bool, error)
	UpdateProfile(userID int64, req *models.UpdateProfileRequest) (*models.User, error)
	ChangePassword(userID int64, currentPassword, newPassword string) error
}

type authService struct {
	userRepo repositories.UserRepository
	logRepo  repositories.LogRepository
}

func NewAuthService(userRepo repositories.UserRepository, logRepo repositories.LogRepository) AuthService {
	return &authService{
		userRepo: userRepo,
		logRepo:  logRepo,
	}
}

func (s *authService) Register(username, password string) (*models.User, string, error) {
	// Check if username already exists
	existingUser, err := s.userRepo.GetUserByUsername(username)
	if err != sql.ErrNoRows {
		if err != nil {
			return nil, "", fmt.Errorf("database error: %w", err)
		}
		if existingUser != nil {
			return nil, "", errors.New("username already exists")
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.userRepo.CreateUser(username, hashedPassword)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Log registration event
	metadata := map[string]interface{}{
		"username": username,
		"user_id":  user.ID,
	}
	s.logRepo.CreateLog(&user.ID, "user_registration", "User registered successfully", metadata)

	return user, token, nil
}

func (s *authService) Login(username, password string) (*models.User, string, error) {
	// Get user from database
	user, err := s.userRepo.GetUserByUsername(username)
	if err == sql.ErrNoRows {
		return nil, "", errors.New("invalid username or password")
	}
	if err != nil {
		return nil, "", fmt.Errorf("database error: %w", err)
	}

	// Check password
	if !utils.CheckPassword(password, user.PasswordHash) {
		// Log failed login attempt
		metadata := map[string]interface{}{
			"username": username,
			"reason":   "invalid_password",
		}
		s.logRepo.CreateLog(&user.ID, "failed_login", "Failed login attempt", metadata)
		return nil, "", errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Log successful login
	metadata := map[string]interface{}{
		"username": username,
		"user_id":  user.ID,
	}
	s.logRepo.CreateLog(&user.ID, "user_login", "User logged in successfully", metadata)

	return user, token, nil
}

func (s *authService) GetUserInfo(userID int64) (*models.UserInfo, error) {
	// Get user information
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user statistics
	stats, err := s.userRepo.GetUserStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	userInfo := &models.UserInfo{
		User:  *user,
		Stats: *stats,
	}

	return userInfo, nil
}

func (s *authService) CheckUsernameExists(username string) (bool, error) {
	exists, err := s.userRepo.CheckUsernameExists(username)
	if err != nil {
		return false, fmt.Errorf("failed to check username: %w", err)
	}

	// Log the check (optional, for monitoring)
	metadata := map[string]interface{}{
		"username": username,
		"exists":   exists,
	}
	s.logRepo.CreateLog(nil, "username_check", "Username availability checked", metadata)

	return exists, nil
}

func (s *authService) UpdateProfile(userID int64, req *models.UpdateProfileRequest) (*models.User, error) {
	// Update user profile
	user, err := s.userRepo.UpdateProfile(userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Log profile update
	metadata := map[string]interface{}{
		"user_id": userID,
		"changes": req,
	}
	s.logRepo.CreateLog(&userID, "profile_updated", "User profile updated", metadata)

	return user, nil
}

func (s *authService) ChangePassword(userID int64, currentPassword, newPassword string) error {
	// Get user from database
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check current password
	if !utils.CheckPassword(currentPassword, user.PasswordHash) {
		// Log failed password change attempt
		metadata := map[string]interface{}{
			"user_id": userID,
			"reason":  "invalid_current_password",
		}
		s.logRepo.CreateLog(&userID, "password_change_failed", "Failed password change attempt", metadata)
		return errors.New("invalid current password")
	}

	// Hash new password
	hashedNewPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password in database
	err = s.userRepo.UpdatePassword(userID, hashedNewPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Log successful password change
	metadata := map[string]interface{}{
		"user_id": userID,
	}
	s.logRepo.CreateLog(&userID, "password_changed", "Password changed successfully", metadata)

	return nil
}

// Task Service
type TaskService interface {
	CreateTask(userID int64, req *models.CreateTaskRequest) (*models.Task, error)
	GetUserTasks(userID int64, page, pageSize int) ([]*models.Task, int64, error)
	GetTaskByID(taskID, userID int64) (*models.Task, error)
	UpdateTask(taskID, userID int64, req *models.UpdateTaskRequest) (*models.Task, error)
	DeleteTask(taskID, userID int64) error
	MarkTaskCompleted(taskID, userID int64, isCompleted bool) error
	GetTasksByCategory(userID int64, category string) ([]*models.Task, error)
	GetTasksByPriority(userID int64, priority int16) ([]*models.Task, error)
	ExportTasks(userID int64, format string) ([]byte, error)
	ImportTasks(userID int64, data []byte, format string) error

	// Enhanced Task Management Methods
	GetTasksDueToday(userID int64) ([]*models.Task, error)
	GetTasksDueThisWeek(userID int64) ([]*models.Task, error)
	GetOverdueTasks(userID int64) ([]*models.Task, error)
	SearchTasks(userID int64, query string, filters map[string]interface{}) ([]*models.Task, error)
	DuplicateTask(taskID, userID int64) (*models.Task, error)
	GetDashboardSummary(userID int64) (*models.DashboardSummary, error)
	GetUpcomingTasks(userID int64, limit int) ([]*models.UpcomingTask, error)
	GetRecentActivity(userID int64, limit int) ([]*models.RecentActivity, error)
	GetUserCategories(userID int64) ([]string, error)
}

type taskService struct {
	taskRepo repositories.TaskRepository
	logRepo  repositories.LogRepository
}

func NewTaskService(taskRepo repositories.TaskRepository, logRepo repositories.LogRepository) TaskService {
	return &taskService{
		taskRepo: taskRepo,
		logRepo:  logRepo,
	}
}

func (s *taskService) CreateTask(userID int64, req *models.CreateTaskRequest) (*models.Task, error) {
	task := &models.Task{
		UserID:             userID,
		TaskName:           req.TaskName,
		Description:        req.Description,
		Category:           req.Category,
		Priority:           req.Priority,
		DueDate:            req.DueDate,
		IsRecurring:        req.IsRecurring,
		RecurringFrequency: req.RecurringFrequency,
	}

	createdTask, err := s.taskRepo.CreateTask(task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Log task creation
	metadata := map[string]interface{}{
		"task_id":   createdTask.ID,
		"task_name": createdTask.TaskName,
		"category":  createdTask.Category,
		"priority":  createdTask.Priority,
	}
	s.logRepo.CreateLog(&userID, "task_created", fmt.Sprintf("Task '%s' created", createdTask.TaskName), metadata)

	return createdTask, nil
}

func (s *taskService) GetUserTasks(userID int64, page, pageSize int) ([]*models.Task, int64, error) {
	tasks, total, err := s.taskRepo.GetTasksByUserIDPaginated(userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tasks: %w", err)
	}
	return tasks, total, nil
}

func (s *taskService) GetTaskByID(taskID, userID int64) (*models.Task, error) {
	task, err := s.taskRepo.GetTaskByID(taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (s *taskService) UpdateTask(taskID, userID int64, req *models.UpdateTaskRequest) (*models.Task, error) {
	// Get existing task first
	existingTask, err := s.taskRepo.GetTaskByID(taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// Update fields
	if req.TaskName != nil {
		existingTask.TaskName = *req.TaskName
	}
	if req.Description != nil {
		existingTask.Description = req.Description
	}
	if req.Category != nil {
		existingTask.Category = req.Category
	}
	if req.Priority != nil {
		existingTask.Priority = *req.Priority
	}
	if req.DueDate != nil {
		existingTask.DueDate = req.DueDate
	}
	if req.IsCompleted != nil {
		existingTask.IsCompleted = *req.IsCompleted
	}
	if req.IsRecurring != nil {
		existingTask.IsRecurring = *req.IsRecurring
	}
	if req.RecurringFrequency != nil {
		existingTask.RecurringFrequency = req.RecurringFrequency
	}

	updatedTask, err := s.taskRepo.UpdateTask(existingTask)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Log task update
	metadata := map[string]interface{}{
		"task_id":   updatedTask.ID,
		"task_name": updatedTask.TaskName,
		"changes":   req,
	}
	s.logRepo.CreateLog(&userID, "task_updated", fmt.Sprintf("Task '%s' updated", updatedTask.TaskName), metadata)

	return updatedTask, nil
}

func (s *taskService) DeleteTask(taskID, userID int64) error {
	// Get task details for logging before deletion
	task, err := s.taskRepo.GetTaskByID(taskID, userID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	err = s.taskRepo.DeleteTask(taskID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Log task deletion
	metadata := map[string]interface{}{
		"task_id":   taskID,
		"task_name": task.TaskName,
	}
	s.logRepo.CreateLog(&userID, "task_deleted", fmt.Sprintf("Task '%s' deleted", task.TaskName), metadata)

	return nil
}

func (s *taskService) MarkTaskCompleted(taskID, userID int64, isCompleted bool) error {
	// Get task details for logging
	task, err := s.taskRepo.GetTaskByID(taskID, userID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	err = s.taskRepo.MarkTaskCompleted(taskID, userID, isCompleted)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Log task completion status change
	status := "completed"
	if !isCompleted {
		status = "uncompleted"
	}

	metadata := map[string]interface{}{
		"task_id":      taskID,
		"task_name":    task.TaskName,
		"is_completed": isCompleted,
	}
	s.logRepo.CreateLog(&userID, "task_status_changed", fmt.Sprintf("Task '%s' marked as %s", task.TaskName, status), metadata)

	return nil
}

func (s *taskService) GetTasksByCategory(userID int64, category string) ([]*models.Task, error) {
	tasks, err := s.taskRepo.GetTasksByCategory(userID, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by category: %w", err)
	}
	return tasks, nil
}

func (s *taskService) GetTasksByPriority(userID int64, priority int16) ([]*models.Task, error) {
	tasks, err := s.taskRepo.GetTasksByPriority(userID, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by priority: %w", err)
	}
	return tasks, nil
}

func (s *taskService) ExportTasks(userID int64, format string) ([]byte, error) {
	tasks, _, err := s.taskRepo.GetTasksByUserIDPaginated(userID, 1, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for export: %w", err)
	}

	var exportData []byte
	switch format {
	case "json":
		exportData, err = utils.ExportTasksAsJSON(tasks)
	case "csv":
		exportData, err = utils.ExportTasksAsCSV(tasks)
	default:
		return nil, errors.New("unsupported export format")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to export tasks: %w", err)
	}

	// Log export
	metadata := map[string]interface{}{
		"format":     format,
		"task_count": len(tasks),
	}
	s.logRepo.CreateLog(&userID, "tasks_exported", fmt.Sprintf("Exported %d tasks in %s format", len(tasks), format), metadata)

	return exportData, nil
}

func (s *taskService) ImportTasks(userID int64, data []byte, format string) error {
	var tasks []*models.Task
	var err error

	switch format {
	case "json":
		tasks, err = utils.ImportTasksFromJSON(data)
	case "csv":
		tasks, err = utils.ImportTasksFromCSV(data)
	default:
		return errors.New("unsupported import format")
	}

	if err != nil {
		return fmt.Errorf("failed to parse import data: %w", err)
	}

	// Create tasks
	var importedCount int
	for _, task := range tasks {
		task.UserID = userID // Ensure task belongs to current user
		_, err := s.taskRepo.CreateTask(task)
		if err != nil {
			// Log error but continue with other tasks
			continue
		}
		importedCount++
	}

	// Log import
	metadata := map[string]interface{}{
		"format":         format,
		"total_tasks":    len(tasks),
		"imported_tasks": importedCount,
		"failed_imports": len(tasks) - importedCount,
	}
	s.logRepo.CreateLog(&userID, "tasks_imported", fmt.Sprintf("Imported %d/%d tasks from %s", importedCount, len(tasks), format), metadata)

	return nil
}

// Enhanced Task Management Methods Implementation
func (s *taskService) GetTasksDueToday(userID int64) ([]*models.Task, error) {
	tasksChan := make(chan []*models.Task)
	errChan := make(chan error)
	go func() {
		tasks, _, err := s.taskRepo.GetTasksByUserIDPaginated(userID, 1, 10000)
		if err != nil {
			errChan <- err
			return
		}
		tasksChan <- tasks
	}()

	var tasks []*models.Task
	select {
	case t := <-tasksChan:
		tasks = t
	case err := <-errChan:
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	dueTodayTasks := make([]*models.Task, 0)
	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.Format("2006-01-02") == today {
			dueTodayTasks = append(dueTodayTasks, task)
		}
	}

	go func() {
		metadata := map[string]interface{}{
			"user_id":    userID,
			"query_type": "due_today",
			"count":      len(dueTodayTasks),
		}
		s.logRepo.CreateLog(&userID, "tasks_queried", fmt.Sprintf("Retrieved %d tasks due today", len(dueTodayTasks)), metadata)
	}()

	return dueTodayTasks, nil
}

func (s *taskService) GetTasksDueThisWeek(userID int64) ([]*models.Task, error) {
	tasksChan := make(chan []*models.Task)
	errChan := make(chan error)
	go func() {
		tasks, _, err := s.taskRepo.GetTasksByUserIDPaginated(userID, 1, 10000)
		if err != nil {
			errChan <- err
			return
		}
		tasksChan <- tasks
	}()

	var tasks []*models.Task
	select {
	case t := <-tasksChan:
		tasks = t
	case err := <-errChan:
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	weekTasks := make([]*models.Task, 0)
	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.After(weekStart) && task.DueDate.Before(weekEnd) {
			weekTasks = append(weekTasks, task)
		}
	}

	go func() {
		metadata := map[string]interface{}{
			"user_id":    userID,
			"query_type": "due_this_week",
			"count":      len(weekTasks),
			"week_start": weekStart.Format("2006-01-02"),
			"week_end":   weekEnd.Format("2006-01-02"),
		}
		s.logRepo.CreateLog(&userID, "tasks_queried", fmt.Sprintf("Retrieved %d tasks due this week", len(weekTasks)), metadata)
	}()

	return weekTasks, nil
}

func (s *taskService) GetOverdueTasks(userID int64) ([]*models.Task, error) {
	tasksChan := make(chan []*models.Task)
	errChan := make(chan error)
	go func() {
		tasks, _, err := s.taskRepo.GetTasksByUserIDPaginated(userID, 1, 10000)
		if err != nil {
			errChan <- err
			return
		}
		tasksChan <- tasks
	}()

	var tasks []*models.Task
	select {
	case t := <-tasksChan:
		tasks = t
	case err := <-errChan:
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	now := time.Now()
	overdueTasks := make([]*models.Task, 0)
	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.Before(now) && !task.IsCompleted {
			overdueTasks = append(overdueTasks, task)
		}
	}

	go func() {
		metadata := map[string]interface{}{
			"user_id":    userID,
			"query_type": "overdue",
			"count":      len(overdueTasks),
		}
		s.logRepo.CreateLog(&userID, "tasks_queried", fmt.Sprintf("Retrieved %d overdue tasks", len(overdueTasks)), metadata)
	}()

	return overdueTasks, nil
}

func (s *taskService) SearchTasks(userID int64, query string, filters map[string]interface{}) ([]*models.Task, error) {
	tasks, _, err := s.taskRepo.GetTasksByUserIDPaginated(userID, 1, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	filteredTasks := make([]*models.Task, 0, len(tasks))
	numWorkers := 8
	taskChan := make(chan *models.Task)
	resultChan := make(chan *models.Task)
	doneChan := make(chan struct{})

	// Worker pool for filtering
	for i := 0; i < numWorkers; i++ {
		go func() {
			for task := range taskChan {
				match := false
				if query != "" {
					if strings.Contains(strings.ToLower(task.TaskName), strings.ToLower(query)) {
						match = true
					} else if task.Description != nil && strings.Contains(strings.ToLower(*task.Description), strings.ToLower(query)) {
						match = true
					}
				} else {
					match = true
				}
				if match {
					resultChan <- task
				}
			}
			doneChan <- struct{}{}
		}()
	}

	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		close(taskChan)
	}()

	go func() {
		finished := 0
		for {
			select {
			case t := <-resultChan:
				filteredTasks = append(filteredTasks, t)
			case <-doneChan:
				finished++
				if finished == numWorkers {
					close(resultChan)
					return
				}
			}
		}
	}()

	// Wait for all workers to finish
	for i := 0; i < numWorkers; i++ {
		<-doneChan
	}

	go func() {
		metadata := map[string]interface{}{
			"user_id": userID,
			"query":   query,
			"filters": filters,
			"results": len(filteredTasks),
		}
		s.logRepo.CreateLog(&userID, "tasks_searched", fmt.Sprintf("Searched tasks with query '%s', found %d results", query, len(filteredTasks)), metadata)
	}()

	return filteredTasks, nil
}

func (s *taskService) DuplicateTask(taskID, userID int64) (*models.Task, error) {
	// Get original task
	originalTask, err := s.taskRepo.GetTaskByID(taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("original task not found: %w", err)
	}

	// Create duplicate with modified name
	duplicateTask := &models.Task{
		UserID:             userID,
		TaskName:           originalTask.TaskName + " (Copy)",
		Description:        originalTask.Description,
		Category:           originalTask.Category,
		Priority:           originalTask.Priority,
		DueDate:            originalTask.DueDate,
		IsRecurring:        originalTask.IsRecurring,
		RecurringFrequency: originalTask.RecurringFrequency,
		IsCompleted:        false, // New task should not be completed
	}

	createdTask, err := s.taskRepo.CreateTask(duplicateTask)
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate task: %w", err)
	}

	// Log duplication
	metadata := map[string]interface{}{
		"original_task_id": taskID,
		"new_task_id":      createdTask.ID,
		"original_name":    originalTask.TaskName,
		"new_name":         createdTask.TaskName,
	}
	s.logRepo.CreateLog(&userID, "task_duplicated", fmt.Sprintf("Task '%s' duplicated as '%s'", originalTask.TaskName, createdTask.TaskName), metadata)

	return createdTask, nil
}

func (s *taskService) GetDashboardSummary(userID int64) (*models.DashboardSummary, error) {
	tasksChan := make(chan []*models.Task)
	errChan := make(chan error)
	go func() {
		tasks, _, err := s.taskRepo.GetTasksByUserIDPaginated(userID, 1, 10000)
		if err != nil {
			errChan <- err
			return
		}
		tasksChan <- tasks
	}()

	var tasks []*models.Task
	select {
	case t := <-tasksChan:
		tasks = t
	case err := <-errChan:
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	summary := &models.DashboardSummary{
		TotalTasks: int64(len(tasks)),
	}

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

	return summary, nil
}

func (s *taskService) GetUpcomingTasks(userID int64, limit int) ([]*models.UpcomingTask, error) {
	tasksChan := make(chan []*models.Task)
	errChan := make(chan error)
	go func() {
		tasks, _, err := s.taskRepo.GetTasksByUserIDPaginated(userID, 1, 10000)
		if err != nil {
			errChan <- err
			return
		}
		tasksChan <- tasks
	}()

	var tasks []*models.Task
	select {
	case t := <-tasksChan:
		tasks = t
	case err := <-errChan:
		return nil, fmt.Errorf("failed to get upcoming tasks: %w", err)
	}

	now := time.Now()
	upcomingTasks := make([]*models.UpcomingTask, 0)
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
				IsUrgent: daysLeft <= 3,
			}
			upcomingTasks = append(upcomingTasks, upcomingTask)
			if len(upcomingTasks) >= limit {
				break
			}
		}
	}

	return upcomingTasks, nil
}

func (s *taskService) GetRecentActivity(userID int64, limit int) ([]*models.RecentActivity, error) {
	// Get recent logs for the user
	logs, err := s.logRepo.GetLogsByUserID(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	var activities []*models.RecentActivity
	for _, log := range logs {
		activity := &models.RecentActivity{
			ID:          log.ID,
			Type:        log.EventType,
			Description: log.Description,
			CreatedAt:   log.CreatedAt,
		}

		// Extract task name from metadata if available
		if taskName, exists := log.Metadata["task_name"]; exists {
			if name, ok := taskName.(string); ok {
				activity.TaskName = &name
			}
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// Log Service
type LogService interface {
	GetUserLogs(userID int64, limit int) ([]*models.Log, error)
	GetSystemLogs(eventType string, limit int) ([]*models.Log, error)
	CreateSystemLog(eventType, description string, metadata map[string]interface{}) error
}

type logService struct {
	logRepo repositories.LogRepository
}

func NewLogService(logRepo repositories.LogRepository) LogService {
	return &logService{
		logRepo: logRepo,
	}
}

func (s *logService) GetUserLogs(userID int64, limit int) ([]*models.Log, error) {
	logs, err := s.logRepo.GetLogsByUserID(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user logs: %w", err)
	}
	return logs, nil
}

func (s *logService) GetSystemLogs(eventType string, limit int) ([]*models.Log, error) {
	logs, err := s.logRepo.GetLogsByEventType(eventType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get system logs: %w", err)
	}
	return logs, nil
}

func (s *logService) CreateSystemLog(eventType, description string, metadata map[string]interface{}) error {
	err := s.logRepo.CreateLog(nil, eventType, description, metadata)
	if err != nil {
		return fmt.Errorf("failed to create system log: %w", err)
	}
	return nil
}

// Habit Service
type HabitService interface {
	CreateHabit(userID int64, req *models.CreateHabitRequest) (*models.Habit, error)
	GetUserHabits(userID int64, page, pageSize int) ([]*models.Habit, int64, error)
	GetHabitByID(habitID, userID int64) (*models.Habit, error)
	UpdateHabit(habitID, userID int64, req *models.UpdateHabitRequest) (*models.Habit, error)
	DeleteHabit(habitID, userID int64) error
	MarkHabitAchieved(habitID, userID int64, isAchieved bool) error
	TrackHabit(habitID, userID int64) error
	GetHabitsByType(userID int64, habitType string) ([]*models.Habit, error)
	ExportHabits(userID int64, format string) ([]byte, error)
	ImportHabits(userID int64, data []byte, format string) error

	// Pomodoro methods
	StartPomodoroSession(userID int64, req *models.StartPomodoroRequest) (*models.PomodoroSession, error)
	CompletePomodoroSession(sessionID, userID int64) (*models.PomodoroSession, error)
	GetPomodoroStats(userID int64) (*models.PomodoroStats, error)

	// Goal methods
	CreateGoal(userID int64, req *models.CreateGoalRequest) (*models.Goal, error)
	GetGoals(userID int64) ([]*models.Goal, error)
	GetGoal(goalID, userID int64) (*models.Goal, error)
	UpdateGoal(goalID, userID int64, req *models.UpdateGoalRequest) (*models.Goal, error)
	DeleteGoal(goalID, userID int64) error
	UpdateGoalProgress(goalID, userID int64, req *models.UpdateGoalProgressRequest) (*models.Goal, error)
}

type habitService struct {
	habitRepo    repositories.HabitRepository
	logRepo      repositories.LogRepository
	pomodoroRepo repositories.PomodoroRepository
	goalRepo     repositories.GoalRepository
}

func NewHabitService(habitRepo repositories.HabitRepository, logRepo repositories.LogRepository, pomodoroRepo repositories.PomodoroRepository, goalRepo repositories.GoalRepository) HabitService {
	return &habitService{
		habitRepo:    habitRepo,
		logRepo:      logRepo,
		pomodoroRepo: pomodoroRepo,
		goalRepo:     goalRepo,
	}
}

func (s *habitService) CreateHabit(userID int64, req *models.CreateHabitRequest) (*models.Habit, error) {
	habit := &models.Habit{
		UserID:      userID,
		Name:        req.Name,
		Type:        req.Type,
		TargetValue: req.TargetValue,
	}

	createdHabit, err := s.habitRepo.CreateHabit(habit)
	if err != nil {
		return nil, fmt.Errorf("failed to create habit: %w", err)
	}

	// Log habit creation
	metadata := map[string]interface{}{
		"habit_id":     createdHabit.ID,
		"habit_name":   createdHabit.Name,
		"habit_type":   createdHabit.Type,
		"target_value": createdHabit.TargetValue,
	}
	s.logRepo.CreateLog(&userID, "habit_created", fmt.Sprintf("Habit '%s' created", createdHabit.Name), metadata)

	return createdHabit, nil
}

func (s *habitService) GetUserHabits(userID int64, page, pageSize int) ([]*models.Habit, int64, error) {
	habits, total, err := s.habitRepo.GetHabitsByUserIDPaginated(userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get habits: %w", err)
	}
	return habits, total, nil
}

func (s *habitService) GetHabitByID(habitID, userID int64) (*models.Habit, error) {
	habit, err := s.habitRepo.GetHabitByID(habitID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get habit: %w", err)
	}
	return habit, nil
}

func (s *habitService) UpdateHabit(habitID, userID int64, req *models.UpdateHabitRequest) (*models.Habit, error) {
	// Get existing habit first
	existingHabit, err := s.habitRepo.GetHabitByID(habitID, userID)
	if err != nil {
		return nil, fmt.Errorf("habit not found: %w", err)
	}

	// Update fields
	if req.Name != nil {
		existingHabit.Name = *req.Name
	}
	if req.Type != nil {
		existingHabit.Type = *req.Type
	}
	if req.TargetValue != nil {
		existingHabit.TargetValue = req.TargetValue
	}
	if req.IsAchieved != nil {
		existingHabit.IsAchieved = *req.IsAchieved
	}

	updatedHabit, err := s.habitRepo.UpdateHabit(existingHabit)
	if err != nil {
		return nil, fmt.Errorf("failed to update habit: %w", err)
	}

	// Log habit update
	metadata := map[string]interface{}{
		"habit_id":   updatedHabit.ID,
		"habit_name": updatedHabit.Name,
		"changes":    req,
	}
	s.logRepo.CreateLog(&userID, "habit_updated", fmt.Sprintf("Habit '%s' updated", updatedHabit.Name), metadata)

	return updatedHabit, nil
}

func (s *habitService) DeleteHabit(habitID, userID int64) error {
	// Get habit details for logging before deletion
	habit, err := s.habitRepo.GetHabitByID(habitID, userID)
	if err != nil {
		return fmt.Errorf("habit not found: %w", err)
	}

	err = s.habitRepo.DeleteHabit(habitID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete habit: %w", err)
	}

	// Log habit deletion
	metadata := map[string]interface{}{
		"habit_id":   habitID,
		"habit_name": habit.Name,
	}
	s.logRepo.CreateLog(&userID, "habit_deleted", fmt.Sprintf("Habit '%s' deleted", habit.Name), metadata)

	return nil
}

func (s *habitService) MarkHabitAchieved(habitID, userID int64, isAchieved bool) error {
	// Get habit details for logging
	habit, err := s.habitRepo.GetHabitByID(habitID, userID)
	if err != nil {
		return fmt.Errorf("habit not found: %w", err)
	}

	err = s.habitRepo.MarkHabitAchieved(habitID, userID, isAchieved)
	if err != nil {
		return fmt.Errorf("failed to update habit achievement status: %w", err)
	}

	// Log habit achievement status change
	status := "achieved"
	if !isAchieved {
		status = "not achieved"
	}

	metadata := map[string]interface{}{
		"habit_id":    habitID,
		"habit_name":  habit.Name,
		"is_achieved": isAchieved,
	}
	s.logRepo.CreateLog(&userID, "habit_achievement_changed", fmt.Sprintf("Habit '%s' marked as %s", habit.Name, status), metadata)

	return nil
}

func (s *habitService) TrackHabit(habitID, userID int64) error {
	// Get habit details for logging
	habit, err := s.habitRepo.GetHabitByID(habitID, userID)
	if err != nil {
		return fmt.Errorf("habit not found: %w", err)
	}

	err = s.habitRepo.TrackHabit(habitID, userID)
	if err != nil {
		return fmt.Errorf("failed to track habit: %w", err)
	}

	// Log habit tracking
	metadata := map[string]interface{}{
		"habit_id":   habitID,
		"habit_name": habit.Name,
	}
	s.logRepo.CreateLog(&userID, "habit_tracked", fmt.Sprintf("Habit '%s' tracked", habit.Name), metadata)

	return nil
}

func (s *habitService) GetHabitsByType(userID int64, habitType string) ([]*models.Habit, error) {
	habits, err := s.habitRepo.GetHabitsByType(userID, habitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get habits by type: %w", err)
	}
	return habits, nil
}

func (s *habitService) ExportHabits(userID int64, format string) ([]byte, error) {
	habits, _, err := s.habitRepo.GetHabitsByUserIDPaginated(userID, 1, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get habits for export: %w", err)
	}

	var exportData []byte
	switch format {
	case "json":
		exportData, err = utils.ExportHabitsAsJSON(habits)
	case "csv":
		exportData, err = utils.ExportHabitsAsCSV(habits)
	default:
		return nil, errors.New("unsupported export format")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to export habits: %w", err)
	}

	// Log export
	metadata := map[string]interface{}{
		"format":      format,
		"habit_count": len(habits),
	}
	s.logRepo.CreateLog(&userID, "habits_exported", fmt.Sprintf("Exported %d habits in %s format", len(habits), format), metadata)

	return exportData, nil
}

func (s *habitService) ImportHabits(userID int64, data []byte, format string) error {
	var habits []*models.Habit
	var err error

	switch format {
	case "json":
		habits, err = utils.ImportHabitsFromJSON(data)
	case "csv":
		habits, err = utils.ImportHabitsFromCSV(data)
	default:
		return errors.New("unsupported import format")
	}

	if err != nil {
		return fmt.Errorf("failed to parse import data: %w", err)
	}

	// Create habits
	var importedCount int
	for _, habit := range habits {
		habit.UserID = userID // Ensure habit belongs to current user
		_, err := s.habitRepo.CreateHabit(habit)
		if err != nil {
			// Log error but continue with other habits
			continue
		}
		importedCount++
	}

	// Log import
	metadata := map[string]interface{}{
		"format":          format,
		"total_habits":    len(habits),
		"imported_habits": importedCount,
		"failed_imports":  len(habits) - importedCount,
	}
	s.logRepo.CreateLog(&userID, "habits_imported", fmt.Sprintf("Imported %d/%d habits from %s", importedCount, len(habits), format), metadata)

	return nil
}

// Pomodoro Session methods
func (s *habitService) StartPomodoroSession(userID int64, req *models.StartPomodoroRequest) (*models.PomodoroSession, error) {
	session := &models.PomodoroSession{
		UserID:    userID,
		TaskID:    req.TaskID,
		Duration:  req.Duration,
		StartedAt: time.Now(),
	}

	createdSession, err := s.pomodoroRepo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to start pomodoro session: %w", err)
	}

	// Log session start
	metadata := map[string]interface{}{
		"session_id": createdSession.ID,
		"duration":   createdSession.Duration,
		"task_id":    createdSession.TaskID,
	}
	s.logRepo.CreateLog(&userID, "pomodoro_started", fmt.Sprintf("Started %d-minute pomodoro session", createdSession.Duration), metadata)

	return createdSession, nil
}

func (s *habitService) CompletePomodoroSession(sessionID, userID int64) (*models.PomodoroSession, error) {
	// Get session to verify ownership
	session, err := s.pomodoroRepo.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session.UserID != userID {
		return nil, errors.New("session not found")
	}

	// Complete the session
	err = s.pomodoroRepo.CompleteSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to complete session: %w", err)
	}

	// Get updated session
	updatedSession, err := s.pomodoroRepo.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated session: %w", err)
	}

	// Log session completion
	metadata := map[string]interface{}{
		"session_id": sessionID,
		"duration":   updatedSession.Duration,
		"task_id":    updatedSession.TaskID,
	}
	s.logRepo.CreateLog(&userID, "pomodoro_completed", fmt.Sprintf("Completed %d-minute pomodoro session", updatedSession.Duration), metadata)

	return updatedSession, nil
}

func (s *habitService) GetPomodoroStats(userID int64) (*models.PomodoroStats, error) {
	stats, err := s.pomodoroRepo.GetSessionStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pomodoro stats: %w", err)
	}

	return stats, nil
}

// Goal methods
func (s *habitService) CreateGoal(userID int64, req *models.CreateGoalRequest) (*models.Goal, error) {
	goal := &models.Goal{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		TargetValue: req.TargetValue,
		Unit:        req.Unit,
		DueDate:     req.DueDate,
	}

	createdGoal, err := s.goalRepo.CreateGoal(goal)
	if err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	// Log goal creation
	metadata := map[string]interface{}{
		"goal_id":      createdGoal.ID,
		"category":     createdGoal.Category,
		"target_value": createdGoal.TargetValue,
		"unit":         createdGoal.Unit,
	}
	s.logRepo.CreateLog(&userID, "goal_created", fmt.Sprintf("Goal '%s' created", createdGoal.Title), metadata)

	return createdGoal, nil
}

func (s *habitService) GetGoals(userID int64) ([]*models.Goal, error) {
	goals, err := s.goalRepo.GetGoalsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goals: %w", err)
	}

	return goals, nil
}

func (s *habitService) GetGoal(goalID, userID int64) (*models.Goal, error) {
	goal, err := s.goalRepo.GetGoalByID(goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	if goal.UserID != userID {
		return nil, errors.New("goal not found")
	}

	return goal, nil
}

func (s *habitService) UpdateGoal(goalID, userID int64, req *models.UpdateGoalRequest) (*models.Goal, error) {
	existingGoal, err := s.goalRepo.GetGoalByID(goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	if existingGoal.UserID != userID {
		return nil, errors.New("goal not found")
	}

	// Update fields if provided
	if req.Title != nil {
		existingGoal.Title = *req.Title
	}
	if req.Description != nil {
		existingGoal.Description = req.Description
	}
	if req.Category != nil {
		existingGoal.Category = *req.Category
	}
	if req.TargetValue != nil {
		existingGoal.TargetValue = *req.TargetValue
	}
	if req.CurrentValue != nil {
		existingGoal.CurrentValue = *req.CurrentValue
	}
	if req.Unit != nil {
		existingGoal.Unit = *req.Unit
	}
	if req.DueDate != nil {
		existingGoal.DueDate = req.DueDate
	}
	if req.IsCompleted != nil {
		existingGoal.IsCompleted = *req.IsCompleted
	}

	updatedGoal, err := s.goalRepo.UpdateGoal(existingGoal)
	if err != nil {
		return nil, fmt.Errorf("failed to update goal: %w", err)
	}

	// Log goal update
	metadata := map[string]interface{}{
		"goal_id":         updatedGoal.ID,
		"previous_status": existingGoal.IsCompleted,
		"new_status":      updatedGoal.IsCompleted,
	}
	s.logRepo.CreateLog(&userID, "goal_updated", fmt.Sprintf("Goal '%s' updated", updatedGoal.Title), metadata)

	return updatedGoal, nil
}

func (s *habitService) DeleteGoal(goalID, userID int64) error {
	goal, err := s.goalRepo.GetGoalByID(goalID)
	if err != nil {
		return fmt.Errorf("failed to get goal: %w", err)
	}

	if goal.UserID != userID {
		return errors.New("goal not found")
	}

	err = s.goalRepo.DeleteGoal(goalID)
	if err != nil {
		return fmt.Errorf("failed to delete goal: %w", err)
	}

	// Log goal deletion
	metadata := map[string]interface{}{
		"goal_id":  goalID,
		"title":    goal.Title,
		"category": goal.Category,
	}
	s.logRepo.CreateLog(&userID, "goal_deleted", fmt.Sprintf("Goal '%s' deleted", goal.Title), metadata)

	return nil
}

func (s *habitService) UpdateGoalProgress(goalID, userID int64, req *models.UpdateGoalProgressRequest) (*models.Goal, error) {
	goal, err := s.goalRepo.GetGoalByID(goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	if goal.UserID != userID {
		return nil, errors.New("goal not found")
	}

	// Update progress
	err = s.goalRepo.UpdateGoalProgress(goalID, req.Progress)
	if err != nil {
		return nil, fmt.Errorf("failed to update goal progress: %w", err)
	}

	// Get updated goal
	updatedGoal, err := s.goalRepo.GetGoalByID(goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated goal: %w", err)
	}

	// Log progress update
	metadata := map[string]interface{}{
		"goal_id":        goalID,
		"previous_value": goal.CurrentValue,
		"new_value":      updatedGoal.CurrentValue,
		"target_value":   updatedGoal.TargetValue,
		"is_completed":   updatedGoal.IsCompleted,
	}
	s.logRepo.CreateLog(&userID, "goal_progress_updated", fmt.Sprintf("Goal '%s' progress updated to %d/%d %s", updatedGoal.Title, updatedGoal.CurrentValue, updatedGoal.TargetValue, updatedGoal.Unit), metadata)

	return updatedGoal, nil
}

// GetUserCategories returns all unique categories used by the user
func (s *taskService) GetUserCategories(userID int64) ([]string, error) {
	tasksChan := make(chan []*models.Task)
	errChan := make(chan error)
	go func() {
		tasks, _, err := s.taskRepo.GetTasksByUserIDPaginated(userID, 1, 10000)
		if err != nil {
			errChan <- err
			return
		}
		tasksChan <- tasks
	}()

	var tasks []*models.Task
	select {
	case t := <-tasksChan:
		tasks = t
	case err := <-errChan:
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}

	categoryMap := make(map[string]bool)
	numWorkers := 8
	taskChan := make(chan *models.Task)
	doneChan := make(chan struct{})

	// Use a mutex to protect categoryMap from concurrent writes
	var mu sync.Mutex

	for i := 0; i < numWorkers; i++ {
		go func() {
			for task := range taskChan {
				if task.Category != nil && *task.Category != "" {
					mu.Lock()
					categoryMap[*task.Category] = true
					mu.Unlock()
				}
			}
			doneChan <- struct{}{}
		}()
	}

	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		close(taskChan)
	}()

	for i := 0; i < numWorkers; i++ {
		<-doneChan
	}

	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}

	return categories, nil
}
