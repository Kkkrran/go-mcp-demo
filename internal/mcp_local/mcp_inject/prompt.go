package mcp_inject

import (
	"context"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/prompt_set"
	"github.com/mark3labs/mcp-go/mcp"
)

func WithBuildHTMLPrompt() prompt_set.Option {
	return func(ps *prompt_set.PromptSet) {
		// 1) 定义 Prompt 的 schema
		p := mcp.NewPrompt("greeting",
			mcp.WithPromptDescription("A friendly greeting prompt"),
			mcp.WithArgument("name",
				mcp.ArgumentDescription("Name of the person to greet"),
			),
		)

		// 2) handler：把参数组织成一组消息模板返回给客户端/模型
		handle := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			name := request.Params.Arguments["name"]
			if name == "" {
				name = "friend"
			}

			return mcp.NewGetPromptResult(
				"A friendly greeting",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(
						mcp.RoleAssistant,
						mcp.NewTextContent(fmt.Sprintf("Hello, %s! How can I help you today?", name)),
					),
				},
			), nil
		}

		ps.Prompts = append(ps.Prompts, &p)
		ps.HandlerFunc[p.Name] = handle
	}
}
