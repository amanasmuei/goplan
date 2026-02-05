package tools

import (
	"context"
	"testing"

	"github.com/goplan/goplan/internal/domain/plan"
	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/domain/task"
	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/repository"
)

// MockTaskRepository implements repository.TaskRepository for testing.
type MockTaskRepository struct {
	tasks map[string]*task.Task
}

func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks: make(map[string]*task.Task),
	}
}

func (m *MockTaskRepository) Create(ctx context.Context, t *task.Task) error {
	m.tasks[t.ID] = t
	return nil
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id string) (*task.Task, error) {
	if t, ok := m.tasks[id]; ok {
		return t, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockTaskRepository) GetByIDWithDetails(ctx context.Context, id string) (*task.Task, error) {
	return m.GetByID(ctx, id)
}

func (m *MockTaskRepository) Update(ctx context.Context, id string, input *task.UpdateTaskInput) (*task.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	if input.Title != nil {
		t.Title = *input.Title
	}
	if input.Description != nil {
		t.Description = input.Description
	}
	if input.Status != nil {
		t.Status = *input.Status
	}
	if input.Priority != nil {
		t.Priority = *input.Priority
	}
	if input.AssigneeID != nil {
		t.AssigneeID = input.AssigneeID
	}
	if input.DueDate != nil {
		t.DueDate = input.DueDate
	}
	return t, nil
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	delete(m.tasks, id)
	return nil
}

func (m *MockTaskRepository) List(ctx context.Context, filter repository.TaskFilterOptions, sort repository.TaskSortOptions, pagination repository.Pagination) (*repository.PaginatedResult[task.Task], error) {
	var items []task.Task
	for _, t := range m.tasks {
		if filter.PlanID != nil && t.PlanID != *filter.PlanID {
			continue
		}
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		items = append(items, *t)
	}
	return repository.NewPaginatedResult(items, int64(len(items)), pagination), nil
}

func (m *MockTaskRepository) ListByPlan(ctx context.Context, planID string, pagination repository.Pagination) (*repository.PaginatedResult[task.Task], error) {
	filter := repository.TaskFilterOptions{PlanID: &planID}
	return m.List(ctx, filter, repository.DefaultTaskSort(), pagination)
}

func (m *MockTaskRepository) GetSubtasks(ctx context.Context, parentID string) ([]*task.Task, error) {
	var subtasks []*task.Task
	for _, t := range m.tasks {
		if t.ParentID != nil && *t.ParentID == parentID {
			subtasks = append(subtasks, t)
		}
	}
	return subtasks, nil
}

func (m *MockTaskRepository) Search(ctx context.Context, planID, query string, pagination repository.Pagination) (*repository.PaginatedResult[task.Task], error) {
	var items []task.Task
	for _, t := range m.tasks {
		if t.PlanID == planID {
			items = append(items, *t)
		}
	}
	return repository.NewPaginatedResult(items, int64(len(items)), pagination), nil
}

func (m *MockTaskRepository) Move(ctx context.Context, id string, status string, position int) error {
	t, ok := m.tasks[id]
	if !ok {
		return shared.ErrNotFound
	}
	t.Status = status
	t.Position = position
	return nil
}

func (m *MockTaskRepository) BatchUpdatePositions(ctx context.Context, updates map[string]int) error {
	for id, pos := range updates {
		if t, ok := m.tasks[id]; ok {
			t.Position = pos
		}
	}
	return nil
}

func (m *MockTaskRepository) AddDependency(ctx context.Context, dep *task.TaskDependency) error {
	return nil
}

func (m *MockTaskRepository) RemoveDependency(ctx context.Context, taskID, dependsOnID string) error {
	return nil
}

func (m *MockTaskRepository) GetDependencies(ctx context.Context, taskID string) ([]*task.TaskDependency, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetDependents(ctx context.Context, taskID string) ([]*task.TaskDependency, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetTasksByStatus(ctx context.Context, planID string) (map[string][]*task.Task, error) {
	result := make(map[string][]*task.Task)
	for _, t := range m.tasks {
		if t.PlanID == planID {
			result[t.Status] = append(result[t.Status], t)
		}
	}
	return result, nil
}

func (m *MockTaskRepository) CountByStatus(ctx context.Context, planID string) (map[string]int, error) {
	result := make(map[string]int)
	for _, t := range m.tasks {
		if t.PlanID == planID {
			result[t.Status]++
		}
	}
	return result, nil
}

func (m *MockTaskRepository) GetOverdue(ctx context.Context, planID string) ([]*task.Task, error) {
	return nil, nil
}

// MockPlanRepository implements repository.PlanRepository for testing.
type MockPlanRepository struct {
	plans map[string]*plan.Plan
}

func NewMockPlanRepository() *MockPlanRepository {
	return &MockPlanRepository{
		plans: make(map[string]*plan.Plan),
	}
}

func (m *MockPlanRepository) Create(ctx context.Context, p *plan.Plan) error {
	m.plans[p.ID] = p
	return nil
}

func (m *MockPlanRepository) GetByID(ctx context.Context, id string) (*plan.Plan, error) {
	if p, ok := m.plans[id]; ok {
		return p, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockPlanRepository) Update(ctx context.Context, id string, input *plan.UpdatePlanInput) (*plan.Plan, error) {
	p, ok := m.plans[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	if input.Name != nil {
		p.Name = *input.Name
	}
	if input.Description != nil {
		p.Description = input.Description
	}
	if input.Status != nil {
		p.Status = *input.Status
	}
	return p, nil
}

func (m *MockPlanRepository) Delete(ctx context.Context, id string) error {
	delete(m.plans, id)
	return nil
}

func (m *MockPlanRepository) List(ctx context.Context, filter repository.PlanFilterOptions, sort repository.PlanSortOptions, pagination repository.Pagination) (*repository.PaginatedResult[plan.Plan], error) {
	var items []plan.Plan
	for _, p := range m.plans {
		if filter.WorkspaceID != nil && p.WorkspaceID != *filter.WorkspaceID {
			continue
		}
		items = append(items, *p)
	}
	return repository.NewPaginatedResult(items, int64(len(items)), pagination), nil
}

func (m *MockPlanRepository) GetByWorkspace(ctx context.Context, workspaceID string, pagination repository.Pagination) (*repository.PaginatedResult[plan.Plan], error) {
	filter := repository.PlanFilterOptions{WorkspaceID: &workspaceID}
	return m.List(ctx, filter, repository.DefaultPlanSort(), pagination)
}

func (m *MockPlanRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	if p, ok := m.plans[id]; ok {
		p.Status = status
		return nil
	}
	return shared.ErrNotFound
}

func (m *MockPlanRepository) UpdateTags(ctx context.Context, id string, tags []string) error {
	if p, ok := m.plans[id]; ok {
		p.Tags = tags
		return nil
	}
	return shared.ErrNotFound
}

// MockCommentRepository implements repository.CommentRepository for testing.
type MockCommentRepository struct {
	comments map[string]*task.Comment
}

func NewMockCommentRepository() *MockCommentRepository {
	return &MockCommentRepository{
		comments: make(map[string]*task.Comment),
	}
}

func (m *MockCommentRepository) Create(ctx context.Context, c *task.Comment) error {
	m.comments[c.ID] = c
	return nil
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id string) (*task.Comment, error) {
	if c, ok := m.comments[id]; ok {
		return c, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockCommentRepository) Update(ctx context.Context, id string, content string, mentions []string) (*task.Comment, error) {
	c, ok := m.comments[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	c.Content = content
	c.Mentions = mentions
	return c, nil
}

func (m *MockCommentRepository) Delete(ctx context.Context, id string) error {
	delete(m.comments, id)
	return nil
}

func (m *MockCommentRepository) ListByTask(ctx context.Context, taskID string) ([]*task.Comment, error) {
	var comments []*task.Comment
	for _, c := range m.comments {
		if c.TaskID == taskID {
			comments = append(comments, c)
		}
	}
	return comments, nil
}

func (m *MockCommentRepository) ListByTaskPaginated(ctx context.Context, taskID string, pagination repository.Pagination) (*repository.PaginatedResult[task.Comment], error) {
	comments, _ := m.ListByTask(ctx, taskID)
	items := make([]task.Comment, len(comments))
	for i, c := range comments {
		items[i] = *c
	}
	return repository.NewPaginatedResult(items, int64(len(items)), pagination), nil
}

func (m *MockCommentRepository) CountByTask(ctx context.Context, taskID string) (int, error) {
	count := 0
	for _, c := range m.comments {
		if c.TaskID == taskID {
			count++
		}
	}
	return count, nil
}

func TestCreateTaskTool(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	planRepo := NewMockPlanRepository()
	tool := NewCreateTaskTool(taskRepo, planRepo)

	// Create test plan
	p := plan.NewPlan("plan-1", "ws-1", "Test Plan", "user-1", shared.PlanDomainSoftware, nil)
	_ = planRepo.Create(context.Background(), p)

	t.Run("creates task successfully", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1", WorkspaceID: "ws-1"}
		args := map[string]interface{}{
			"planId": "plan-1",
			"title":  "New Task",
		}

		result, err := tool.Execute(context.Background(), execCtx, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		taskResp := data["task"].(*task.TaskResponse)
		if taskResp.Title != "New Task" {
			t.Errorf("expected title 'New Task', got '%s'", taskResp.Title)
		}
		if taskResp.Status != "todo" {
			t.Errorf("expected status 'todo', got '%s'", taskResp.Status)
		}
		if taskResp.Priority != shared.TaskPriorityMedium {
			t.Errorf("expected priority 'medium', got '%s'", taskResp.Priority)
		}
	})

	t.Run("creates task with all fields", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1", WorkspaceID: "ws-1"}
		args := map[string]interface{}{
			"planId":      "plan-1",
			"title":       "Full Task",
			"description": "Task description",
			"priority":    "high",
			"assigneeId":  "user-2",
			"dueDate":     "2024-12-31",
			"tags":        []interface{}{"urgent", "backend"},
		}

		result, err := tool.Execute(context.Background(), execCtx, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		taskResp := data["task"].(*task.TaskResponse)
		if taskResp.Priority != "high" {
			t.Errorf("expected priority 'high', got '%s'", taskResp.Priority)
		}
		if *taskResp.AssigneeID != "user-2" {
			t.Errorf("expected assignee 'user-2', got '%s'", *taskResp.AssigneeID)
		}
	})

	t.Run("requires plan ID", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1", WorkspaceID: "ws-1"}
		args := map[string]interface{}{
			"title": "Task without plan",
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for missing plan ID")
		}
	})

	t.Run("requires title", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1", WorkspaceID: "ws-1"}
		args := map[string]interface{}{
			"planId": "plan-1",
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for missing title")
		}
	})

	t.Run("validates priority", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1", WorkspaceID: "ws-1"}
		args := map[string]interface{}{
			"planId":   "plan-1",
			"title":    "Task",
			"priority": "invalid",
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for invalid priority")
		}
	})
}

func TestListTasksTool(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	tool := NewListTasksTool(taskRepo)

	// Create test tasks
	t1 := task.NewTask("task-1", "plan-1", "Task 1", "todo", shared.TaskPriorityMedium)
	t2 := task.NewTask("task-2", "plan-1", "Task 2", "in_progress", shared.TaskPriorityHigh)
	t3 := task.NewTask("task-3", "plan-2", "Task 3", "todo", shared.TaskPriorityLow)
	_ = taskRepo.Create(context.Background(), t1)
	_ = taskRepo.Create(context.Background(), t2)
	_ = taskRepo.Create(context.Background(), t3)

	t.Run("lists tasks for plan", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"planId": "plan-1",
		}

		result, err := tool.Execute(context.Background(), execCtx, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		tasks := data["tasks"].([]*task.TaskResponse)
		if len(tasks) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(tasks))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"planId": "plan-1",
			"status": "todo",
		}

		result, err := tool.Execute(context.Background(), execCtx, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		tasks := data["tasks"].([]*task.TaskResponse)
		if len(tasks) != 1 {
			t.Errorf("expected 1 task, got %d", len(tasks))
		}
	})
}

func TestUpdateTaskTool(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	planRepo := NewMockPlanRepository()
	tool := NewUpdateTaskTool(taskRepo, planRepo)

	// Create test plan and task
	p := plan.NewPlan("plan-1", "ws-1", "Test Plan", "user-1", shared.PlanDomainSoftware, nil)
	_ = planRepo.Create(context.Background(), p)

	tk := task.NewTask("task-1", "plan-1", "Original Title", "todo", shared.TaskPriorityMedium)
	_ = taskRepo.Create(context.Background(), tk)

	t.Run("updates task title", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"taskId": "task-1",
			"title":  "Updated Title",
		}

		result, err := tool.Execute(context.Background(), execCtx, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		taskResp := data["task"].(*task.TaskResponse)
		if taskResp.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got '%s'", taskResp.Title)
		}
	})

	t.Run("validates status against plan", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"taskId": "task-1",
			"status": "invalid_status",
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for invalid status")
		}
	})

	t.Run("requires at least one field", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"taskId": "task-1",
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for no fields to update")
		}
	})
}

