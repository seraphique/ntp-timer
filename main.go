package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/beevik/ntp"
	"github.com/gorilla/mux"
)

//go:embed static/*
var staticFiles embed.FS

type NTPResult struct {
	NTPTime     string `json:"ntpTime"`
	LocalTime   string `json:"localTime"`
	Offset      int64  `json:"offset"`      // 毫秒
	Delay       int64  `json:"delay"`       // 毫秒
	Success     bool   `json:"success"`
	Error       string `json:"error"`
	ServerURL   string `json:"serverUrl"`
}

type SyncSystemTimeResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func setWindowsSystemTime(t time.Time) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("system time sync not supported on %s, use 'sudo date' manually", runtime.GOOS)
	}
	
	// 使用Windows API直接设置系统时间
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setSystemTime := kernel32.NewProc("SetSystemTime")
	
	// 转换为UTC时间
	utc := t.UTC()
	
	// Windows SYSTEMTIME 结构
	systemTime := struct {
		Year         uint16
		Month        uint16
		DayOfWeek    uint16
		Day          uint16
		Hour         uint16
		Minute       uint16
		Second       uint16
		Milliseconds uint16
	}{
		Year:         uint16(utc.Year()),
		Month:        uint16(utc.Month()),
		DayOfWeek:    uint16(utc.Weekday()),
		Day:          uint16(utc.Day()),
		Hour:         uint16(utc.Hour()),
		Minute:       uint16(utc.Minute()),
		Second:       uint16(utc.Second()),
		Milliseconds: uint16(utc.Nanosecond() / 1000000),
	}
	
	ret, _, _ := setSystemTime.Call(uintptr(unsafe.Pointer(&systemTime)))
	if ret == 0 {
		return fmt.Errorf("没有管理员权限，请以管理员身份运行程序")
	}
	
	return nil
}

func main() {
	r := mux.NewRouter()
	
	// 添加路由调试中间件
	r.Use(loggingMiddleware)
	
	// API子路由
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/sync", handleNTPSync).Methods("POST", "OPTIONS")
	api.HandleFunc("/sync-system", handleSyncSystemTime).Methods("POST", "OPTIONS")
	api.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"API working","timestamp":"` + time.Now().Format("2006-01-02 15:04:05") + `"}`))
	}).Methods("GET", "POST", "OPTIONS")
	
	// 静态文件服务（使用嵌入的文件）
	staticFS, _ := fs.Sub(staticFiles, "static")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	
	// 主页路由
	r.HandleFunc("/", handleIndex).Methods("GET")
	
	// 添加通用OPTIONS处理
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		w.WriteHeader(http.StatusOK)
	})
	
	port := ":8080"
	
	// 自动打开浏览器
	go openBrowser("http://localhost" + port)
	
	log.Fatal(http.ListenAndServe(port, r))
}

// 静默中间件（不输出日志）
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func handleNTPSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	var request struct {
		ServerURL string `json:"serverUrl"`
	}
	
	if r.Body != nil {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			result := NTPResult{
				Success: false,
				Error:   "Invalid request format",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(result)
			return
		}
	}
	
	serverURL := request.ServerURL
	if serverURL == "" {
		serverURL = "time.cloud.tencent.com"
	}
	
	result := syncNTP(serverURL)
	
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		w.Write([]byte(`{"success":false,"error":"JSON encoding failed"}`))
	}
}

func syncNTP(serverURL string) NTPResult {
	startTime := time.Now()
	
	response, err := ntp.Query(serverURL)
	if err != nil {
		return NTPResult{
			Success:   false,
			Error:     fmt.Sprintf("连接失败: %v", err),
			ServerURL: serverURL,
		}
	}
	
	delay := time.Since(startTime).Milliseconds()
	
	ntpTime := time.Now().Add(response.ClockOffset)
	localTime := time.Now()
	
	return NTPResult{
		NTPTime:   ntpTime.Format("15:04:05.000"),
		LocalTime: localTime.Format("15:04:05.000"),
		Offset:    response.ClockOffset.Milliseconds(),
		Delay:     delay,
		Success:   true,
		ServerURL: serverURL,
	}
}

func handleSyncSystemTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	var request struct {
		ServerURL string `json:"serverUrl"`
	}
	
	// 更安全的JSON解析
	if r.Body != nil {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			result := SyncSystemTimeResult{
				Success: false,
				Error:   "Invalid request format",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(result)
			return
		}
	}
	
	serverURL := request.ServerURL
	if serverURL == "" {
		serverURL = "time.cloud.tencent.com"
	}
	
	result := syncSystemTime(serverURL)
	
	// 确保返回正确的JSON
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		// 如果JSON编码失败，返回简单的错误信息
		w.Write([]byte(`{"success":false,"error":"JSON encoding failed"}`))
	}
}

func syncSystemTime(serverURL string) SyncSystemTimeResult {
	// 获取NTP时间
	response, err := ntp.Query(serverURL)
	if err != nil {
		return SyncSystemTimeResult{
			Success: false,
			Error:   fmt.Sprintf("NTP同步失败: %v", err),
		}
	}
	
	// 计算准确的NTP时间
	ntpTime := time.Now().Add(response.ClockOffset)
	
	// 设置系统时间
	if runtime.GOOS == "windows" {
		err = setWindowsSystemTime(ntpTime)
		if err != nil {
			return SyncSystemTimeResult{
				Success: false,
				Error:   fmt.Sprintf("设置系统时间失败: %v", err),
			}
		}
	} else {
		// 非Windows系统使用date命令
		err = setUnixSystemTime(ntpTime)
		if err != nil {
			return SyncSystemTimeResult{
				Success: false,
				Error:   fmt.Sprintf("设置系统时间失败: %v", err),
			}
		}
	}
	
	return SyncSystemTimeResult{
		Success: true,
		Message: fmt.Sprintf("系统时间已同步到NTP时间: %s", ntpTime.Format("2006-01-02 15:04:05")),
	}
}


func setUnixSystemTime(t time.Time) error {
	cmd := exec.Command("sudo", "date", "-s", t.Format("2006-01-02 15:04:05"))
	return cmd.Run()
}

func openBrowser(url string) {
	time.Sleep(1 * time.Second) // 等待服务器启动
	
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	
	// 静默运行，不输出错误信息
	_ = err
}
