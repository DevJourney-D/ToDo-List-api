package models

import (
	"encoding/json"
	"strings"
	"time"
)

// CustomTime is a custom time type that can handle multiple date formats
type CustomTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler interface
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "null" || s == "" {
		return nil
	}

	// Try different date formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 without timezone
		"2006-01-02T15:04:05",       // ISO 8601 without timezone
		"2006-01-02",                // Date only
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			ct.Time = t
			return nil
		}
	}

	return &time.ParseError{Layout: "multiple formats", Value: s, Message: "cannot parse date"}
}

// MarshalJSON implements json.Marshaler interface
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(ct.Time.Format("2006-01-02T15:04:05Z07:00"))
}

type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	DisplayName  *string   `json:"display_name" db:"display_name"`
	Email        *string   `json:"email" db:"email"`
	Location     *string   `json:"location" db:"location"`
	Bio          *string   `json:"bio" db:"bio"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Task struct {
	ID                 int64       `json:"id" db:"id"`
	UserID             int64       `json:"user_id" db:"user_id"`
	TaskName           string      `json:"task_name" db:"task_name"`
	Description        *string     `json:"description" db:"description"`
	Category           *string     `json:"category" db:"category"`
	Priority           int16       `json:"priority" db:"priority"`
	DueDate            *CustomTime `json:"due_date" db:"due_date"`
	IsCompleted        bool        `json:"is_completed" db:"is_completed"`
	IsRecurring        bool        `json:"is_recurring" db:"is_recurring"`
	RecurringFrequency *string     `json:"recurring_frequency" db:"recurring_frequency"`
	CreatedAt          time.Time   `json:"created_at" db:"created_at"`
}

type Habit struct {
	ID              int64       `json:"id" db:"id"`
	UserID          int64       `json:"user_id" db:"user_id"`
	Name            string      `json:"name" db:"name"`
	Type            string      `json:"type" db:"type"`
	TargetValue     *string     `json:"target_value" db:"target_value"`
	IsAchieved      bool        `json:"is_achieved" db:"is_achieved"`
	LastTrackedDate *CustomTime `json:"last_tracked_date" db:"last_tracked_date"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
}

type Log struct {
	ID          int64                  `json:"id" db:"id"`
	UserID      *int64                 `json:"user_id" db:"user_id"`
	EventType   string                 `json:"event_type" db:"event_type"`
	Description string                 `json:"description" db:"description"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// DTO models for API requests/responses
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CheckUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
}

type CheckUsernameResponse struct {
	Exists bool `json:"exists"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Email       *string `json:"email"`
	Location    *string `json:"location"`
	Bio         *string `json:"bio"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type UserInfo struct {
	User  User      `json:"user"`
	Stats UserStats `json:"stats"`
}

type UserStats struct {
	TotalTasks     int64 `json:"total_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
	PendingTasks   int64 `json:"pending_tasks"`
	TotalHabits    int64 `json:"total_habits"`
	AchievedHabits int64 `json:"achieved_habits"`
}

// Task request/response models
type CreateTaskRequest struct {
	TaskName           string      `json:"task_name"` // Changed from "name" to "task_name"
	Description        *string     `json:"description"`
	Category           *string     `json:"category"`
	Priority           int16       `json:"priority"`
	DueDate            *CustomTime `json:"due_date"`
	IsRecurring        bool        `json:"is_recurring"`
	RecurringFrequency *string     `json:"recurring_frequency"`
}

type UpdateTaskRequest struct {
	TaskName           *string     `json:"task_name"`
	Description        *string     `json:"description"`
	Category           *string     `json:"category"`
	Priority           *int16      `json:"priority"`
	DueDate            *CustomTime `json:"due_date"`
	IsCompleted        *bool       `json:"is_completed"`
	IsRecurring        *bool       `json:"is_recurring"`
	RecurringFrequency *string     `json:"recurring_frequency"`
}

type TaskResponse struct {
	Tasks []*Task `json:"tasks"`
	Total int     `json:"total"`
}

