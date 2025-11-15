package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TaskListTool implements task list management functionality
type TaskListTool struct {
	lists map[string]*TaskList
	mu    sync.RWMutex
}

// TaskListParams defines the parameters for the TaskList tool
type TaskListParams struct {
	Action      string `json:"action"`       // "create", "update", "list", "delete", "get"
	ListID      string `json:"list_id,omitempty"`
	TaskID      string `json:"task_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority,omitempty"` // "high", "medium", "low"
	Status      string `json:"status,omitempty"`   // "pending", "in_progress", "completed", "cancelled"
	ParentID    string `json:"parent_id,omitempty"`
}

// TaskListResult represents the result of a task list operation
type TaskListResult struct {
	Action  string      `json:"action"`
	ListID  string      `json:"list_id,omitempty"`
	TaskID  string      `json:"task_id,omitempty"`
	Task    *Task       `json:"task,omitempty"`
	Tasks   []*Task     `json:"tasks,omitempty"`
	Lists   []ListInfo  `json:"lists,omitempty"`
	Message string      `json:"message,omitempty"`
}

// TaskList represents a collection of tasks
type TaskList struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Tasks     map[string]*Task `json:"tasks"`
	CreatedAt int64            `json:"created_at"`
	UpdatedAt int64            `json:"updated_at"`
}

