package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

const CtxKeyUserID = "user_id"

func Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		auth := string(c.GetHeader("Authorization"))
		if auth == "" {
			auth = string(c.GetHeader("X-Token"))
		}
		if auth == "" {
			c.AbortWithStatusJSON(401, map[string]any{"code": 401, "message": "unauthorized"})
			return
		}
		token := auth
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			token = strings.TrimSpace(auth[7:])
		}

		uid, err := parseUserIDFromToken(token)
		if err != nil || uid <= 0 {
			c.AbortWithStatusJSON(401, map[string]any{"code": 401, "message": "invalid token"})
			return
		}
		c.Set(CtxKeyUserID, uid)
		c.Next(ctx)
	}
}

// TODO: 替换为真实 token 解码逻辑
func parseUserIDFromToken(token string) (int64, error) {
	// 示例：固定返回 1
	return 1, nil
}
