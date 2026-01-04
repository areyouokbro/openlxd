package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/openlxd/backend/internal/auth"
	"github.com/openlxd/backend/internal/lxd"
	"github.com/openlxd/backend/internal/models"
	"gorm.io/gorm"
)

// ImageAPI 镜像管理API
type ImageAPI struct {
	db        *gorm.DB
	lxdClient *lxd.ClientWrapper
}

// NewImageAPI 创建镜像API实例
func NewImageAPI(db *gorm.DB, lxdClient *lxd.ClientWrapper) *ImageAPI {
	return &ImageAPI{
		db:        db,
		lxdClient: lxdClient,
	}
}

// RemoteImage 远程镜像信息
type RemoteImage struct {
	Alias        string `json:"alias"`
	Distribution string `json:"distribution"`
	Release      string `json:"release"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant"`
	Description  string `json:"description"`
	Size         int64  `json:"size"`
}

// ImportImageRequest 导入镜像请求
type ImportImageRequest struct {
	Alias        string `json:"alias"`
	Architecture string `json:"architecture"`
}

// ListImages 获取本地镜像列表
func (api *ImageAPI) ListImages(w http.ResponseWriter, r *http.Request) {
	var images []models.Image
	if err := api.db.Find(&images).Error; err != nil {
		respondError(w, "Failed to fetch images", http.StatusInternalServerError)
		return
	}

	respondSuccess(w, images)
}

// GetRemoteImages 获取远程镜像列表
func (api *ImageAPI) GetRemoteImages(w http.ResponseWriter, r *http.Request) {
	// 预定义的常用镜像列表
	remoteImages := []RemoteImage{
		// Ubuntu
		{Alias: "ubuntu/24.04", Distribution: "ubuntu", Release: "24.04", Architecture: "amd64", Variant: "default", Description: "Ubuntu 24.04 LTS (Noble Numbat)"},
		{Alias: "ubuntu/22.04", Distribution: "ubuntu", Release: "22.04", Architecture: "amd64", Variant: "default", Description: "Ubuntu 22.04 LTS (Jammy Jellyfish)"},
		{Alias: "ubuntu/20.04", Distribution: "ubuntu", Release: "20.04", Architecture: "amd64", Variant: "default", Description: "Ubuntu 20.04 LTS (Focal Fossa)"},
		{Alias: "ubuntu/18.04", Distribution: "ubuntu", Release: "18.04", Architecture: "amd64", Variant: "default", Description: "Ubuntu 18.04 LTS (Bionic Beaver)"},
		
		// Debian
		{Alias: "debian/12", Distribution: "debian", Release: "12", Architecture: "amd64", Variant: "default", Description: "Debian 12 (Bookworm)"},
		{Alias: "debian/11", Distribution: "debian", Release: "11", Architecture: "amd64", Variant: "default", Description: "Debian 11 (Bullseye)"},
		{Alias: "debian/10", Distribution: "debian", Release: "10", Architecture: "amd64", Variant: "default", Description: "Debian 10 (Buster)"},
		
		// CentOS
		{Alias: "centos/9-Stream", Distribution: "centos", Release: "9-Stream", Architecture: "amd64", Variant: "default", Description: "CentOS 9 Stream"},
		{Alias: "centos/8-Stream", Distribution: "centos", Release: "8-Stream", Architecture: "amd64", Variant: "default", Description: "CentOS 8 Stream"},
		{Alias: "centos/7", Distribution: "centos", Release: "7", Architecture: "amd64", Variant: "default", Description: "CentOS 7"},
		
		// Alpine
		{Alias: "alpine/3.19", Distribution: "alpine", Release: "3.19", Architecture: "amd64", Variant: "default", Description: "Alpine Linux 3.19"},
		{Alias: "alpine/3.18", Distribution: "alpine", Release: "3.18", Architecture: "amd64", Variant: "default", Description: "Alpine Linux 3.18"},
		{Alias: "alpine/3.17", Distribution: "alpine", Release: "3.17", Architecture: "amd64", Variant: "default", Description: "Alpine Linux 3.17"},
		{Alias: "alpine/3.16", Distribution: "alpine", Release: "3.16", Architecture: "amd64", Variant: "default", Description: "Alpine Linux 3.16"},
		
		// Rocky Linux
		{Alias: "rockylinux/9", Distribution: "rockylinux", Release: "9", Architecture: "amd64", Variant: "default", Description: "Rocky Linux 9"},
		{Alias: "rockylinux/8", Distribution: "rockylinux", Release: "8", Architecture: "amd64", Variant: "default", Description: "Rocky Linux 8"},
		
		// Fedora
		{Alias: "fedora/40", Distribution: "fedora", Release: "40", Architecture: "amd64", Variant: "default", Description: "Fedora 40"},
		{Alias: "fedora/39", Distribution: "fedora", Release: "39", Architecture: "amd64", Variant: "default", Description: "Fedora 39"},
		{Alias: "fedora/38", Distribution: "fedora", Release: "38", Architecture: "amd64", Variant: "default", Description: "Fedora 38"},
		
		// Arch Linux
		{Alias: "archlinux/current", Distribution: "archlinux", Release: "current", Architecture: "amd64", Variant: "default", Description: "Arch Linux (Rolling)"},
		
		// Oracle Linux
		{Alias: "oracle/9", Distribution: "oracle", Release: "9", Architecture: "amd64", Variant: "default", Description: "Oracle Linux 9"},
		{Alias: "oracle/8", Distribution: "oracle", Release: "8", Architecture: "amd64", Variant: "default", Description: "Oracle Linux 8"},
	}

	respondSuccess(w, remoteImages)
}