func TestMoveTaskTool(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	planRepo := NewMockPlanRepository()
	tool := NewMoveTaskTool(taskRepo, planRepo)

	// Create test plan and task
	p := plan.NewPlan("plan-1", "ws-1", "Test Plan", "user-1", shared.PlanDomainSoftware, nil)
	_ = planRepo.Create(context.Background(), p)

	tk := task.NewTask("task-1", "plan-1", "Task 1", "todo", shared.TaskPriorityMedium)
	_ = taskRepo.Create(context.Background(), tk)

	t.Run("moves task to new status", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"taskId":   "task-1",
			"status":   "in_progress",
			"position": float64(0),
		}

		result, err := tool.Execute(context.Background(), execCtx, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		taskResp := data["task"].(*task.TaskResponse)
		if taskResp.Status != "in_progress" {
			t.Errorf("expected status 'in_progress', got '%s'", taskResp.Status)
		}
	})

	t.Run("validates status against plan", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"taskId": "task-1",
			"status": "invalid_status",
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for invalid status")
		}
	})
}

func TestSearchTasksTool(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	tool := NewSearchTasksTool(taskRepo)

	// Create test tasks
	t1 := task.NewTask("task-1", "plan-1", "Backend API task", "todo", shared.TaskPriorityMedium)
	t2 := task.NewTask("task-2", "plan-1", "Frontend UI task", "todo", shared.TaskPriorityMedium)
	_ = taskRepo.Create(context.Background(), t1)
	_ = taskRepo.Create(context.Background(), t2)

	t.Run("searches tasks", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"planId": "plan-1",
			"query":  "backend",
		}

		result, err := tool.Execute(context.Background(), execCtx, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		if _, ok := data["tasks"]; !ok {
			t.Error("expected tasks in result")
		}
	})

	t.Run("requires query", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"planId": "plan-1",
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for missing query")
		}
	})
}
