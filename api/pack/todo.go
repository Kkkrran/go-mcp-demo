package pack

import (
	api "github.com/FantasyRL/go-mcp-demo/api/model/api"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
)

// BuildTodoItem 构建单个待办事项响应
func BuildTodoItem(todo *model.Todolists) *api.TodoItem {
	item := &api.TodoItem{
		ID:        todo.ID,
		Title:     todo.Title,
		Content:   todo.Content,
		StartTime: todo.StartTime.UnixMilli(),
		EndTime:   todo.EndTime.UnixMilli(),
		IsAllDay:  todo.IsAllDay,
		Status:    todo.Status,
		Priority:  todo.Priority,
		CreatedAt: todo.CreatedAt.UnixMilli(),
		UpdatedAt: todo.UpdatedAt.UnixMilli(),
	}

	if todo.RemindAt != nil {
		remindAt := todo.RemindAt.UnixMilli()
		item.RemindAt = &remindAt
	}

	if todo.Category != nil {
		item.Category = todo.Category
	}

	return item
} // BuildTodoList 构建待办事项列表响应
func BuildTodoList(todos []*model.Todolists) []*api.TodoItem {
	items := make([]*api.TodoItem, 0, len(todos))
	for _, todo := range todos {
		items = append(items, BuildTodoItem(todo))
	}
	return items
}
