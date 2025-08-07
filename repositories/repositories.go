package repositories

import (
	"database/sql"
	"encoding/json"
	"todo-backend/models"
)

type UserRepository interface {
	CreateUser(username, passwordHash string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(userID int64) (*models.User, error)
	GetUserStats(userID int64) (*models.UserStats, error)
	CheckUsernameExists(username string) (bool, error)
	UpdateProfile(userID int64, profile *models.UpdateProfileRequest) (*models.User, error)
	UpdatePassword(userID int64, newPasswordHash string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(username, passwordHash string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, display_name, email, location, bio, created_at",
		username, passwordHash,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.Location, &user.Bio, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(
		"SELECT id, username, password_hash, display_name, email, location, bio, created_at FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.DisplayName, &user.Email, &user.Location, &user.Bio, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(
		"SELECT id, username, display_name, email, location, bio, created_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.Location, &user.Bio, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateProfile(userID int64, profile *models.UpdateProfileRequest) (*models.User, error) {
	query := `
		UPDATE users 
		SET display_name = COALESCE($2, display_name),
			email = COALESCE($3, email),
			location = COALESCE($4, location),
			bio = COALESCE($5, bio)
		WHERE id = $1
		RETURNING id, username, display_name, email, location, bio, created_at`

	var user models.User
	err := r.db.QueryRow(query, userID, profile.DisplayName, profile.Email, profile.Location, profile.Bio).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.Location, &user.Bio, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdatePassword(userID int64, newPasswordHash string) error {
	_, err := r.db.Exec("UPDATE users SET password_hash = $1 WHERE id = $2", newPasswordHash, userID)
	return err
}

func (r *userRepository) GetUserStats(userID int64) (*models.UserStats, error) {
	var stats models.UserStats

	// Get task statistics
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM tasks WHERE user_id = $1",
		userID,
	).Scan(&stats.TotalTasks)
	if err != nil {
		stats.TotalTasks = 0
	}

	err = r.db.QueryRow(
		"SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND is_completed = true",
		userID,
	).Scan(&stats.CompletedTasks)
	if err != nil {
		stats.CompletedTasks = 0
	}

	stats.PendingTasks = stats.TotalTasks - stats.CompletedTasks

	// Get habit statistics
	err = r.db.QueryRow(
		"SELECT COUNT(*) FROM habits WHERE user_id = $1",
		userID,
	).Scan(&stats.TotalHabits)
	if err != nil {
		stats.TotalHabits = 0
	}

	err = r.db.QueryRow(
		"SELECT COUNT(*) FROM habits WHERE user_id = $1 AND is_achieved = true",
		userID,
	).Scan(&stats.AchievedHabits)
	if err != nil {
		stats.AchievedHabits = 0
	}

	return &stats, nil
}

func (r *userRepository) CheckUsernameExists(username string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE username = $1",
		username,
	).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Task Repository
type TaskRepository interface {
	CreateTask(task *models.Task) (*models.Task, error)
	GetTasksByUserID(userID int64) ([]*models.Task, error)
	GetTaskByID(taskID, userID int64) (*models.Task, error)
	UpdateTask(task *models.Task) (*models.Task, error)
	DeleteTask(taskID, userID int64) error
	MarkTaskCompleted(taskID, userID int64, isCompleted bool) error
	GetTasksByCategory(userID int64, category string) ([]*models.Task, error)
	GetTasksByPriority(userID int64, priority int16) ([]*models.Task, error)
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) CreateTask(task *models.Task) (*models.Task, error) {
	var createdTask models.Task
	err := r.db.QueryRow(`
		INSERT INTO tasks (user_id, task_name, description, category, priority, due_date, is_recurring, recurring_frequency) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id, user_id, task_name, description, category, priority, due_date, is_completed, is_recurring, recurring_frequency, created_at`,
		task.UserID, task.TaskName, task.Description, task.Category, task.Priority, task.DueDate, task.IsRecurring, task.RecurringFrequency,
	).Scan(&createdTask.ID, &createdTask.UserID, &createdTask.TaskName, &createdTask.Description,
		&createdTask.Category, &createdTask.Priority, &createdTask.DueDate, &createdTask.IsCompleted,
		&createdTask.IsRecurring, &createdTask.RecurringFrequency, &createdTask.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &createdTask, nil
}

func (r *taskRepository) GetTasksByUserID(userID int64) ([]*models.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, task_name, description, category, priority, due_date, is_completed, is_recurring, recurring_frequency, created_at 
		FROM tasks WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.UserID, &task.TaskName, &task.Description,
			&task.Category, &task.Priority, &task.DueDate, &task.IsCompleted,
			&task.IsRecurring, &task.RecurringFrequency, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (r *taskRepository) GetTaskByID(taskID, userID int64) (*models.Task, error) {
	var task models.Task
	err := r.db.QueryRow(`
		SELECT id, user_id, task_name, description, category, priority, due_date, is_completed, is_recurring, recurring_frequency, created_at 
		FROM tasks WHERE id = $1 AND user_id = $2`, taskID, userID).Scan(
		&task.ID, &task.UserID, &task.TaskName, &task.Description,
		&task.Category, &task.Priority, &task.DueDate, &task.IsCompleted,
		&task.IsRecurring, &task.RecurringFrequency, &task.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *taskRepository) UpdateTask(task *models.Task) (*models.Task, error) {
	var updatedTask models.Task
	err := r.db.QueryRow(`
		UPDATE tasks SET task_name = $1, description = $2, category = $3, priority = $4, due_date = $5, 
		is_completed = $6, is_recurring = $7, recurring_frequency = $8 
		WHERE id = $9 AND user_id = $10 
		RETURNING id, user_id, task_name, description, category, priority, due_date, is_completed, is_recurring, recurring_frequency, created_at`,
		task.TaskName, task.Description, task.Category, task.Priority, task.DueDate,
		task.IsCompleted, task.IsRecurring, task.RecurringFrequency, task.ID, task.UserID,
	).Scan(&updatedTask.ID, &updatedTask.UserID, &updatedTask.TaskName, &updatedTask.Description,
		&updatedTask.Category, &updatedTask.Priority, &updatedTask.DueDate, &updatedTask.IsCompleted,
		&updatedTask.IsRecurring, &updatedTask.RecurringFrequency, &updatedTask.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &updatedTask, nil
}

func (r *taskRepository) DeleteTask(taskID, userID int64) error {
	result, err := r.db.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", taskID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *taskRepository) MarkTaskCompleted(taskID, userID int64, isCompleted bool) error {
	result, err := r.db.Exec("UPDATE tasks SET is_completed = $1 WHERE id = $2 AND user_id = $3", isCompleted, taskID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *taskRepository) GetTasksByCategory(userID int64, category string) ([]*models.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, task_name, description, category, priority, due_date, is_completed, is_recurring, recurring_frequency, created_at 
		FROM tasks WHERE user_id = $1 AND category = $2 ORDER BY created_at DESC`, userID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.UserID, &task.TaskName, &task.Description,
			&task.Category, &task.Priority, &task.DueDate, &task.IsCompleted,
			&task.IsRecurring, &task.RecurringFrequency, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (r *taskRepository) GetTasksByPriority(userID int64, priority int16) ([]*models.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, task_name, description, category, priority, due_date, is_completed, is_recurring, recurring_frequency, created_at 
		FROM tasks WHERE user_id = $1 AND priority = $2 ORDER BY created_at DESC`, userID, priority)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.UserID, &task.TaskName, &task.Description,
			&task.Category, &task.Priority, &task.DueDate, &task.IsCompleted,
			&task.IsRecurring, &task.RecurringFrequency, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// Log Repository
type LogRepository interface {
	CreateLog(userID *int64, eventType, description string, metadata map[string]interface{}) error
	GetLogsByUserID(userID int64, limit int) ([]*models.Log, error)
	GetLogsByEventType(eventType string, limit int) ([]*models.Log, error)
}

type logRepository struct {
	db *sql.DB
}

func NewLogRepository(db *sql.DB) LogRepository {
	return &logRepository{db: db}
}

func (r *logRepository) CreateLog(userID *int64, eventType, description string, metadata map[string]interface{}) error {
	var metadataJSON []byte
	var err error

	if metadata != nil {
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}

	_, err = r.db.Exec(
		"INSERT INTO logs (user_id, event_type, description, metadata) VALUES ($1, $2, $3, $4)",
		userID, eventType, description, metadataJSON,
	)

	return err
}

func (r *logRepository) GetLogsByUserID(userID int64, limit int) ([]*models.Log, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, event_type, description, metadata, created_at 
		FROM logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.Log
	for rows.Next() {
		var log models.Log
		var metadataJSON []byte

		err := rows.Scan(&log.ID, &log.UserID, &log.EventType, &log.Description, &metadataJSON, &log.CreatedAt)
		if err != nil {
			return nil, err
		}

		if metadataJSON != nil {
			err = json.Unmarshal(metadataJSON, &log.Metadata)
			if err != nil {
				return nil, err
			}
		}

		logs = append(logs, &log)
	}

	return logs, nil
}

func (r *logRepository) GetLogsByEventType(eventType string, limit int) ([]*models.Log, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, event_type, description, metadata, created_at 
		FROM logs WHERE event_type = $1 ORDER BY created_at DESC LIMIT $2`, eventType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.Log
	for rows.Next() {
		var log models.Log
		var metadataJSON []byte

		err := rows.Scan(&log.ID, &log.UserID, &log.EventType, &log.Description, &metadataJSON, &log.CreatedAt)
		if err != nil {
			return nil, err
		}

		if metadataJSON != nil {
			err = json.Unmarshal(metadataJSON, &log.Metadata)
			if err != nil {
				return nil, err
			}
		}

		logs = append(logs, &log)
	}

	return logs, nil
}

// Habit Repository
type HabitRepository interface {
	CreateHabit(habit *models.Habit) (*models.Habit, error)
	GetHabitsByUserID(userID int64) ([]*models.Habit, error)
	GetHabitByID(habitID, userID int64) (*models.Habit, error)
	UpdateHabit(habit *models.Habit) (*models.Habit, error)
	DeleteHabit(habitID, userID int64) error
	MarkHabitAchieved(habitID, userID int64, isAchieved bool) error
	TrackHabit(habitID, userID int64) error
	GetHabitsByType(userID int64, habitType string) ([]*models.Habit, error)
}

type habitRepository struct {
	db *sql.DB
}

func NewHabitRepository(db *sql.DB) HabitRepository {
	return &habitRepository{db: db}
}

func (r *habitRepository) CreateHabit(habit *models.Habit) (*models.Habit, error) {
	query := `
		INSERT INTO habits (user_id, name, type, target_value, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at`

	err := r.db.QueryRow(query, habit.UserID, habit.Name, habit.Type, habit.TargetValue).Scan(
		&habit.ID, &habit.CreatedAt)
	if err != nil {
		return nil, err
	}

	return habit, nil
}

func (r *habitRepository) GetHabitsByUserID(userID int64) ([]*models.Habit, error) {
	query := `
		SELECT id, user_id, name, type, target_value, is_achieved, last_tracked_date, created_at
		FROM habits
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []*models.Habit
	for rows.Next() {
		habit := &models.Habit{}
		err := rows.Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.Type,
			&habit.TargetValue, &habit.IsAchieved, &habit.LastTrackedDate, &habit.CreatedAt)
		if err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}

	return habits, nil
}

func (r *habitRepository) GetHabitByID(habitID, userID int64) (*models.Habit, error) {
	query := `
		SELECT id, user_id, name, type, target_value, is_achieved, last_tracked_date, created_at
		FROM habits
		WHERE id = $1 AND user_id = $2`

	habit := &models.Habit{}
	err := r.db.QueryRow(query, habitID, userID).Scan(
		&habit.ID, &habit.UserID, &habit.Name, &habit.Type,
		&habit.TargetValue, &habit.IsAchieved, &habit.LastTrackedDate, &habit.CreatedAt)
	if err != nil {
		return nil, err
	}

	return habit, nil
}

func (r *habitRepository) UpdateHabit(habit *models.Habit) (*models.Habit, error) {
	query := `
		UPDATE habits
		SET name = $1, type = $2, target_value = $3, is_achieved = $4
		WHERE id = $5 AND user_id = $6
		RETURNING id, user_id, name, type, target_value, is_achieved, last_tracked_date, created_at`

	err := r.db.QueryRow(query, habit.Name, habit.Type, habit.TargetValue,
		habit.IsAchieved, habit.ID, habit.UserID).Scan(
		&habit.ID, &habit.UserID, &habit.Name, &habit.Type,
		&habit.TargetValue, &habit.IsAchieved, &habit.LastTrackedDate, &habit.CreatedAt)
	if err != nil {
		return nil, err
	}

	return habit, nil
}

func (r *habitRepository) DeleteHabit(habitID, userID int64) error {
	query := `DELETE FROM habits WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, habitID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *habitRepository) MarkHabitAchieved(habitID, userID int64, isAchieved bool) error {
	query := `
		UPDATE habits
		SET is_achieved = $1
		WHERE id = $2 AND user_id = $3`

	result, err := r.db.Exec(query, isAchieved, habitID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *habitRepository) TrackHabit(habitID, userID int64) error {
	query := `
		UPDATE habits
		SET last_tracked_date = NOW()
		WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, habitID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *habitRepository) GetHabitsByType(userID int64, habitType string) ([]*models.Habit, error) {
	query := `
		SELECT id, user_id, name, type, target_value, is_achieved, last_tracked_date, created_at
		FROM habits
		WHERE user_id = $1 AND type = $2
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID, habitType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []*models.Habit
	for rows.Next() {
		habit := &models.Habit{}
		err := rows.Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.Type,
			&habit.TargetValue, &habit.IsAchieved, &habit.LastTrackedDate, &habit.CreatedAt)
		if err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}

	return habits, nil
}
