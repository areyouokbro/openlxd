package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/openlxd/backend/internal/migration"
	"github.com/openlxd/backend/internal/models"
)

// HandleCreateMigrationTask 创建迁移任务
func HandleCreateMigrationTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContainerName  string `json:"container_name"`
		SourceHost     string `json:"source_host"`
		TargetHost     string `json:"target_host"`
		MigrationType  string `json:"migration_type"` // live, cold
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	// 验证参数
	if req.ContainerName == "" || req.TargetHost == "" {
		respondJSON(w, 400, "容器名称和目标主机不能为空", nil)
		return
	}

	if req.SourceHost == "" {
		req.SourceHost = "local"
	}

	if req.MigrationType == "" {
		req.MigrationType = "cold"
	}

	// 创建迁移任务
	task, err := migration.CreateMigrationTask(req.ContainerName, req.SourceHost, req.TargetHost, req.MigrationType)
	if err != nil {
		respondJSON(w, 500, err.Error(), nil)
		return
	}

	// 异步执行迁移
	go func() {
		migration.ExecuteMigration(task.ID)
	}()

	respondJSON(w, 200, "迁移任务已创建", task)
}

// HandleGetMigrationTasks 获取迁移任务列表
func HandleGetMigrationTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := migration.GetMigrationTasks()
	if err != nil {
		respondJSON(w, 500, err.Error(), nil)
		return
	}

	respondJSON(w, 200, "获取成功", tasks)
}

// HandleGetMigrationTask 获取单个迁移任务
func HandleGetMigrationTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("id")
	if taskIDStr == "" {
		respondJSON(w, 400, "任务ID不能为空", nil)
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		respondJSON(w, 400, "任务ID格式错误", nil)
		return
	}

	task, err := migration.GetMigrationTask(uint(taskID))
	if err != nil {
		respondJSON(w, 404, "任务不存在", nil)
		return
	}

	respondJSON(w, 200, "获取成功", task)
}

// HandleGetMigrationLogs 获取迁移日志
func HandleGetMigrationLogs(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		respondJSON(w, 400, "任务ID不能为空", nil)
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		respondJSON(w, 400, "任务ID格式错误", nil)
		return
	}

	logs, err := migration.GetMigrationLogs(uint(taskID))
	if err != nil {
		respondJSON(w, 500, err.Error(), nil)
		return
	}

	respondJSON(w, 200, "获取成功", logs)
}

// HandleCancelMigration 取消迁移
func HandleCancelMigration(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID uint `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	err := migration.CancelMigration(req.TaskID)
	if err != nil {
		respondJSON(w, 500, err.Error(), nil)
		return
	}

	respondJSON(w, 200, "任务已取消", nil)
}

// HandleRollbackMigration 回滚迁移
func HandleRollbackMigration(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID uint `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	err := migration.RollbackMigration(req.TaskID)
	if err != nil {
		respondJSON(w, 500, err.Error(), nil)
		return
	}

	respondJSON(w, 200, "迁移已回滚", nil)
}

// HandleCreateRemoteHost 创建远程主机配置
func HandleCreateRemoteHost(w http.ResponseWriter, r *http.Request) {
	var host models.RemoteHost

	if err := json.NewDecoder(r.Body).Decode(&host); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	// 验证参数
	if host.Name == "" || host.Address == "" {
		respondJSON(w, 400, "主机名称和地址不能为空", nil)
		return
	}

	if host.Port == 0 {
		host.Port = 8443
	}

	if host.Protocol == "" {
		host.Protocol = "https"
	}

	// 创建主机配置
	if err := models.DB.Create(&host).Error; err != nil {
		respondJSON(w, 500, "创建失败: " + err.Error(), nil)
		return
	}

	respondJSON(w, 200, "创建成功", host)
}

// HandleGetRemoteHosts 获取远程主机列表
func HandleGetRemoteHosts(w http.ResponseWriter, r *http.Request) {
	var hosts []models.RemoteHost
	if err := models.DB.Find(&hosts).Error; err != nil {
		respondJSON(w, 500, err.Error(), nil)
		return
	}

	respondJSON(w, 200, "获取成功", hosts)
}

// HandleDeleteRemoteHost 删除远程主机配置
func HandleDeleteRemoteHost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID uint `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, "请求参数错误", nil)
		return
	}

	if err := models.DB.Delete(&models.RemoteHost{}, req.ID).Error; err != nil {
		respondJSON(w, 500, "删除失败: " + err.Error(), nil)
		return
	}

	respondJSON(w, 200, "删除成功", nil)
}
