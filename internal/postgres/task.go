package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/domain/task"
	"github.com/goplan/goplan/internal/postgres/sqlc"
	"github.com/goplan/goplan/internal/repository"
)

// TaskRepository implements repository.TaskRepository using PostgreSQL.
type TaskRepository struct {
	pool *pgxpool.Pool
}

// NewTaskRepository creates a new TaskRepository.
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

// Create creates a new task.
func (r *TaskRepository) Create(ctx context.Context, t *task.Task) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	// Get max position for the plan
	maxPos, err := q.GetMaxTaskPosition(ctx, t.PlanID)
	if err != nil {
		return MapError(err, "task")
	}

	customFieldsJSON, err := json.Marshal(t.CustomFieldValues)
	if err != nil {
		return err
	}

	result, err := q.CreateTask(ctx, sqlc.CreateTaskParams{
		PlanID:            t.PlanID,
		PhaseID:           uuidFromPtr(t.PhaseID),
		ParentID:          uuidFromPtr(t.ParentID),
		Title:             t.Title,
		Description:       textPtrFromPtr(t.Description),
		Status:            t.Status,
		Priority:          textPtr(t.Priority),
		AssigneeID:        uuidFromPtr(t.AssigneeID),
		DueDate:           dateFromPtr(t.DueDate),
		EstimatedHours:    numericFromPtr(t.EstimatedHours),
		CustomFieldValues: customFieldsJSON,
		Tags:              t.Tags,
		Position:          int4FromInt(int(maxPos + 1)),
	})
	if err != nil {
		return MapError(err, "task")
	}

	t.ID = result.ID
	t.Position = int4ToInt(result.Position)
	t.CreatedAt = result.CreatedAt.Time
	t.UpdatedAt = result.UpdatedAt.Time

	return nil
}

// GetByID retrieves a task by ID.
func (r *TaskRepository) GetByID(ctx context.Context, id string) (*task.Task, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetTaskByID(ctx, id)
	if err != nil {
		return nil, MapError(err, "task")
	}

	return sqlcTaskToDomain(result), nil
}

// GetByIDWithDetails retrieves a task with subtasks, dependencies, and comments.
func (r *TaskRepository) GetByIDWithDetails(ctx context.Context, id string) (*task.Task, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetTaskByID(ctx, id)
	if err != nil {
		return nil, MapError(err, "task")
	}

	t := sqlcTaskToDomain(result)

	// Get subtasks
	subtasks, err := q.ListSubtasks(ctx, uuidFromString(id))
	if err != nil {
		return nil, MapError(err, "task")
	}
	t.Subtasks = make([]task.Task, len(subtasks))
	for i, st := range subtasks {
		t.Subtasks[i] = *sqlcTaskToDomain(st)
	}

	// Get dependencies
	deps, err := q.GetTaskDependencies(ctx, id)
	if err != nil {
		return nil, MapError(err, "task")
	}
	t.Dependencies = make([]task.TaskDependency, len(deps))
	for i, d := range deps {
		t.Dependencies[i] = task.TaskDependency{
			TaskID:      id,
			DependsOnID: d.ID,
			Type:        "depends_on",
		}
	}

	return t, nil
}

// Update updates task fields.
func (r *TaskRepository) Update(ctx context.Context, id string, input *task.UpdateTaskInput) (*task.Task, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	params := sqlc.UpdateTaskParams{
		ID: id,
	}
	if input.Title != nil {
		params.Title = textPtr(*input.Title)
	}
	if input.Description != nil {
		params.Description = textPtr(*input.Description)
	}
	if input.Status != nil {
		params.Status = textPtr(*input.Status)
	}
	if input.Priority != nil {
		params.Priority = textPtr(*input.Priority)
	}
	if input.PhaseID != nil {
		params.PhaseID = uuidFromPtr(input.PhaseID)
	}
	if input.AssigneeID != nil {
		params.AssigneeID = uuidFromPtr(input.AssigneeID)
	}
	if input.DueDate != nil {
		params.DueDate = dateFromPtr(input.DueDate)
	}
	if input.EstimatedHours != nil {
		params.EstimatedHours = numericFromPtr(input.EstimatedHours)
	}
	if input.CustomFieldValues != nil {
		customFieldsJSON, err := json.Marshal(input.CustomFieldValues)
		if err != nil {
			return nil, err
		}
		params.CustomFieldValues = customFieldsJSON
	}
	if input.Tags != nil {
		params.Tags = input.Tags
	}
	if input.Position != nil {
		params.Position = int4FromInt(*input.Position)
	}

	result, err := q.UpdateTask(ctx, params)
	if err != nil {
		return nil, MapError(err, "task")
	}

	return sqlcTaskToDomain(result), nil
}

// Delete deletes a task by ID.
func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.DeleteTask(ctx, id)
	if err != nil {
		return MapError(err, "task")
	}

	return nil
}

