package repositories

import (
	"database/sql"
	"todo-backend/models"
)

// Pomodoro Session Repository
type PomodoroRepository interface {
	CreateSession(session *models.PomodoroSession) (*models.PomodoroSession, error)
	GetSessionsByUserID(userID int64) ([]*models.PomodoroSession, error)
	GetSessionByID(sessionID int64) (*models.PomodoroSession, error)
	UpdateSession(session *models.PomodoroSession) (*models.PomodoroSession, error)
	CompleteSession(sessionID int64) error
	DeleteSession(sessionID int64) error
	GetSessionStats(userID int64) (*models.PomodoroStats, error)
}

type pomodoroRepository struct {
	db *sql.DB
}

func NewPomodoroRepository(db *sql.DB) PomodoroRepository {
	return &pomodoroRepository{db: db}
}

func (r *pomodoroRepository) CreateSession(session *models.PomodoroSession) (*models.PomodoroSession, error) {
	var createdSession models.PomodoroSession
	err := r.db.QueryRow(`
		INSERT INTO pomodoro_sessions (user_id, task_id, duration, started_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, user_id, task_id, duration, is_completed, started_at, completed_at`,
		session.UserID, session.TaskID, session.Duration, session.StartedAt,
	).Scan(&createdSession.ID, &createdSession.UserID, &createdSession.TaskID,
		&createdSession.Duration, &createdSession.IsCompleted, &createdSession.StartedAt, &createdSession.CompletedAt)

	if err != nil {
		return nil, err
	}

	return &createdSession, nil
}

