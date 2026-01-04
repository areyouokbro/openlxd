package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/openlxd/backend/internal/models"
	"gorm.io/gorm"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从请求头获取Token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// 解析Bearer Token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// 解析Token
			claims, err := ParseToken(tokenString)
			if err != nil {
				respondError(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// 从数据库获取用户信息
			var user models.User
			if err := db.First(&user, claims.UserID).Error; err != nil {
				respondError(w, "User not found", http.StatusUnauthorized)
				return
			}

			// 检查用户状态
			if !user.IsActive() {
				respondError(w, "User account is not active", http.StatusForbidden)
				return
			}

			// 将用户信息存入上下文
			ctx := context.WithValue(r.Context(), UserContextKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

	// APIKeyMiddleware API Key认证中间件（兼容lxdapi的X-API-Hash）
	func APIKeyMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 从请求头获取API Key（兼容X-API-Key和X-API-Hash）
				apiKey := r.Header.Get("X-API-Key")
				if apiKey == "" {
					apiKey = r.Header.Get("X-API-Hash")
				}
				if apiKey == "" {
					respondError(w, "Missing API key", http.StatusUnauthorized)
					return
				}

			// 从数据库查找用户
			var user models.User
			if err := db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
				respondError(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// 检查用户状态
			if !user.IsActive() {
				respondError(w, "User account is not active", http.StatusForbidden)
				return
			}

			// 将用户信息存入上下文
			ctx := context.WithValue(r.Context(), UserContextKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			respondError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin() {
			respondError(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// respondError 返回错误响应
func respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