// List retrieves tasks with filtering, sorting, and pagination.
func (r *TaskRepository) List(ctx context.Context, filter repository.TaskFilterOptions, sort repository.TaskSortOptions, pagination repository.Pagination) (*repository.PaginatedResult[task.Task], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	if filter.PlanID == nil {
		return nil, nil
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	params := sqlc.ListTasksByPlanFilteredParams{
		PlanID: *filter.PlanID,
		Limit:  int32(pagination.PageSize),
		Offset: int32(offset),
	}

	if filter.PhaseID != nil {
		params.PhaseID = uuidFromPtr(filter.PhaseID)
	}
	if filter.ParentID != nil {
		params.ParentID = uuidFromPtr(filter.ParentID)
	}
	if filter.Status != nil {
		params.Status = textPtr(*filter.Status)
	}
	if filter.Priority != nil {
		params.Priority = textPtr(*filter.Priority)
	}
	if filter.AssigneeID != nil {
		params.AssigneeID = uuidFromPtr(filter.AssigneeID)
	}
	if filter.DueBefore != nil {
		params.DueBefore = dateFromPtr(filter.DueBefore)
	}
	if filter.DueAfter != nil {
		params.DueAfter = dateFromPtr(filter.DueAfter)
	}

	tasks, err := q.ListTasksByPlanFiltered(ctx, params)
	if err != nil {
		return nil, MapError(err, "task")
	}

	countParams := sqlc.CountTasksByPlanFilteredParams{
		PlanID: *filter.PlanID,
	}
	if filter.PhaseID != nil {
		countParams.PhaseID = uuidFromPtr(filter.PhaseID)
	}
	if filter.ParentID != nil {
		countParams.ParentID = uuidFromPtr(filter.ParentID)
	}
	if filter.Status != nil {
		countParams.Status = textPtr(*filter.Status)
	}
	if filter.Priority != nil {
		countParams.Priority = textPtr(*filter.Priority)
	}
	if filter.AssigneeID != nil {
		countParams.AssigneeID = uuidFromPtr(filter.AssigneeID)
	}
	if filter.DueBefore != nil {
		countParams.DueBefore = dateFromPtr(filter.DueBefore)
	}
	if filter.DueAfter != nil {
		countParams.DueAfter = dateFromPtr(filter.DueAfter)
	}

	count, err := q.CountTasksByPlanFiltered(ctx, countParams)
	if err != nil {
		return nil, MapError(err, "task")
	}

	items := make([]task.Task, len(tasks))
	for i, t := range tasks {
		items[i] = *sqlcTaskToDomain(t)
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// ListByPlan retrieves all tasks for a plan.
func (r *TaskRepository) ListByPlan(ctx context.Context, planID string, pagination repository.Pagination) (*repository.PaginatedResult[task.Task], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	tasks, err := q.ListTasksByPlan(ctx, sqlc.ListTasksByPlanParams{
		PlanID: planID,
		Limit:  int32(pagination.PageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, MapError(err, "task")
	}

	count, err := q.CountTasksByPlan(ctx, planID)
	if err != nil {
		return nil, MapError(err, "task")
	}

	items := make([]task.Task, len(tasks))
	for i, t := range tasks {
		items[i] = *sqlcTaskToDomain(t)
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// GetSubtasks retrieves direct subtasks of a task.
func (r *TaskRepository) GetSubtasks(ctx context.Context, parentID string) ([]*task.Task, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	tasks, err := q.ListSubtasks(ctx, uuidFromString(parentID))
	if err != nil {
		return nil, MapError(err, "task")
	}

	result := make([]*task.Task, len(tasks))
	for i, t := range tasks {
		result[i] = sqlcTaskToDomain(t)
	}

	return result, nil
}

// Search performs full-text search on tasks.
func (r *TaskRepository) Search(ctx context.Context, planID, query string, pagination repository.Pagination) (*repository.PaginatedResult[task.Task], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	tasks, err := q.SearchTasks(ctx, sqlc.SearchTasksParams{
		PlanID:         planID,
		PlaintoTsquery: query,
		Limit:          int32(pagination.PageSize),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, MapError(err, "task")
	}

	count, err := q.CountSearchTasks(ctx, sqlc.CountSearchTasksParams{
		PlanID:         planID,
		PlaintoTsquery: query,
	})
	if err != nil {
		return nil, MapError(err, "task")
	}

	items := make([]task.Task, len(tasks))
	for i, t := range tasks {
		items[i] = *sqlcTaskToDomain(t)
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// Move updates task status and position.
func (r *TaskRepository) Move(ctx context.Context, id string, status string, position int) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	// Get current task to preserve phase_id
	current, err := q.GetTaskByID(ctx, id)
	if err != nil {
		return MapError(err, "task")
	}

	_, err = q.MoveTask(ctx, sqlc.MoveTaskParams{
		ID:       id,
		PhaseID:  current.PhaseID,
		Position: int4FromInt(position),
	})
	if err != nil {
		return MapError(err, "task")
	}

	// Update status separately
	_, err = q.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return MapError(err, "task")
	}

	return nil
}

// BatchUpdatePositions updates positions for multiple tasks.
func (r *TaskRepository) BatchUpdatePositions(ctx context.Context, updates map[string]int) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	for id, position := range updates {
		_, err := q.UpdateTaskPosition(ctx, sqlc.UpdateTaskPositionParams{
			ID:       id,
			Position: int4FromInt(position),
		})
		if err != nil {
			return MapError(err, "task")
		}
	}

	return nil
}

// AddDependency adds a task dependency.
func (r *TaskRepository) AddDependency(ctx context.Context, dep *task.TaskDependency) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.AddTaskDependency(ctx, sqlc.AddTaskDependencyParams{
		TaskID:      dep.TaskID,
		DependsOnID: dep.DependsOnID,
	})
	if err != nil {
		return MapError(err, "task_dependency")
	}

	return nil
}

// RemoveDependency removes a task dependency.
func (r *TaskRepository) RemoveDependency(ctx context.Context, taskID, dependsOnID string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.RemoveTaskDependency(ctx, sqlc.RemoveTaskDependencyParams{
		TaskID:      taskID,
		DependsOnID: dependsOnID,
	})
	if err != nil {
		return MapError(err, "task_dependency")
	}

	return nil
}

// GetDependencies retrieves all dependencies for a task.
func (r *TaskRepository) GetDependencies(ctx context.Context, taskID string) ([]*task.TaskDependency, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	deps, err := q.GetTaskDependencies(ctx, taskID)
	if err != nil {
		return nil, MapError(err, "task_dependency")
	}

	result := make([]*task.TaskDependency, len(deps))
	for i, d := range deps {
		result[i] = &task.TaskDependency{
			TaskID:      taskID,
			DependsOnID: d.ID,
			Type:        "depends_on",
		}
	}

	return result, nil
}

// GetDependents retrieves all tasks that depend on this task.
func (r *TaskRepository) GetDependents(ctx context.Context, taskID string) ([]*task.TaskDependency, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	deps, err := q.GetTaskDependents(ctx, taskID)
	if err != nil {
		return nil, MapError(err, "task_dependency")
	}

	result := make([]*task.TaskDependency, len(deps))
	for i, d := range deps {
		result[i] = &task.TaskDependency{
			TaskID:      d.ID,
			DependsOnID: taskID,
			Type:        "depends_on",
		}
	}

	return result, nil
}

// GetTasksByStatus retrieves tasks grouped by status for a plan.
func (r *TaskRepository) GetTasksByStatus(ctx context.Context, planID string) (map[string][]*task.Task, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	tasks, err := q.GetTasksForKanban(ctx, planID)
	if err != nil {
		return nil, MapError(err, "task")
	}

	result := make(map[string][]*task.Task)
	for _, t := range tasks {
		domainTask := sqlcTaskToDomain(t)
		result[domainTask.Status] = append(result[domainTask.Status], domainTask)
	}

	return result, nil
}

// CountByStatus counts tasks by status for a plan.
func (r *TaskRepository) CountByStatus(ctx context.Context, planID string) (map[string]int, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	counts, err := q.CountTasksByStatus(ctx, planID)
	if err != nil {
		return nil, MapError(err, "task")
	}

	result := make(map[string]int)
	for _, c := range counts {
		result[c.Status] = int(c.Count)
	}

	return result, nil
}

// GetOverdue retrieves overdue tasks.
func (r *TaskRepository) GetOverdue(ctx context.Context, planID string) ([]*task.Task, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	tasks, err := q.GetOverdueTasks(ctx, planID)
	if err != nil {
		return nil, MapError(err, "task")
	}

	result := make([]*task.Task, len(tasks))
	for i, t := range tasks {
		result[i] = sqlcTaskToDomain(t)
	}

	return result, nil
}

// Helper functions

func sqlcTaskToDomain(t sqlc.Task) *task.Task {
	var customFieldValues map[string]interface{}
	if len(t.CustomFieldValues) > 0 {
		if err := json.Unmarshal(t.CustomFieldValues, &customFieldValues); err != nil {
			customFieldValues = make(map[string]interface{})
		}
	} else {
		customFieldValues = make(map[string]interface{})
	}

	tags := t.Tags
	if tags == nil {
		tags = []string{}
	}

	priority := textToString(t.Priority)
	if priority == "" {
		priority = shared.TaskPriorityMedium
	}

	return &task.Task{
		ID:                t.ID,
		PlanID:            t.PlanID,
		PhaseID:           uuidToPtr(t.PhaseID),
		ParentID:          uuidToPtr(t.ParentID),
		Title:             t.Title,
		Description:       textToPtr(t.Description),
		Status:            t.Status,
		Priority:          priority,
		AssigneeID:        uuidToPtr(t.AssigneeID),
		DueDate:           dateToPtr(t.DueDate),
		EstimatedHours:    numericToFloat64Ptr(t.EstimatedHours),
		CustomFieldValues: customFieldValues,
		Tags:              tags,
		Position:          int4ToInt(t.Position),
		CreatedAt:         t.CreatedAt.Time,
		UpdatedAt:         t.UpdatedAt.Time,
		Subtasks:          []task.Task{},
		Dependencies:      []task.TaskDependency{},
		Comments:          []task.Comment{},
	}
}
