package infra

import (
	"context"
	"errors"

	"github.com/FantasyRL/go-mcp-demo/internal/mcp/repository"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
	"gorm.io/gorm"
)

var _ repository.MCPRepository = (*MCPInfra)(nil)

// MCPInfra implements repository.MCPRepository
type MCPInfra struct {
	db *gorm.DB
}

// NewMCPRepository creates a new MCPRepository instance
func NewMCPRepository() repository.MCPRepository {
	clientSet := base.GetGlobalClientSet()
	if clientSet == nil {
		panic("global ClientSet not initialized")
	}
	return &MCPInfra{
		db: clientSet.ActualDB,
	}
}

// ListTodosByUserID 获取用户的所有待办事项列表
func (r *MCPInfra) ListTodosByUserID(ctx context.Context, userID string) ([]*model.Todolists, error) {
	var todos []*model.Todolists

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&todos).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return todos, nil
}
