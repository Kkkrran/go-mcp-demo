package repository

import (
	"context"

	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
)

// MCPRepository MCP服务的数据访问层接口
type MCPRepository interface {
	// ListTodosByUserID 获取用户的所有待办事项列表
	ListTodosByUserID(ctx context.Context, userID string) ([]*model.Todolists, error)
}
