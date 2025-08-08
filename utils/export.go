package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"todo-backend/models"
)

// Export functions
func ExportTasksAsJSON(tasks []*models.Task) ([]byte, error) {
	return json.MarshalIndent(tasks, "", "  ")
}

func ExportTasksAsCSV(tasks []*models.Task) ([]byte, error) {
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)

	// Write header
	header := []string{
		"ID", "TaskName", "Description", "Category", "Priority",
		"DueDate", "IsCompleted", "IsRecurring", "RecurringFrequency", "CreatedAt",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write data
	for _, task := range tasks {
		record := []string{
			strconv.FormatInt(task.ID, 10),
			task.TaskName,
			stringValueOrEmpty(task.Description),
			stringValueOrEmpty(task.Category),
			strconv.FormatInt(int64(task.Priority), 10),
			timeValueOrEmpty(task.DueDate),
			strconv.FormatBool(task.IsCompleted),
			strconv.FormatBool(task.IsRecurring),
			stringValueOrEmpty(task.RecurringFrequency),
			task.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return []byte(csvData.String()), nil
}

// Import functions
func ImportTasksFromJSON(data []byte) ([]*models.Task, error) {
	var tasks []*models.Task
	err := json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return tasks, nil
}

func ImportTasksFromCSV(data []byte) ([]*models.Task, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 { // At least header + 1 data row
		return nil, fmt.Errorf("CSV file must contain at least one data row")
	}

	var tasks []*models.Task

	// Skip header row (index 0)
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 10 {
			continue // Skip incomplete rows
		}

		task := &models.Task{}

		// Parse TaskName (required)
		if record[1] == "" {
			continue // Skip rows without task name
		}
		task.TaskName = record[1]

		// Parse Description
		if record[2] != "" {
			desc := record[2]
			task.Description = &desc
		}

		// Parse Category
		if record[3] != "" {
			cat := record[3]
			task.Category = &cat
		}

		// Parse Priority
		if priority, err := strconv.ParseInt(record[4], 10, 16); err == nil {
			task.Priority = int16(priority)
		}

		// Parse DueDate
		if record[5] != "" {
			if dueDate, err := time.Parse(time.RFC3339, record[5]); err == nil {
				customTime := &models.CustomTime{Time: dueDate}
				task.DueDate = customTime
			}
		}

		// Parse IsCompleted
		if isCompleted, err := strconv.ParseBool(record[6]); err == nil {
			task.IsCompleted = isCompleted
		}

		// Parse IsRecurring
		if isRecurring, err := strconv.ParseBool(record[7]); err == nil {
			task.IsRecurring = isRecurring
		}

		// Parse RecurringFrequency
		if record[8] != "" {
			freq := record[8]
			task.RecurringFrequency = &freq
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Helper functions
func stringValueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func timeValueOrEmpty(t *models.CustomTime) string {
	if t == nil {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}

// Habit Export/Import functions
func ExportHabitsAsJSON(habits []*models.Habit) ([]byte, error) {
	return json.MarshalIndent(habits, "", "  ")
}

func ExportHabitsAsCSV(habits []*models.Habit) ([]byte, error) {
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)

	// Write header
	header := []string{
		"ID", "Name", "Type", "TargetValue", "IsAchieved", "LastTrackedDate", "CreatedAt",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write data
	for _, habit := range habits {
		record := []string{
			strconv.FormatInt(habit.ID, 10),
			habit.Name,
			habit.Type,
			stringValueOrEmpty(habit.TargetValue),
			strconv.FormatBool(habit.IsAchieved),
			timeValueOrEmpty(habit.LastTrackedDate),
			habit.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return []byte(csvData.String()), nil
}

func ImportHabitsFromJSON(data []byte) ([]*models.Habit, error) {
	var habits []*models.Habit
	err := json.Unmarshal(data, &habits)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return habits, nil
}

func ImportHabitsFromCSV(data []byte) ([]*models.Habit, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 { // At least header + 1 data row
		return nil, fmt.Errorf("CSV file must contain at least one data row")
	}

	var habits []*models.Habit

	// Skip header row (index 0)
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 7 {
			continue // Skip incomplete rows
		}

		habit := &models.Habit{}

		// Parse Name (required)
		if record[1] == "" {
			continue // Skip rows without habit name
		}
		habit.Name = record[1]

		// Parse Type (required)
		if record[2] == "" {
			continue // Skip rows without habit type
		}
		habit.Type = record[2]

		// Parse TargetValue
		if record[3] != "" {
			targetValue := record[3]
			habit.TargetValue = &targetValue
		}

		// Parse IsAchieved
		if isAchieved, err := strconv.ParseBool(record[4]); err == nil {
			habit.IsAchieved = isAchieved
		}

		// Parse LastTrackedDate
		if record[5] != "" {
			if lastTracked, err := time.Parse(time.RFC3339, record[5]); err == nil {
				customTime := &models.CustomTime{Time: lastTracked}
				habit.LastTrackedDate = customTime
			}
		}

		habits = append(habits, habit)
	}

	return habits, nil
}
