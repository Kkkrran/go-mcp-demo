package application

import (
	"context"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v2"
)

func WithAIScienceAndEngineeringBuildHtmlTool() tool_set.Option {
	return func(toolSet *tool_set.ToolSet) {
		newTool := mcp.NewTool(
			"build_html_to_solve_science_and_engineering_problem",
			mcp.WithDescription("当用户遇到学习问题上的困难时，通过构造可交互的网页来帮助用户理解数学概念和绘制图像, 可以在调用后加上对问题的辅助解析"),
			mcp.WithString("question", mcp.Required(), mcp.Description("用户提出的科学或工程相关的问题")),
		)
		toolSet.Tools = append(toolSet.Tools, &newTool)
		toolSet.HandlerFunc[newTool.Name] = AIScienceAndEngineeringBuildHtml
	}
}

func AIScienceAndEngineeringBuildHtml(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	rawQuestion, ok := args["question"]
	if !ok {
		return mcp.NewToolResultError("missing required arg: question"), nil
	}

	question, ok := rawQuestion.(string)
	if !ok || question == "" {
		return mcp.NewToolResultError("question must be a non-empty string"), nil
	}

	// 从 global ClientSet 获取 AI provider
	clientSet := base.GetGlobalClientSet()
	if clientSet == nil || clientSet.AiProviderCli == nil {
		return mcp.NewToolResultError("AI provider not initialized"), nil
	}

	resp, err := clientSet.AiProviderCli.ChatOpenAI(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(config.AiProvider.Model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPromptHTMLPrinter),
			openai.UserMessage(question),
		},
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(resp.Choices[0].Message.Content), nil
}

const systemPromptHTMLPrinter = `
你是一名叫ssibal的html生成小助手：
1. 你只能使用<htmath>标签来展示内容，除此之外不会进行任何输出。
2. 使用<htmath>标签渲染HTML内容，特别适合数学图形和函数可视化，格式为<htmath>HTML代码</htmath>
3. 你可以正常使用Markdown格式化文本，也可以使用MathJax展示数学公式。在html图像中尽量使用中文进行展示
4. 在讲解数学知识时，你会充分利用你的<htmath>能力来帮助用户更好地理解各种概念。

示例:
- 当用户想要可视化sin(x)曲线，可以回复：<htmath><html><div id="plot"></div>
<script src="https://cdn.plot.ly/plotly-2.30.0.min.js"></script>
<script type="text/javascript">
document.addEventListener('DOMContentLoaded', function() {
  setTimeout(function() {
    try {
      const plotDiv = document.getElementById('plot');
      if(plotDiv && window.Plotly) {
        Plotly.newPlot(plotDiv, [{
          x: Array.from({length: 100}, (_, i) => i * 0.1),
          y: Array.from({length: 100}, (_, i) => Math.sin(i * 0.1)),
          type: 'scatter'
        }]);
      } else {
        console.error('Plot div not found or Plotly not loaded');
      }
    } catch(e) {
      console.error('Error creating plot:', e);
    }
  }, 500);
});
</script></html></htmath>

请根据这些特殊格式回应用户。
`
