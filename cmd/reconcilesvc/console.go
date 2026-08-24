package main

import (
	"net/http"
	"os"
)

// monitorPage 读取监控页面文件。
func monitorPage() ([]byte, error) {
	return os.ReadFile("web/monitor.html")
}

// handleMonitor 返回对账监控页面。
func handleMonitor(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	page, err := monitorPage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}
