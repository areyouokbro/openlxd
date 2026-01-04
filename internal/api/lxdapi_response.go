package api

import (
	"encoding/json"
	"net/http"
)

// LXDAPIResponse lxdapi响应格式
type LXDAPIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// RespondLXDAPISuccess 返回lxdapi成功响应
func RespondLXDAPISuccess(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LXDAPIResponse{
		Code: 200,
		Msg:  message,
		Data: data,
	})
}

// RespondLXDAPIError 返回lxdapi错误响应
func RespondLXDAPIError(w http.ResponseWriter, message string, httpCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(LXDAPIResponse{
		Code: httpCode,
		Msg:  message,
		Data: nil,
	})
}
