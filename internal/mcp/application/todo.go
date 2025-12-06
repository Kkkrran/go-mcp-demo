package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FantasyRL/go-mcp-demo/internal/mcp/infra"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/mark3labs/mcp-go/mcp"
)

// WithTodoTools 注册待办事项相关的 MCP 工具
func WithTodoTools() tool_set.Option {
	return func(ts *tool_set.ToolSet) {
		// 初始化 repository
		repo := infra.NewMCPRepository()

		// 定义 get_todos 工具
		tool := mcp.NewTool(
			"get_todos",
			mcp.WithDescription("获取指定用户的待办事项列表"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("用户ID")),
		)

		// 注册工具
		ts.Tools = append(ts.Tools, &tool)
		ts.HandlerFunc[tool.Name] = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// 解析参数
			args := req.GetArguments()
			rawUserID, ok := args["user_id"]
			if !ok {
				return mcp.NewToolResultError("missing required arg: user_id"), nil
			}

			userID, ok := rawUserID.(string)
			if !ok || userID == "" {
				return mcp.NewToolResultError("user_id must be a non-empty string"), nil
			}

			// 查询数据库
			todos, err := repo.ListTodosByUserID(ctx, userID)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error querying todos: %v", err)), nil
			}

			// 如果没有待办事项
			if len(todos) == 0 {
				return mcp.NewToolResultText("[]"), nil
			}

			// 构造返回数据结构
			type TodoItem struct {
				ID        string `json:"id"`
				Title     string `json:"title"`
				Content   string `json:"content"`
				StartTime string `json:"start_time"`
				EndTime   string `json:"end_time"`
				IsAllDay  int16  `json:"is_all_day"`
				Status    int16  `json:"status"`
				Priority  int16  `json:"priority"`
				Category  string `json:"category"`
			}

			var items []TodoItem
			for _, todo := range todos {
				category := ""
				if todo.Category != nil {
					category = *todo.Category
				}

				items = append(items, TodoItem{
					ID:        todo.ID,
					Title:     todo.Title,
					Content:   todo.Content,
					StartTime: todo.StartTime.Format("2006-01-02 15:04:05"),
					EndTime:   todo.EndTime.Format("2006-01-02 15:04:05"),
					IsAllDay:  todo.IsAllDay,
					Status:    todo.Status,
					Priority:  todo.Priority,
					Category:  category,
				})
			}

			// 序列化为 JSON
			jsonData, err := json.MarshalIndent(items, "", "  ")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
			}

			return mcp.NewToolResultText(string(jsonData)), nil
		}
	}
}
