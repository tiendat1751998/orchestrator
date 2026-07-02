# Micro-Task 2.39: Tạo kernel/config/watcher.go (Config Hot-Reload)

## Thông tin
- **File tạo**: `kernel/config/watcher.go`
- **Package**: `config`
- **Dependencies trước**: 2.04 (loader.go), 2.05 (validator.go)
- **Thời gian**: 20 phút
- **Verify**: `go build ./kernel/config/...`

## Purpose
Triển khai cơ chế nạp lại cấu hình động (Hot-Reload) mà không cần khởi động lại toàn bộ kernel. Hệ thống sử dụng một trình theo dõi tệp tin (Config Watcher) hoạt động dựa trên cơ chế thăm dò (polling ModTime) cực kỳ đáng tin cậy trên đa nền tảng (cross-platform), tự động tải lại cấu hình, chạy validator và gọi callback khi phát hiện tệp tin YAML bị thay đổi.

## EXACT code to create

```go
package config

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"
)

// Watcher monitors a configuration file for changes and triggers reload callback.
// Uses polling based on file modification time (ModTime) for cross-platform reliability.
type Watcher struct {
	filePath string
	interval time.Duration
	onChange func(*Config)
	
	mu       sync.Mutex
	lastMod  time.Time
	running  bool
	stopChan chan struct{}
}

// NewWatcher creates a new Config Watcher.
//
// Parameters:
//   - path: Absolute path to the YAML config file.
//   - interval: How frequently to check the file for changes (default: 5s if <= 0).
//   - onChange: Callback function called with the new configuration.
func NewWatcher(path string, interval time.Duration, onChange func(*Config)) *Watcher {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &Watcher{
		filePath: path,
		interval: interval,
		onChange: onChange,
		stopChan: make(chan struct{}),
	}
}

// Start begins monitoring the config file in a background goroutine.
// Blocks until initial validation succeeds, then runs monitor loop in background.
func (w *Watcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return errors.New("config: watcher is already running")
	}

	// Fetch initial modification time
	info, err := os.Stat(w.filePath)
	if err != nil {
		w.mu.Unlock()
		return err
	}
	w.lastMod = info.ModTime()
	w.running = true
	w.mu.Unlock()

	// Spawn background polling loop
	go w.pollLoop()

	return nil
}

// pollLoop runs the polling checks at configured intervals.
func (w *Watcher) pollLoop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.checkFile()
		}
	}
}

// checkFile inspects the file modification time and triggers reload if updated.
func (w *Watcher) checkFile() {
	info, err := os.Stat(w.filePath)
	if err != nil {
		// File might be temporarily locked or deleted during edit.
		// Ignore temporary read errors to avoid crashing.
		return
	}

	w.mu.Lock()
	modTime := info.ModTime()
	if !modTime.After(w.lastMod) {
		w.mu.Unlock()
		return
	}
	w.lastMod = modTime
	w.mu.Unlock()

	// File changed. Load new config.
	newCfg, err := Load(w.filePath)
	if err != nil {
		// If load/validation fails, ignore and keep old config (robustness).
		// We do not want a broken config edit to crash a running kernel.
		return
	}

	// Trigger callback
	if w.onChange != nil {
		w.onChange(newCfg)
	}
}

// Stop halts the configuration monitoring.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	close(w.stopChan)
	w.running = false
	// Recreate channel for potential restart
	w.stopChan = make(chan struct{})
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Panic hoặc sập Kernel khi người dùng sửa cấu hình sai cú pháp
```go
// ❌ SAI:
newCfg, err := Load(w.filePath)
if err != nil {
    panic(err) // Crash toàn bộ hệ thống đang chạy khi người dùng gõ sai 1 ký tự YAML.
}

// ✅ ĐÚNG:
newCfg, err := Load(w.filePath)
if err != nil {
    return // Bỏ qua bản cấu hình lỗi, tiếp tục sử dụng bản cấu hình cũ đã nạp sẵn trong bộ nhớ.
}
```
Khi hot-reload, sự an toàn của kernel đang chạy là ưu tiên số một. Mọi cấu hình lỗi phát sinh khi người dùng đang lưu tệp tin dở dang PHẢI được bỏ qua một cách im lặng để bảo vệ tính sẵn sàng của hệ thống.

### Pitfall 2: Race condition khi cập nhật cấu hình
Biến `Config` toàn cục không được phép ghi trực tiếp. Lớp chứa cấu hình trong Kernel phải sử dụng cơ chế bảo vệ bằng mutex hoặc atomic pointer khi cập nhật bản sao cấu hình mới để tránh dữ liệu bị phân mảnh (race condition) khi các workers đang đọc cấu hình đồng thời.

## Checklist
- [ ] File `kernel/config/watcher.go` tồn tại
- [ ] Định nghĩa `Watcher` struct với mutex bảo vệ trạng thái `running` và `lastMod`
- [ ] Cơ chế thăm dò sử dụng `os.Stat(filePath).ModTime()`
- [ ] Bỏ qua lỗi YAML cú pháp lỗi (Load fail) khi reload để bảo vệ runtime
- [ ] Dùng `time.NewTicker` kết hợp `select` để thoát khỏi vòng lặp nhanh chóng khi dừng
- [ ] Phương thức `Stop()` an toàn khi gọi nhiều lần (idempotent)
- [ ] Tái tạo `stopChan` khi dừng để phục vụ khả năng khởi động lại
- [ ] `go build ./kernel/config/...` không lỗi