// Habit request/response models
type CreateHabitRequest struct {
	Name        string  `json:"name" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	TargetValue *string `json:"target_value"`
}

type UpdateHabitRequest struct {
	Name        *string `json:"name"`
	Type        *string `json:"type"`
	TargetValue *string `json:"target_value"`
	IsAchieved  *bool   `json:"is_achieved"`
}

// Export/Import models
type ExportRequest struct {
	Format string `json:"format" binding:"required,oneof=json csv"`
}

type ImportRequest struct {
	Format string `json:"format" binding:"required,oneof=json csv"`
	Data   string `json:"data" binding:"required"`
}

// Personal Growth Models
type Goal struct {
	ID           int64      `json:"id" db:"id"`
	UserID       int64      `json:"user_id" db:"user_id"`
	Title        string     `json:"title" db:"title"`
	Description  *string    `json:"description" db:"description"`
	Category     string     `json:"category" db:"category"` // short-term, long-term
	TargetValue  int32      `json:"target_value" db:"target_value"`
	CurrentValue int32      `json:"current_value" db:"current_value"`
	Unit         string     `json:"unit" db:"unit"` // tasks, hours, days, etc.
	DueDate      *time.Time `json:"due_date" db:"due_date"`
	IsCompleted  bool       `json:"is_completed" db:"is_completed"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type PomodoroSession struct {
	ID          int64      `json:"id" db:"id"`
	UserID      int64      `json:"user_id" db:"user_id"`
	TaskID      *int64     `json:"task_id" db:"task_id"`
	Duration    int32      `json:"duration" db:"duration"` // in minutes
	IsCompleted bool       `json:"is_completed" db:"is_completed"`
	StartedAt   time.Time  `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}

type PomodoroStats struct {
	TotalSessions     int64 `json:"total_sessions"`
	CompletedSessions int64 `json:"completed_sessions"`
	TotalMinutes      int64 `json:"total_minutes"`
}

// Analytics Models
type PerformanceReport struct {
	Period            string  `json:"period"` // weekly, monthly
	TasksCompleted    int64   `json:"tasks_completed"`
	TasksCreated      int64   `json:"tasks_created"`
	CompletionRate    float64 `json:"completion_rate"`
	HabitsTracked     int64   `json:"habits_tracked"`
	PomodoroSessions  int64   `json:"pomodoro_sessions"`
	ProductivityScore float64 `json:"productivity_score"`
}

type TimeAllocation struct {
	Category   string  `json:"category"`
	Priority   int16   `json:"priority"`
	TaskCount  int64   `json:"task_count"`
	Percentage float64 `json:"percentage"`
}

type HabitStreak struct {
	HabitID       int64       `json:"habit_id"`
	HabitName     string      `json:"habit_name"`
	CurrentStreak int32       `json:"current_streak"`
	LongestStreak int32       `json:"longest_streak"`
	LastTracked   *CustomTime `json:"last_tracked"`
}

// Request/Response Models for Personal Growth
type CreateGoalRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description"`
	Category    string     `json:"category" binding:"required,oneof=short-term long-term"`
	TargetValue int32      `json:"target_value" binding:"required,min=1"`
	Unit        string     `json:"unit" binding:"required"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateGoalRequest struct {
	Title        *string    `json:"title"`
	Description  *string    `json:"description"`
	Category     *string    `json:"category"`
	TargetValue  *int32     `json:"target_value"`
	CurrentValue *int32     `json:"current_value"`
	Unit         *string    `json:"unit"`
	DueDate      *time.Time `json:"due_date"`
	IsCompleted  *bool      `json:"is_completed"`
}

type UpdateGoalProgressRequest struct {
	Progress int32 `json:"progress" binding:"required,min=0"`
}

type StartPomodoroRequest struct {
	TaskID   *int64 `json:"task_id"`
	Duration int32  `json:"duration" binding:"required,min=1,max=60"` // 1-60 minutes
}

// Enhanced Task Management Models
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type UpdateTaskStatusRequest struct {
	Status TaskStatus `json:"status" binding:"required,oneof=pending in_progress completed cancelled"`
}

type RescheduleTaskRequest struct {
	NewDueDate CustomTime `json:"new_due_date" binding:"required"`
	Reason     *string    `json:"reason"`
}

type SearchTasksRequest struct {
	Query       string     `json:"query"`
	Category    *string    `json:"category"`
	Priority    *int16     `json:"priority"`
	Status      *string    `json:"status"`
	DueDateFrom *time.Time `json:"due_date_from"`
	DueDateTo   *time.Time `json:"due_date_to"`
}

// Dashboard Models
type DashboardSummary struct {
	TotalTasks        int64   `json:"total_tasks"`
	CompletedToday    int64   `json:"completed_today"`
	DueToday          int64   `json:"due_today"`
	Overdue           int64   `json:"overdue"`
	CompletionRate    float64 `json:"completion_rate"`
	ProductivityTrend string  `json:"productivity_trend"` // up, down, stable
}

type UpcomingTask struct {
	ID       int64       `json:"id"`
	TaskName string      `json:"task_name"`
	Category *string     `json:"category"`
	Priority int16       `json:"priority"`
	DueDate  *CustomTime `json:"due_date"`
	DaysLeft int         `json:"days_left"`
	IsUrgent bool        `json:"is_urgent"`
}

type RecentActivity struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"` // task_created, task_completed, habit_tracked
	Description string    `json:"description"`
	TaskName    *string   `json:"task_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type CalendarView struct {
	Date  string  `json:"date"`
	Tasks []*Task `json:"tasks"`
	Count int     `json:"count"`
}
