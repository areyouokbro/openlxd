package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/openlxd/backend/internal/auth"
	"github.com/openlxd/backend/internal/models"
	"gorm.io/gorm"
)

// UserAPI 用户管理API
type UserAPI struct {
	db *gorm.DB
}

// NewUserAPI 创建用户API实例
func NewUserAPI(db *gorm.DB) *UserAPI {
	return &UserAPI{db: db}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string       `json:"token"`
	User     *models.User `json:"user"`
}

// Register 用户注册
func (api *UserAPI) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证输入
	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondError(w, "Username, email and password are required", http.StatusBadRequest)
		return
	}

	// 验证密码强度
	if err := auth.ValidatePassword(req.Password); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := api.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		respondError(w, "Username already exists", http.StatusConflict)
		return
	}

	// 检查邮箱是否已存在
	if err := api.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		respondError(w, "Email already exists", http.StatusConflict)
		return
	}

	// 加密密码
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// 生成API密钥
	apiKey, err := auth.GenerateAPIKey()
	if err != nil {
		respondError(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	// 创建用户
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		APIKey:       apiKey,
		Role:         "user",
		Status:       "active",
	}

	if err := api.db.Create(&user).Error; err != nil {
		respondError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// 生成Token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		respondError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	respondSuccess(w, LoginResponse{
		Token: token,
		User:  &user,
	})
}

// Login 用户登录
func (api *UserAPI) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证输入
	if req.Username == "" || req.Password == "" {
		respondError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// 查找用户
	var user models.User
	if err := api.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		respondError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 验证密码
	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		respondError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 检查用户状态
	if !user.IsActive() {
		respondError(w, "User account is not active", http.StatusForbidden)
		return
	}

	// 生成Token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		respondError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	respondSuccess(w, LoginResponse{
		Token: token,
		User:  &user,
	})
}

// GetProfile 获取用户信息
func (api *UserAPI) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	respondSuccess(w, user)
}

// RegenerateAPIKey 重新生成API密钥
func (api *UserAPI) RegenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 生成新的API密钥
	apiKey, err := auth.GenerateAPIKey()
	if err != nil {
		respondError(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	// 更新数据库
	if err := api.db.Model(user).Update("api_key", apiKey).Error; err != nil {
		respondError(w, "Failed to update API key", http.StatusInternalServerError)
		return
	}

	user.APIKey = apiKey
	respondSuccess(w, user)
}

// ListUsers 获取用户列表（管理员）
func (api *UserAPI) ListUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := api.db.Find(&users).Error; err != nil {
		respondError(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	respondSuccess(w, users)
}

// UpdateUserStatus 更新用户状态（管理员）
func (api *UserAPI) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uint   `json:"user_id"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证状态
	validStatuses := []string{"active", "suspended", "deleted"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		respondError(w, "Invalid status", http.StatusBadRequest)
		return
	}

	// 更新用户状态
	if err := api.db.Model(&models.User{}).Where("id = ?", req.UserID).Update("status", req.Status).Error; err != nil {
		respondError(w, "Failed to update user status", http.StatusInternalServerError)
		return
	}

	respondSuccess(w, map[string]string{"message": "User status updated successfully"})
}

// UpdateUserRole 更新用户角色（管理员）
func (api *UserAPI) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uint   `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证角色
	if req.Role != "admin" && req.Role != "user" {
		respondError(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// 更新用户角色
	if err := api.db.Model(&models.User{}).Where("id = ?", req.UserID).Update("role", req.Role).Error; err != nil {
		respondError(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	respondSuccess(w, map[string]string{"message": "User role updated successfully"})
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

// respondSuccess 返回成功响应
func respondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// GetUserContainers 获取用户的容器列表
func (api *UserAPI) GetUserContainers(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var containers []models.Container
	query := api.db

	// 如果不是管理员，只显示自己的容器
	if !user.IsAdmin() {
		query = query.Where("user_id = ?", user.ID)
	}

	if err := query.Find(&containers).Error; err != nil {
		respondError(w, "Failed to fetch containers", http.StatusInternalServerError)
		return
	}

	respondSuccess(w, containers)
}

// CheckContainerOwnership 检查容器所有权
func (api *UserAPI) CheckContainerOwnership(userID, containerID uint) bool {
	var container models.Container
	if err := api.db.First(&container, containerID).Error; err != nil {
		return false
	}

	// 检查容器是否属于该用户
	return container.UserID == userID
}

// NormalizeUsername 规范化用户名（转小写，去空格）
func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
