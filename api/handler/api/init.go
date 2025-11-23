package api

import (
	"context"

	"github.com/FantasyRL/go-mcp-demo/internal/host/application"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
)

var clientSet *base.ClientSet
var host *application.Host

func Init() {
	clientSet = base.NewClientSet(
		base.WithMCPClient([]string{constant.ServiceNameMCPLocal, constant.ServiceNameMCPRemote}),
		base.WithAiProviderClient(),
		base.WithDB(),
	)
	host = application.NewHost(context.Background(), clientSet)
}

func GetHost() *application.Host { return host }
