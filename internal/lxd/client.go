package lxd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

var (
	httpClient *http.Client
	socketPath string
	isAvailable bool
)

// InitLXD 初始化 LXD 客户端连接
func InitLXD(socket string) error {
	if socket == "" {
		socket = "/var/snap/lxd/common/lxd/unix.socket"
	}
	socketPath = socket
	
	// 创建 Unix Socket HTTP 客户端
	httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: 30 * time.Second,
	}
	
	// 测试连接
	resp, err := httpClient.Get("http://unix/1.0")
	if err != nil {
		log.Printf("警告: 无法连接到 LXD (%s): %v", socketPath, err)
		log.Println("后端将以 Mock 模式运行，API 接口可用但不会真实操作容器")
		isAvailable = false
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if metadata, ok := result["metadata"].(map[string]interface{}); ok {
			log.Printf("LXD 连接成功: %v (API: %v)", 
				metadata["environment"].(map[string]interface{})["server_name"],
				metadata["api_version"])
		}
		isAvailable = true
	} else {
		log.Printf("警告: LXD API 返回错误状态: %d", resp.StatusCode)
		isAvailable = false
	}
	
	return nil
}

// IsLXDAvailable 检查 LXD 是否可用
func IsLXDAvailable() bool {
	return isAvailable
}

// lxdRequest 发送 LXD API 请求
func lxdRequest(method, path string, body interface{}) (map[string]interface{}, error) {
	if !isAvailable {
		return nil, fmt.Errorf("LXD not available")
	}
	
	var reqBody io.Reader
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequest(method, "http://unix"+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// 检查操作状态
	if statusCode, ok := result["status_code"].(float64); ok && statusCode != 200 {
		return nil, fmt.Errorf("LXD API error: %v", result["error"])
	}
	
	// 如果是异步操作，等待完成
	if result["type"] == "async" {
		opID := result["operation"].(string)
		return waitOperation(opID)
	}
	
	return result, nil
}

// waitOperation 等待异步操作完成
func waitOperation(opID string) (map[string]interface{}, error) {
	for i := 0; i < 60; i++ {
		resp, err := httpClient.Get("http://unix" + opID)
		if err != nil {
			return nil, err
		}
		
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		
		if metadata, ok := result["metadata"].(map[string]interface{}); ok {
			if status, ok := metadata["status"].(string); ok {
				if status == "Success" {
					return result, nil
				} else if status == "Failure" {
					return nil, fmt.Errorf("operation failed: %v", metadata["err"])
				}
			}
		}
		
		time.Sleep(1 * time.Second)
	}
	
	return nil, fmt.Errorf("operation timeout")
}

// ListContainers 列出所有容器
func ListContainers() ([]string, error) {
	result, err := lxdRequest("GET", "/1.0/instances?recursion=1", nil)
	if err != nil {
		return nil, err
	}
	
	var containers []string
	if metadata, ok := result["metadata"].([]interface{}); ok {
		for _, item := range metadata {
			if inst, ok := item.(map[string]interface{}); ok {
				containers = append(containers, inst["name"].(string))
			}
		}
	}
	
	return containers, nil
}

// GetContainer 获取容器信息
func GetContainer(name string) (map[string]interface{}, error) {
	result, err := lxdRequest("GET", "/1.0/instances/"+name, nil)
	if err != nil {
		return nil, err
	}
	
	if metadata, ok := result["metadata"].(map[string]interface{}); ok {
		return metadata, nil
	}
	
	return nil, fmt.Errorf("invalid response")
}

// GetContainerState 获取容器状态
func GetContainerState(name string) (map[string]interface{}, error) {
	result, err := lxdRequest("GET", "/1.0/instances/"+name+"/state", nil)
	if err != nil {
		return nil, err
	}
	
	if metadata, ok := result["metadata"].(map[string]interface{}); ok {
		return metadata, nil
	}
	
	return nil, fmt.Errorf("invalid response")
}
