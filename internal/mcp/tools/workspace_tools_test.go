package tools

import (
	"context"
	"testing"

	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/domain/workspace"
	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/repository"
)

// MockWorkspaceRepository implements repository.WorkspaceRepository for testing.
type MockWorkspaceRepository struct {
	workspaces map[string]*workspace.Workspace
	members    map[string][]*workspace.WorkspaceMember
	slugExists map[string]bool
}

func NewMockWorkspaceRepository() *MockWorkspaceRepository {
	return &MockWorkspaceRepository{
		workspaces: make(map[string]*workspace.Workspace),
		members:    make(map[string][]*workspace.WorkspaceMember),
		slugExists: make(map[string]bool),
	}
}

func (m *MockWorkspaceRepository) Create(ctx context.Context, w *workspace.Workspace) error {
	m.workspaces[w.ID] = w
	m.slugExists[w.Slug] = true
	// Add owner as member
	member := workspace.NewWorkspaceMember(w.ID, w.OwnerID, "owner")
	m.members[w.ID] = append(m.members[w.ID], member)
	return nil
}

func (m *MockWorkspaceRepository) GetByID(ctx context.Context, id string) (*workspace.Workspace, error) {
	if ws, ok := m.workspaces[id]; ok {
		return ws, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockWorkspaceRepository) GetBySlug(ctx context.Context, slug string) (*workspace.Workspace, error) {
	for _, ws := range m.workspaces {
		if ws.Slug == slug {
			return ws, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (m *MockWorkspaceRepository) Update(ctx context.Context, id string, input *workspace.UpdateWorkspaceInput) (*workspace.Workspace, error) {
	ws, ok := m.workspaces[id]
	if !ok {
		return nil, shared.ErrNotFound
	}
	if input.Name != nil {
		ws.Name = *input.Name
	}
	if input.Settings != nil {
		ws.Settings = *input.Settings
	}
	return ws, nil
}

func (m *MockWorkspaceRepository) Delete(ctx context.Context, id string) error {
	delete(m.workspaces, id)
	return nil
}

func (m *MockWorkspaceRepository) List(ctx context.Context, filter repository.WorkspaceFilterOptions, pagination repository.Pagination) (*repository.PaginatedResult[workspace.Workspace], error) {
	var items []workspace.Workspace
	for _, ws := range m.workspaces {
		items = append(items, *ws)
	}
	return repository.NewPaginatedResult(items, int64(len(items)), pagination), nil
}

func (m *MockWorkspaceRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return m.slugExists[slug], nil
}

func (m *MockWorkspaceRepository) AddMember(ctx context.Context, member *workspace.WorkspaceMember) error {
	m.members[member.WorkspaceID] = append(m.members[member.WorkspaceID], member)
	return nil
}

func (m *MockWorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	return nil
}

func (m *MockWorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	return nil
}

func (m *MockWorkspaceRepository) GetMember(ctx context.Context, workspaceID, userID string) (*workspace.WorkspaceMember, error) {
	members := m.members[workspaceID]
	for _, member := range members {
		if member.UserID == userID {
			return member, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (m *MockWorkspaceRepository) ListMembers(ctx context.Context, workspaceID string) ([]*workspace.WorkspaceMember, error) {
	return m.members[workspaceID], nil
}

func (m *MockWorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*workspace.Workspace, error) {
	var result []*workspace.Workspace
	for wsID, members := range m.members {
		for _, member := range members {
			if member.UserID == userID {
				if ws, ok := m.workspaces[wsID]; ok {
					result = append(result, ws)
				}
				break
			}
		}
	}
	return result, nil
}

func TestListWorkspacesTool(t *testing.T) {
	repo := NewMockWorkspaceRepository()
	tool := NewListWorkspacesTool(repo)

	// Create test workspace
	ws := workspace.NewWorkspace("ws-1", "Test Workspace", "test-ws", "user-1", nil)
	_ = repo.Create(context.Background(), ws)

	t.Run("lists user workspaces", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		result, err := tool.Execute(context.Background(), execCtx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		workspaces := data["workspaces"].([]*workspace.WorkspaceResponse)
		if len(workspaces) != 1 {
			t.Errorf("expected 1 workspace, got %d", len(workspaces))
		}
	})

	t.Run("requires user ID", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{}
		_, err := tool.Execute(context.Background(), execCtx, nil)
		if err == nil {
			t.Error("expected error for missing user ID")
		}
	})
}

func TestCreateWorkspaceTool(t *testing.T) {
	repo := NewMockWorkspaceRepository()
	tool := NewCreateWorkspaceTool(repo)

	t.Run("creates workspace successfully", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"name": "New Workspace",
			"slug": "new-workspace",
		}

		result, err := tool.Execute(context.Background(), execCtx, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		ws := data["workspace"].(*workspace.WorkspaceResponse)
		if ws.Name != "New Workspace" {
			t.Errorf("expected name 'New Workspace', got '%s'", ws.Name)
		}
		if ws.Slug != "new-workspace" {
			t.Errorf("expected slug 'new-workspace', got '%s'", ws.Slug)
		}
	})

	t.Run("fails on duplicate slug", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"name": "Another Workspace",
			"slug": "new-workspace", // Already exists
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for duplicate slug")
		}
	})

	t.Run("requires name", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"slug": "test-slug",
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for missing name")
		}
	})

	t.Run("requires slug", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"name": "Test Workspace",
		}

		_, err := tool.Execute(context.Background(), execCtx, args)
		if err == nil {
			t.Error("expected error for missing slug")
		}
	})
}

func TestGetWorkspaceTool(t *testing.T) {
	repo := NewMockWorkspaceRepository()
	tool := NewGetWorkspaceTool(repo)

	// Create test workspace
	ws := workspace.NewWorkspace("ws-1", "Test Workspace", "test-ws", "user-1", nil)
	_ = repo.Create(context.Background(), ws)

	t.Run("gets workspace by ID", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1"}
		args := map[string]interface{}{
			"workspaceId": "ws-1",
		}

		result, err := tool.Execute(context.Background(), execCtx, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		wsResp := data["workspace"].(*workspace.WorkspaceResponse)
		if wsResp.ID != "ws-1" {
			t.Errorf("expected ID 'ws-1', got '%s'", wsResp.ID)
		}

		members := data["members"].([]*workspace.WorkspaceMemberResponse)
		if len(members) != 1 {
			t.Errorf("expected 1 member, got %d", len(members))
		}
	})

	t.Run("uses execution context workspace ID", func(t *testing.T) {
		execCtx := mcp.ExecutionContext{UserID: "user-1", WorkspaceID: "ws-1"}

		result, err := tool.Execute(context.Background(), execCtx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := result.(map[string]interface{})
		wsResp := data["workspace"].(*workspace.WorkspaceResponse)
		if wsResp.ID != "ws-1" {
			t.Errorf("expected ID 'ws-1', got '%s'", wsResp.ID)
		}
	})
}
