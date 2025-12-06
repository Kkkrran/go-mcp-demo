package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/west2-online/jwch"
)

// WithCourseTools 注册课表相关的 MCP 工具
func WithCourseTools() tool_set.Option {
	return func(ts *tool_set.ToolSet) {
		tool := mcp.NewTool(
			"get_course",
			mcp.WithDescription("从Redis缓存获取指定用户的课表信息"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("用户ID")),
			mcp.WithString("term", mcp.Required(), mcp.Description("学期代码，如 202501")),
		)

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

			rawTerm, ok := args["term"]
			if !ok {
				return mcp.NewToolResultError("missing required arg: term"), nil
			}
			term, ok := rawTerm.(string)
			if !ok || term == "" {
				return mcp.NewToolResultError("term must be a non-empty string"), nil
			}

			// 获取 Redis 客户端
			clientSet := base.GetGlobalClientSet()
			if clientSet == nil || clientSet.Cache == nil {
				return mcp.NewToolResultError("Redis client not initialized"), nil
			}

			// 构造 Redis key（与 host 服务保持一致）
			courseKey := fmt.Sprintf("course:%s:%s", userID, term)

			// 查询 Redis
			data, err := clientSet.Cache.Get(ctx, courseKey).Bytes()
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get course from cache: %v", err)), nil
			}

			// 反序列化课表数据
			var courses []*jwch.Course
			if err := json.Unmarshal(data, &courses); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to unmarshal course data: %v", err)), nil
			}

			if len(courses) == 0 {
				return mcp.NewToolResultText("[]"), nil
			}

			// 构造返回的数据结构
			type ScheduleRule struct {
				StartClass int    `json:"start_class"`
				EndClass   int    `json:"end_class"`
				StartWeek  int    `json:"start_week"`
				EndWeek    int    `json:"end_week"`
				Weekday    int    `json:"weekday"`
				Single     bool   `json:"single"`
				Double     bool   `json:"double"`
				Adjust     bool   `json:"adjust"`
				Location   string `json:"location"`
			}

			type CourseItem struct {
				Name          string         `json:"name"`
				Teacher       string         `json:"teacher"`
				ScheduleRules []ScheduleRule `json:"schedule_rules"`
				Remark        string         `json:"remark,omitempty"`
			}

			var items []CourseItem
			for _, course := range courses {
				rules := make([]ScheduleRule, 0, len(course.ScheduleRules))
				for _, rule := range course.ScheduleRules {
					rules = append(rules, ScheduleRule{
						StartClass: rule.StartClass,
						EndClass:   rule.EndClass,
						StartWeek:  rule.StartWeek,
						EndWeek:    rule.EndWeek,
						Weekday:    rule.Weekday,
						Single:     rule.Single,
						Double:     rule.Double,
						Adjust:     rule.Adjust,
						Location:   rule.Location,
					})
				}

				items = append(items, CourseItem{
					Name:          course.Name,
					Teacher:       course.Teacher,
					ScheduleRules: rules,
					Remark:        course.Remark,
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