// Task represents a single task
type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Status      string   `json:"status"`
	ParentID    string   `json:"parent_id,omitempty"`
	Children    []string `json:"children,omitempty"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
	CompletedAt int64    `json:"completed_at,omitempty"`
}

// ListInfo represents basic information about a task list
type ListInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	TaskCount int    `json:"task_count"`
	CreatedAt int64  `json:"created_at"`
}

// NewTaskListTool creates a new TaskList tool instance
func NewTaskListTool() *TaskListTool {
	return &TaskListTool{
		lists: make(map[string]*TaskList),
	}
}

// Name returns the tool's name
func (t *TaskListTool) Name() string {
	return "tasklist"
}

// Description returns the tool's description
func (t *TaskListTool) Description() string {
	return "Create and manage hierarchical task lists with support for priorities, status tracking, and task dependencies"
}

// Schema returns the JSON schema for the tool's parameters
func (t *TaskListTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform: 'create', 'update', 'list', 'delete', 'get'",
				"enum":        []string{"create", "update", "list", "delete", "get"},
			},
			"list_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the task list",
			},
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the task",
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Title of the task or list",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description of the task",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"description": "Task priority: 'high', 'medium', or 'low'",
				"enum":        []string{"high", "medium", "low"},
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Task status: 'pending', 'in_progress', 'completed', or 'cancelled'",
				"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
			},
			"parent_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the parent task for hierarchical tasks",
			},
		},
		"required": []string{"action"},
	}
}

// Validate checks if the parameters are valid
func (t *TaskListTool) Validate(params json.RawMessage) error {
	var p TaskListParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.Action == "" {
		return fmt.Errorf("action is required")
	}

	validActions := map[string]bool{"create": true, "update": true, "list": true, "delete": true, "get": true}
	if !validActions[p.Action] {
		return fmt.Errorf("invalid action: %s", p.Action)
	}

	if p.Priority != "" && p.Priority != "high" && p.Priority != "medium" && p.Priority != "low" {
		return fmt.Errorf("priority must be 'high', 'medium', or 'low'")
	}

	if p.Status != "" && p.Status != "pending" && p.Status != "in_progress" && p.Status != "completed" && p.Status != "cancelled" {
		return fmt.Errorf("status must be 'pending', 'in_progress', 'completed', or 'cancelled'")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *TaskListTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p TaskListParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	switch p.Action {
	case "create":
		return t.createTask(p)
	case "update":
		return t.updateTask(p)
	case "list":
		return t.listTasks(p)
	case "delete":
		return t.deleteTask(p)
	case "get":
		return t.getTask(p)
	default:
		return nil, fmt.Errorf("unknown action: %s", p.Action)
	}
}

// createTask creates a new task or task list
func (t *TaskListTool) createTask(p TaskListParams) (*TaskListResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If no list_id, create a new list
	if p.ListID == "" {
		listID := uuid.New().String()
		title := p.Title
		if title == "" {
			title = "New Task List"
		}

		t.lists[listID] = &TaskList{
			ID:        listID,
			Title:     title,
			Tasks:     make(map[string]*Task),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}

		return &TaskListResult{
			Action:  "create",
			ListID:  listID,
			Message: "Task list created successfully",
		}, nil
	}

	// Create task in existing list
	list, exists := t.lists[p.ListID]
	if !exists {
		return nil, fmt.Errorf("task list not found: %s", p.ListID)
	}

	taskID := uuid.New().String()
	priority := p.Priority
	if priority == "" {
		priority = "medium"
	}

	task := &Task{
		ID:          taskID,
		Title:       p.Title,
		Description: p.Description,
		Priority:    priority,
		Status:      "pending",
		ParentID:    p.ParentID,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	list.Tasks[taskID] = task
	list.UpdatedAt = time.Now().Unix()

	// Update parent's children list
	if p.ParentID != "" {
		if parent, ok := list.Tasks[p.ParentID]; ok {
			parent.Children = append(parent.Children, taskID)
		}
	}

	return &TaskListResult{
		Action:  "create",
		ListID:  p.ListID,
		TaskID:  taskID,
		Task:    task,
		Message: "Task created successfully",
	}, nil
}

// updateTask updates an existing task
func (t *TaskListTool) updateTask(p TaskListParams) (*TaskListResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if p.ListID == "" || p.TaskID == "" {
		return nil, fmt.Errorf("list_id and task_id are required for update")
	}

	list, exists := t.lists[p.ListID]
	if !exists {
		return nil, fmt.Errorf("task list not found: %s", p.ListID)
	}

	task, exists := list.Tasks[p.TaskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", p.TaskID)
	}

	// Update fields
	if p.Title != "" {
		task.Title = p.Title
	}
	if p.Description != "" {
		task.Description = p.Description
	}
	if p.Priority != "" {
		task.Priority = p.Priority
	}
	if p.Status != "" {
		task.Status = p.Status
		if p.Status == "completed" {
			task.CompletedAt = time.Now().Unix()
		}
	}

	task.UpdatedAt = time.Now().Unix()
	list.UpdatedAt = time.Now().Unix()

	return &TaskListResult{
		Action:  "update",
		ListID:  p.ListID,
		TaskID:  p.TaskID,
		Task:    task,
		Message: "Task updated successfully",
	}, nil
}

// listTasks lists all tasks in a list or all lists
func (t *TaskListTool) listTasks(p TaskListParams) (*TaskListResult, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// List all task lists
	if p.ListID == "" {
		var lists []ListInfo
		for _, list := range t.lists {
			lists = append(lists, ListInfo{
				ID:        list.ID,
				Title:     list.Title,
				TaskCount: len(list.Tasks),
				CreatedAt: list.CreatedAt,
			})
		}

		return &TaskListResult{
			Action: "list",
			Lists:  lists,
		}, nil
	}

	// List tasks in a specific list
	list, exists := t.lists[p.ListID]
	if !exists {
		return nil, fmt.Errorf("task list not found: %s", p.ListID)
	}

	var tasks []*Task
	for _, task := range list.Tasks {
		tasks = append(tasks, task)
	}

	return &TaskListResult{
		Action: "list",
		ListID: p.ListID,
		Tasks:  tasks,
	}, nil
}

// deleteTask deletes a task or task list
func (t *TaskListTool) deleteTask(p TaskListParams) (*TaskListResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Delete task list
	if p.TaskID == "" {
		if p.ListID == "" {
			return nil, fmt.Errorf("list_id is required for deletion")
		}

		if _, exists := t.lists[p.ListID]; !exists {
			return nil, fmt.Errorf("task list not found: %s", p.ListID)
		}

		delete(t.lists, p.ListID)

		return &TaskListResult{
			Action:  "delete",
			ListID:  p.ListID,
			Message: "Task list deleted successfully",
		}, nil
	}

	// Delete task
	if p.ListID == "" {
		return nil, fmt.Errorf("list_id is required for task deletion")
	}

	list, exists := t.lists[p.ListID]
	if !exists {
		return nil, fmt.Errorf("task list not found: %s", p.ListID)
	}

	if _, exists := list.Tasks[p.TaskID]; !exists {
		return nil, fmt.Errorf("task not found: %s", p.TaskID)
	}

	delete(list.Tasks, p.TaskID)
	list.UpdatedAt = time.Now().Unix()

	return &TaskListResult{
		Action:  "delete",
		ListID:  p.ListID,
		TaskID:  p.TaskID,
		Message: "Task deleted successfully",
	}, nil
}

// getTask retrieves a specific task
func (t *TaskListTool) getTask(p TaskListParams) (*TaskListResult, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if p.ListID == "" || p.TaskID == "" {
		return nil, fmt.Errorf("list_id and task_id are required")
	}

	list, exists := t.lists[p.ListID]
	if !exists {
		return nil, fmt.Errorf("task list not found: %s", p.ListID)
	}

	task, exists := list.Tasks[p.TaskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", p.TaskID)
	}

	return &TaskListResult{
		Action: "get",
		ListID: p.ListID,
		TaskID: p.TaskID,
		Task:   task,
	}, nil
}
