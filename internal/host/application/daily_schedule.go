package application

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	openai "github.com/openai/openai-go/v2"
)

// dailySchedulePrompt 专门用于生成每日日程的系统提示词
const dailySchedulePrompt = `你是一个智能日程助手，需要根据用户的课表和待办事项，生成今日的完整日程安排。

## 任务说明
1. 调用 get_course 工具获取用户的课表信息（使用当前学期代码 202501）
2. 调用 get_todos 工具获取用户的待办事项列表
3. 分析今天是星期几，筛选出今天的课程
4. 结合课程和待办事项，生成一份清晰的今日安排

## 学期代码规则
- 当前是 2025年12月，当前学期是 202501（2025年秋季学期）
- 学期代码格式：YYYYSS，01表示秋季学期，02表示春季学期

## 课程节次与时间对应关系
- 第 1-2 节：08:20 - 10:00
- 第 3-4 节：10:20 - 12:00
- 第 5-6 节：14:00 - 15:40
- 第 7-8 节：15:50 - 17:30
- 第 9-11 节：19:00 - 21:35

## 输出格式要求
生成简洁清晰的今日安排，格式如下：

📅 今日课程安排
- 08:20-10:00 课程名称（教师）@ 地点
- 10:20-12:00 课程名称（教师）@ 地点

📝 今日待办事项
- [优先级1] 标题 (截止时间)
- [优先级2] 标题 (截止时间)

💡 温馨提示
- 提醒用户注意重要事项
- 给出合理的时间规划建议

注意：
1. 只显示今天的课程，根据 weekday 字段过滤
2. 考虑单双周规则（single/double 字段）
3. 待办事项按优先级排序（1最高，4最低）
4. 只显示未完成的待办（status=0）
5. 如果今天没有课程或待办，友好地告知用户
`

// GetDailySchedule 获取每日日程安排（带Redis缓存）
func (h *Host) GetDailySchedule(userID string) (string, error) {
	// 1. 检查 Redis 缓存
	cacheKey := fmt.Sprintf("daily_schedule:%s", userID)

	if h.templateRepository.IsKeyExist(h.ctx, cacheKey) {
		cached, err := h.templateRepository.GetDailyScheduleCache(h.ctx, cacheKey)
		if err == nil && cached != "" {
			logger.Infof("GetDailySchedule: cache hit for user %s", userID)
			return cached, nil
		}
		logger.Warnf("GetDailySchedule: cache read failed: %v", err)
	}

	// 2. 缓存不存在，调用 AI 生成
	schedule, err := h.generateDailySchedule(userID)
	if err != nil {
		return "", fmt.Errorf("generate daily schedule failed: %w", err)
	}

	// 3. 存入 Redis（24小时过期）
	if err := h.templateRepository.SetDailyScheduleCache(h.ctx, cacheKey, schedule); err != nil {
		logger.Errorf("GetDailySchedule: cache write failed: %v", err)
		// 不影响返回，继续执行
	}

	return schedule, nil
}

// generateDailySchedule 使用 AI 生成每日日程
func (h *Host) generateDailySchedule(userID string) (string, error) {
	ctx := h.ctx

	// 获取当前时间信息
	now := time.Now()
	weekdayMap := map[time.Weekday]string{
		time.Monday:    "星期一",
		time.Tuesday:   "星期二",
		time.Wednesday: "星期三",
		time.Thursday:  "星期四",
		time.Friday:    "星期五",
		time.Saturday:  "星期六",
		time.Sunday:    "星期日",
	}
	weekdayName := weekdayMap[now.Weekday()]
	dateInfo := fmt.Sprintf("今天是 %s，%s", now.Format("2006年01月02日"), weekdayName)

	// 构建对话历史（只包含系统提示词和用户请求）
	hist := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(dailySchedulePrompt),
		openai.UserMessage(fmt.Sprintf("%s。请帮我生成今天的日程安排。我的用户ID是：%s", dateInfo, userID)),
	}

	// 只注册 get_todos 和 get_course 这两个工具
	allTools := h.mcpCli.ConvertToolsToOpenAI()
	tools := make([]openai.ChatCompletionToolUnionParam, 0, 2)
	for _, tool := range allTools {
		if tool.OfFunction != nil {
			name := tool.OfFunction.Function.Name
			if name == "get_todos" || name == "get_course" {
				tools = append(tools, tool)
			}
		}
	}

	// 工具调用循环
	round := 0
	maxRounds := 5 // 限制最多5轮，避免死循环

	for {
		round++
		if round > maxRounds {
			return "", fmt.Errorf("达到最大工具调用轮次(%d)", maxRounds)
		}

		// 调用 OpenAI API
		params := openai.ChatCompletionNewParams{
			Model:    openai.ChatModel(config.AiProvider.Model),
			Messages: hist,
		}
		if len(tools) > 0 {
			params.Tools = tools
		}
		if config.AiProvider.Options.MaxTokens != nil {
			params.MaxTokens = openai.Int(int64(*config.AiProvider.Options.MaxTokens))
		}
		if config.AiProvider.Options.Temperature != nil {
			params.Temperature = openai.Float(*config.AiProvider.Options.Temperature)
		}

		resp, err := h.aiProviderCli.ChatOpenAI(ctx, params)
		if err != nil {
			return "", fmt.Errorf("ChatOpenAI API error: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("模型返回为空")
		}

		// 检查是否需要工具调用
		if resp.Choices[0].FinishReason != "tool_calls" || len(resp.Choices[0].Message.ToolCalls) == 0 {
			// 无工具调用，返回最终结果
			return resp.Choices[0].Message.Content, nil
		}

		// 有工具调用，构建 assistant 消息
		toolCallsParam := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(resp.Choices[0].Message.ToolCalls))
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			toolCallsParam = append(toolCallsParam, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID:   tc.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				},
			})
		}

		assistantWithCalls := openai.ChatCompletionAssistantMessageParam{
			Role:      "assistant",
			ToolCalls: toolCallsParam,
		}
		hist = append(hist, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantWithCalls})

		// 执行所有工具调用
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			name := tc.Function.Name

			// 解析参数
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = map[string]any{"_parse_error": err.Error(), "_raw": tc.Function.Arguments}
			}

			// 特殊处理：自动注入 user_id
			if name == "get_todos" || name == "get_course" {
				args["user_id"] = userID
			}
			// 特殊处理：get_course 需要 term 参数
			if name == "get_course" {
				if _, ok := args["term"]; !ok {
					args["term"] = "202501" // 默认当前学期
				}
			}

			logger.Infof("DailySchedule: calling tool %s with args %v", name, args)

			// 调用 MCP 工具
			out, callErr := h.mcpCli.CallTool(ctx, name, args)
			if callErr != nil {
				out = fmt.Sprintf("tool error: %v", callErr)
				logger.Errorf("DailySchedule: tool %s error: %v", name, callErr)
			}

			// 工具结果回模型
			hist = append(hist, openai.ToolMessage(out, tc.ID))
		}

		// 继续下一轮
	}
}