func (r *pomodoroRepository) GetSessionsByUserID(userID int64) ([]*models.PomodoroSession, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, task_id, duration, is_completed, started_at, completed_at 
		FROM pomodoro_sessions 
		WHERE user_id = $1 
		ORDER BY started_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.PomodoroSession
	for rows.Next() {
		session := &models.PomodoroSession{}
		err := rows.Scan(&session.ID, &session.UserID, &session.TaskID, &session.Duration,
			&session.IsCompleted, &session.StartedAt, &session.CompletedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *pomodoroRepository) GetSessionByID(sessionID int64) (*models.PomodoroSession, error) {
	session := &models.PomodoroSession{}
	err := r.db.QueryRow(`
		SELECT id, user_id, task_id, duration, is_completed, started_at, completed_at 
		FROM pomodoro_sessions 
		WHERE id = $1`, sessionID).Scan(&session.ID, &session.UserID, &session.TaskID,
		&session.Duration, &session.IsCompleted, &session.StartedAt, &session.CompletedAt)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *pomodoroRepository) UpdateSession(session *models.PomodoroSession) (*models.PomodoroSession, error) {
	var updatedSession models.PomodoroSession
	err := r.db.QueryRow(`
		UPDATE pomodoro_sessions 
		SET task_id = $2, duration = $3, is_completed = $4, completed_at = $5
		WHERE id = $1 
		RETURNING id, user_id, task_id, duration, is_completed, started_at, completed_at`,
		session.ID, session.TaskID, session.Duration, session.IsCompleted, session.CompletedAt,
	).Scan(&updatedSession.ID, &updatedSession.UserID, &updatedSession.TaskID,
		&updatedSession.Duration, &updatedSession.IsCompleted, &updatedSession.StartedAt, &updatedSession.CompletedAt)

	if err != nil {
		return nil, err
	}

	return &updatedSession, nil
}

func (r *pomodoroRepository) CompleteSession(sessionID int64) error {
	_, err := r.db.Exec(`
		UPDATE pomodoro_sessions 
		SET is_completed = true, completed_at = CURRENT_TIMESTAMP 
		WHERE id = $1`, sessionID)
	return err
}

func (r *pomodoroRepository) DeleteSession(sessionID int64) error {
	_, err := r.db.Exec("DELETE FROM pomodoro_sessions WHERE id = $1", sessionID)
	return err
}

func (r *pomodoroRepository) GetSessionStats(userID int64) (*models.PomodoroStats, error) {
	stats := &models.PomodoroStats{}
	err := r.db.QueryRow(`
		SELECT 
			COUNT(*) as total_sessions,
			COUNT(CASE WHEN is_completed = true THEN 1 END) as completed_sessions,
			COALESCE(SUM(CASE WHEN is_completed = true THEN duration ELSE 0 END), 0) as total_minutes
		FROM pomodoro_sessions 
		WHERE user_id = $1`, userID).Scan(&stats.TotalSessions, &stats.CompletedSessions, &stats.TotalMinutes)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Goal Repository
type GoalRepository interface {
	CreateGoal(goal *models.Goal) (*models.Goal, error)
	GetGoalsByUserID(userID int64) ([]*models.Goal, error)
	GetGoalByID(goalID int64) (*models.Goal, error)
	UpdateGoal(goal *models.Goal) (*models.Goal, error)
	DeleteGoal(goalID int64) error
	UpdateGoalProgress(goalID int64, progress int32) error
}

type goalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) GoalRepository {
	return &goalRepository{db: db}
}

func (r *goalRepository) CreateGoal(goal *models.Goal) (*models.Goal, error) {
	var createdGoal models.Goal
	err := r.db.QueryRow(`
		INSERT INTO goals (user_id, title, description, category, target_value, unit, due_date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, user_id, title, description, category, target_value, current_value, unit, due_date, is_completed, created_at`,
		goal.UserID, goal.Title, goal.Description, goal.Category, goal.TargetValue, goal.Unit, goal.DueDate,
	).Scan(&createdGoal.ID, &createdGoal.UserID, &createdGoal.Title, &createdGoal.Description,
		&createdGoal.Category, &createdGoal.TargetValue, &createdGoal.CurrentValue, &createdGoal.Unit,
		&createdGoal.DueDate, &createdGoal.IsCompleted, &createdGoal.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &createdGoal, nil
}

func (r *goalRepository) GetGoalsByUserID(userID int64) ([]*models.Goal, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, title, description, category, target_value, current_value, unit, due_date, is_completed, created_at 
		FROM goals 
		WHERE user_id = $1 
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []*models.Goal
	for rows.Next() {
		goal := &models.Goal{}
		err := rows.Scan(&goal.ID, &goal.UserID, &goal.Title, &goal.Description, &goal.Category,
			&goal.TargetValue, &goal.CurrentValue, &goal.Unit, &goal.DueDate, &goal.IsCompleted, &goal.CreatedAt)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}

	return goals, nil
}

func (r *goalRepository) GetGoalByID(goalID int64) (*models.Goal, error) {
	goal := &models.Goal{}
	err := r.db.QueryRow(`
		SELECT id, user_id, title, description, category, target_value, current_value, unit, due_date, is_completed, created_at 
		FROM goals 
		WHERE id = $1`, goalID).Scan(&goal.ID, &goal.UserID, &goal.Title, &goal.Description,
		&goal.Category, &goal.TargetValue, &goal.CurrentValue, &goal.Unit, &goal.DueDate, &goal.IsCompleted, &goal.CreatedAt)

	if err != nil {
		return nil, err
	}

	return goal, nil
}

func (r *goalRepository) UpdateGoal(goal *models.Goal) (*models.Goal, error) {
	var updatedGoal models.Goal
	err := r.db.QueryRow(`
		UPDATE goals 
		SET title = $2, description = $3, category = $4, target_value = $5, current_value = $6, 
			unit = $7, due_date = $8, is_completed = $9
		WHERE id = $1 
		RETURNING id, user_id, title, description, category, target_value, current_value, unit, due_date, is_completed, created_at`,
		goal.ID, goal.Title, goal.Description, goal.Category, goal.TargetValue, goal.CurrentValue,
		goal.Unit, goal.DueDate, goal.IsCompleted,
	).Scan(&updatedGoal.ID, &updatedGoal.UserID, &updatedGoal.Title, &updatedGoal.Description,
		&updatedGoal.Category, &updatedGoal.TargetValue, &updatedGoal.CurrentValue, &updatedGoal.Unit,
		&updatedGoal.DueDate, &updatedGoal.IsCompleted, &updatedGoal.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &updatedGoal, nil
}

func (r *goalRepository) DeleteGoal(goalID int64) error {
	_, err := r.db.Exec("DELETE FROM goals WHERE id = $1", goalID)
	return err
}

func (r *goalRepository) UpdateGoalProgress(goalID int64, progress int32) error {
	_, err := r.db.Exec(`
		UPDATE goals 
		SET current_value = $2, is_completed = (current_value >= target_value) 
		WHERE id = $1`, goalID, progress)
	return err
}
