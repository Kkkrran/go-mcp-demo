package mcp

import (
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
)

func InjectDependencies() {
	// 初始化全局 ClientSet，包含数据库、缓存和 AI provider
	clientSet := base.NewClientSet(
		base.WithDB(),
		base.WithCache(),
		base.WithAiProviderClient(),
	)
	base.SetGlobalClientSet(clientSet)
}
