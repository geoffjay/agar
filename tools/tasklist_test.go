package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestTaskListTool_Name(t *testing.T) {
	tool := NewTaskListTool()
	if tool.Name() != "tasklist" {
		t.Errorf("Expected name 'tasklist', got '%s'", tool.Name())
	}
}

func TestTaskListTool_Validate(t *testing.T) {
	tool := NewTaskListTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid create action",
			params:  `{"action": "create", "title": "Test List"}`,
			wantErr: false,
		},
		{
			name:    "valid update action",
			params:  `{"action": "update", "list_id": "123", "task_id": "456", "status": "completed"}`,
			wantErr: false,
		},
		{
			name:    "valid list action",
			params:  `{"action": "list"}`,
			wantErr: false,
		},
		{
			name:    "missing action",
			params:  `{"title": "Test"}`,
			wantErr: true,
		},
		{
			name:    "invalid action",
			params:  `{"action": "invalid"}`,
			wantErr: true,
		},
		{
			name:    "invalid priority",
			params:  `{"action": "create", "priority": "invalid"}`,
			wantErr: true,
		},
		{
			name:    "invalid status",
			params:  `{"action": "update", "status": "invalid"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.Validate(json.RawMessage(tt.params))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskListTool_CreateList(t *testing.T) {
	tool := NewTaskListTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "create",
		"title":  "My Task List",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	taskResult, ok := result.(*TaskListResult)
	if !ok {
		t.Fatal("Result is not a TaskListResult")
	}

	if taskResult.Action != "create" {
		t.Errorf("Expected action 'create', got '%s'", taskResult.Action)
	}

	if taskResult.ListID == "" {
		t.Error("Expected list_id to be set")
	}
}

func TestTaskListTool_CreateTask(t *testing.T) {
	tool := NewTaskListTool()
	ctx := context.Background()

	// First create a list
	createListParams := map[string]interface{}{
		"action": "create",
		"title":  "Test List",
	}
	paramsJSON, _ := json.Marshal(createListParams)

	listResult, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	listID := listResult.(*TaskListResult).ListID

	// Now create a task in the list
	createTaskParams := map[string]interface{}{
		"action":      "create",
		"list_id":     listID,
		"title":       "Test Task",
		"description": "Test task description",
		"priority":    "high",
	}
	paramsJSON, _ = json.Marshal(createTaskParams)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	taskResult := result.(*TaskListResult)

	if taskResult.TaskID == "" {
		t.Error("Expected task_id to be set")
	}

	if taskResult.Task == nil {
		t.Fatal("Expected task to be returned")
	}

	if taskResult.Task.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", taskResult.Task.Title)
	}

	if taskResult.Task.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", taskResult.Task.Priority)
	}

	if taskResult.Task.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", taskResult.Task.Status)
	}
}

func TestTaskListTool_UpdateTask(t *testing.T) {
	tool := NewTaskListTool()
	ctx := context.Background()

	// Create list and task
	createListParams := map[string]interface{}{
		"action": "create",
	}
	paramsJSON, _ := json.Marshal(createListParams)
	listResult, _ := tool.Execute(ctx, paramsJSON)
	listID := listResult.(*TaskListResult).ListID

	createTaskParams := map[string]interface{}{
		"action":  "create",
		"list_id": listID,
		"title":   "Original Title",
	}
	paramsJSON, _ = json.Marshal(createTaskParams)
	taskResult, _ := tool.Execute(ctx, paramsJSON)
	taskID := taskResult.(*TaskListResult).TaskID

	// Update the task
	updateParams := map[string]interface{}{
		"action":  "update",
		"list_id": listID,
		"task_id": taskID,
		"title":   "Updated Title",
		"status":  "completed",
	}
	paramsJSON, _ = json.Marshal(updateParams)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	updateResult := result.(*TaskListResult)

	if updateResult.Task.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updateResult.Task.Title)
	}

	if updateResult.Task.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", updateResult.Task.Status)
	}

	if updateResult.Task.CompletedAt == 0 {
		t.Error("Expected completed_at to be set when status is completed")
	}
}

func TestTaskListTool_ListLists(t *testing.T) {
	tool := NewTaskListTool()
	ctx := context.Background()

	// Create multiple lists
	for i := 0; i < 3; i++ {
		params := map[string]interface{}{
			"action": "create",
			"title":  "List " + string(rune('A'+i)),
		}
		paramsJSON, _ := json.Marshal(params)
		tool.Execute(ctx, paramsJSON)
	}

	// List all lists
	listParams := map[string]interface{}{
		"action": "list",
	}
	paramsJSON, _ := json.Marshal(listParams)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	listResult := result.(*TaskListResult)

	if len(listResult.Lists) != 3 {
		t.Errorf("Expected 3 lists, got %d", len(listResult.Lists))
	}
}

func TestTaskListTool_ListTasks(t *testing.T) {
	tool := NewTaskListTool()
	ctx := context.Background()

	// Create list
	createListParams := map[string]interface{}{
		"action": "create",
	}
	paramsJSON, _ := json.Marshal(createListParams)
	listResult, _ := tool.Execute(ctx, paramsJSON)
	listID := listResult.(*TaskListResult).ListID

	// Create multiple tasks
	for i := 0; i < 3; i++ {
		params := map[string]interface{}{
			"action":  "create",
			"list_id": listID,
			"title":   "Task " + string(rune('1'+i)),
		}
		paramsJSON, _ := json.Marshal(params)
		tool.Execute(ctx, paramsJSON)
	}

	// List tasks in the list
	listParams := map[string]interface{}{
		"action":  "list",
		"list_id": listID,
	}
	paramsJSON, _ = json.Marshal(listParams)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	taskListResult := result.(*TaskListResult)

	if len(taskListResult.Tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(taskListResult.Tasks))
	}
}

func TestTaskListTool_GetTask(t *testing.T) {
	tool := NewTaskListTool()
	ctx := context.Background()

	// Create list and task
	createListParams := map[string]interface{}{
		"action": "create",
	}
	paramsJSON, _ := json.Marshal(createListParams)
	listResult, _ := tool.Execute(ctx, paramsJSON)
	listID := listResult.(*TaskListResult).ListID

	createTaskParams := map[string]interface{}{
		"action":  "create",
		"list_id": listID,
		"title":   "Test Task",
	}
	paramsJSON, _ = json.Marshal(createTaskParams)
	taskResult, _ := tool.Execute(ctx, paramsJSON)
	taskID := taskResult.(*TaskListResult).TaskID

	// Get the task
	getParams := map[string]interface{}{
		"action":  "get",
		"list_id": listID,
		"task_id": taskID,
	}
	paramsJSON, _ = json.Marshal(getParams)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	getResult := result.(*TaskListResult)

	if getResult.Task == nil {
		t.Fatal("Expected task to be returned")
	}

	if getResult.Task.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", getResult.Task.Title)
	}
}

func TestTaskListTool_DeleteTask(t *testing.T) {
	tool := NewTaskListTool()
	ctx := context.Background()

	// Create list and task
	createListParams := map[string]interface{}{
		"action": "create",
	}
	paramsJSON, _ := json.Marshal(createListParams)
	listResult, _ := tool.Execute(ctx, paramsJSON)
	listID := listResult.(*TaskListResult).ListID

	createTaskParams := map[string]interface{}{
		"action":  "create",
		"list_id": listID,
		"title":   "Task to Delete",
	}
	paramsJSON, _ = json.Marshal(createTaskParams)
	taskResult, _ := tool.Execute(ctx, paramsJSON)
	taskID := taskResult.(*TaskListResult).TaskID

	// Delete the task
	deleteParams := map[string]interface{}{
		"action":  "delete",
		"list_id": listID,
		"task_id": taskID,
	}
	paramsJSON, _ = json.Marshal(deleteParams)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	deleteResult := result.(*TaskListResult)

	if deleteResult.Action != "delete" {
		t.Errorf("Expected action 'delete', got '%s'", deleteResult.Action)
	}

	// Verify task is deleted
	getParams := map[string]interface{}{
		"action":  "get",
		"list_id": listID,
		"task_id": taskID,
	}
	paramsJSON, _ = json.Marshal(getParams)

	_, err = tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error when getting deleted task")
	}
}

func TestTaskListTool_DeleteList(t *testing.T) {
	tool := NewTaskListTool()
	ctx := context.Background()

	// Create list
	createListParams := map[string]interface{}{
		"action": "create",
		"title":  "List to Delete",
	}
	paramsJSON, _ := json.Marshal(createListParams)
	listResult, _ := tool.Execute(ctx, paramsJSON)
	listID := listResult.(*TaskListResult).ListID

	// Delete the list
	deleteParams := map[string]interface{}{
		"action":  "delete",
		"list_id": listID,
	}
	paramsJSON, _ = json.Marshal(deleteParams)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	deleteResult := result.(*TaskListResult)

	if deleteResult.Action != "delete" {
		t.Errorf("Expected action 'delete', got '%s'", deleteResult.Action)
	}

	// Verify list is deleted
	listParams := map[string]interface{}{
		"action":  "list",
		"list_id": listID,
	}
	paramsJSON, _ = json.Marshal(listParams)

	_, err = tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error when listing deleted list")
	}
}

func TestTaskListTool_HierarchicalTasks(t *testing.T) {
	tool := NewTaskListTool()
	ctx := context.Background()

	// Create list
	createListParams := map[string]interface{}{
		"action": "create",
	}
	paramsJSON, _ := json.Marshal(createListParams)
	listResult, _ := tool.Execute(ctx, paramsJSON)
	listID := listResult.(*TaskListResult).ListID

	// Create parent task
	createParentParams := map[string]interface{}{
		"action":  "create",
		"list_id": listID,
		"title":   "Parent Task",
	}
	paramsJSON, _ = json.Marshal(createParentParams)
	parentResult, _ := tool.Execute(ctx, paramsJSON)
	parentID := parentResult.(*TaskListResult).TaskID

	// Create child task
	createChildParams := map[string]interface{}{
		"action":    "create",
		"list_id":   listID,
		"title":     "Child Task",
		"parent_id": parentID,
	}
	paramsJSON, _ = json.Marshal(createChildParams)

	childResult, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	childTask := childResult.(*TaskListResult).Task

	if childTask.ParentID != parentID {
		t.Errorf("Expected parent_id '%s', got '%s'", parentID, childTask.ParentID)
	}

	// Verify parent has child
	getParentParams := map[string]interface{}{
		"action":  "get",
		"list_id": listID,
		"task_id": parentID,
	}
	paramsJSON, _ = json.Marshal(getParentParams)

	parentResultGet, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Failed to get parent: %v", err)
	}

	parentTask := parentResultGet.(*TaskListResult).Task

	if len(parentTask.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(parentTask.Children))
	}
}