// ImportImage 导入镜像
func (api *ImageAPI) ImportImage(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 只有管理员可以导入镜像
	if !user.IsAdmin() {
		respondError(w, "Admin access required", http.StatusForbidden)
		return
	}

	var req ImportImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Alias == "" {
		respondError(w, "Image alias is required", http.StatusBadRequest)
		return
	}

	if req.Architecture == "" {
		req.Architecture = "amd64"
	}

	// 检查镜像是否已存在
	var existingImage models.Image
	if err := api.db.Where("alias = ?", req.Alias).First(&existingImage).Error; err == nil {
		if existingImage.IsImported() {
			respondError(w, "Image already imported", http.StatusConflict)
			return
		}
	}

	// 解析镜像别名
	distribution, release := parseImageAlias(req.Alias)

	// 创建镜像记录
	image := models.Image{
		Alias:        req.Alias,
		Distribution: distribution,
		Release:      release,
		Architecture: req.Architecture,
		Variant:      "default",
		Status:       "downloading",
	}

	if err := api.db.Create(&image).Error; err != nil {
		respondError(w, "Failed to create image record", http.StatusInternalServerError)
		return
	}

	// 异步导入镜像
	go func() {
		// 从 images.linuxcontainers.org 导入镜像
		err := api.lxdClient.ImportImage(req.Alias, req.Architecture)
		
		now := time.Now()
		if err != nil {
			// 导入失败
			api.db.Model(&image).Updates(map[string]interface{}{
				"status":      "failed",
				"description": fmt.Sprintf("Import failed: %v", err),
			})
		} else {
			// 导入成功
			api.db.Model(&image).Updates(map[string]interface{}{
				"status":      "imported",
				"imported_at": &now,
			})
		}
	}()

	respondSuccess(w, map[string]interface{}{
		"image_id": image.ID,
		"alias":    image.Alias,
		"status":   image.Status,
		"message":  "Image import started",
	})
}

// DeleteImage 删除镜像
func (api *ImageAPI) DeleteImage(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 只有管理员可以删除镜像
	if !user.IsAdmin() {
		respondError(w, "Admin access required", http.StatusForbidden)
		return
	}

	alias := r.URL.Query().Get("alias")
	if alias == "" {
		respondError(w, "Image alias is required", http.StatusBadRequest)
		return
	}

	// 查找镜像
	var image models.Image
	if err := api.db.Where("alias = ?", alias).First(&image).Error; err != nil {
		respondError(w, "Image not found", http.StatusNotFound)
		return
	}

	// 删除LXD镜像
	if err := api.lxdClient.DeleteImage(image.Fingerprint); err != nil {
		respondError(w, fmt.Sprintf("Failed to delete image from LXD: %v", err), http.StatusInternalServerError)
		return
	}

	// 从数据库删除
	api.db.Delete(&image)

	respondSuccess(w, map[string]string{
		"message": "Image deleted successfully",
	})
}

// GetImageInfo 获取镜像信息
func (api *ImageAPI) GetImageInfo(w http.ResponseWriter, r *http.Request) {
	alias := r.URL.Query().Get("alias")
	if alias == "" {
		respondError(w, "Image alias is required", http.StatusBadRequest)
		return
	}

	var image models.Image
	if err := api.db.Where("alias = ?", alias).First(&image).Error; err != nil {
		respondError(w, "Image not found", http.StatusNotFound)
		return
	}

	respondSuccess(w, image)
}

// SyncImages 同步LXD镜像到数据库
func (api *ImageAPI) SyncImages(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 只有管理员可以同步镜像
	if !user.IsAdmin() {
		respondError(w, "Admin access required", http.StatusForbidden)
		return
	}

	// 获取LXD镜像列表
	lxdImages, err := api.lxdClient.ListImages()
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to list LXD images: %v", err), http.StatusInternalServerError)
		return
	}

	syncCount := 0
	for _, lxdImage := range lxdImages {
		// 检查镜像是否已存在
		var existingImage models.Image
		if err := api.db.Where("fingerprint = ?", lxdImage.Fingerprint).First(&existingImage).Error; err != nil {
			// 镜像不存在，创建新记录
			alias := lxdImage.Alias
			if alias == "" {
				alias = lxdImage.Fingerprint[:12]
			}

			distribution, release := parseImageAlias(alias)
			now := time.Now()

			image := models.Image{
				Alias:        alias,
				Fingerprint:  lxdImage.Fingerprint,
				Distribution: distribution,
				Release:      release,
				Architecture: lxdImage.Architecture,
				Description:  lxdImage.Description,
				Size:         lxdImage.Size,
				Status:       "imported",
				ImportedAt:   &now,
			}

			if err := api.db.Create(&image).Error; err == nil {
				syncCount++
			}
		}
	}

	respondSuccess(w, map[string]interface{}{
		"synced": syncCount,
		"total":  len(lxdImages),
		"message": fmt.Sprintf("Synced %d images", syncCount),
	})
}

// parseImageAlias 解析镜像别名
func parseImageAlias(alias string) (distribution, release string) {
	// 简单的解析逻辑
	// 格式: distribution/release
	for i, c := range alias {
		if c == '/' {
			distribution = alias[:i]
			release = alias[i+1:]
			return
		}
	}
	
	// 如果没有找到 /，则整个字符串作为 distribution
	distribution = alias
	release = "latest"
	return
}
