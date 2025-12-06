package mcp

import (
	"github.com/FantasyRL/go-mcp-demo/internal/mcp/application"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
)

func InjectDependencies() {
	// Inject dependencies here

	AIProviderClient := ai_provider.NewAiProviderClient()
	application.NewAISESolver(AIProviderClient)

	// 初始化全局 ClientSet，包含数据库连接
	clientSet := base.NewClientSet(
		base.WithDB(),
		base.WithCache(),
	)
	base.SetGlobalClientSet(clientSet)
}
