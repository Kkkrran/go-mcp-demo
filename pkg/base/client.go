package base

import (
	"sync"

	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_client"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/registry"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	instance       *ClientSet
	globalInstance *ClientSet
	once           sync.Once
	globalMutex    sync.RWMutex
)

// ClientSet storage various client objects
// Notice: some or all of them maybe nil, we should check obj when use
type ClientSet struct {
	MCPCli           mcp_client.ToolClient
	AiProviderCli    *ai_provider.Client
	RegistryResolver registry.Resolver
	ActualDB         *gorm.DB
	Cache            *redis.Client
	cleanups         []func()
}

type Option func(clientSet *ClientSet)

// NewClientSet will be protected by sync.Once for ensure only 1 instance could be created in 1 lifecycle
func NewClientSet(opt ...Option) *ClientSet {
	once.Do(func() {
		var options []Option
		instance = &ClientSet{}
		options = append(options, opt...)
		for _, opt := range options {
			opt(instance)
		}
	})
	return instance
}
func (cs *ClientSet) Close() {
	for _, cleanup := range cs.cleanups {
		cleanup()
	}
}

// SetGlobalClientSet 设置全局 ClientSet 实例（用于 MCP 服务等场景）
func SetGlobalClientSet(cs *ClientSet) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalInstance = cs
}

// GetGlobalClientSet 获取全局 ClientSet 实例
func GetGlobalClientSet() *ClientSet {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalInstance
}
