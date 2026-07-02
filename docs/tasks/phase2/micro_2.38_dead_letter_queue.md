# Micro-Task 2.38: Tạo kernel/eventbus/dlq.go (Event Bus Dead Letter Queue)

## Thông tin
- **File tạo**: `kernel/eventbus/dlq.go`
- **File cập nhật**: `kernel/eventbus/subscriber.go` (Cập nhật hàm safeHandler để đẩy vào DLQ khi panic)
- **Package**: `eventbus`
- **Dependencies trước**: 2.12 (types.go), 2.14 (subscriber.go), 2.15 (bus.go)
- **Thời gian**: 20 phút
- **Verify**: `go build ./kernel/eventbus/...`

## Purpose
Triển khai bộ đệm thư chết (Dead Letter Queue - DLQ) cho Event Bus. Khi một subscriber handler gặp lỗi crash (panic), Event Bus sẽ bắt lỗi (recover), ghi nhận thông tin chi tiết lỗi cùng payload sự kiện gốc vào DLQ dưới dạng một ring-buffer hữu hạn. Điều này giúp đội vận hành phân tích lỗi không bị mất dấu sự kiện.

## EXACT code to create

### Phần 1: Tạo `kernel/eventbus/dlq.go`

```go
package eventbus

import (
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/event"
)

// DeadLetterEntry represents a failed event execution log.
type DeadLetterEntry struct {
	Event     event.Event `json:"event"`
	Error     string      `json:"error"`
	Timestamp time.Time   `json:"timestamp"`
}

// DeadLetterQueue is a thread-safe circular buffer (ring buffer) for failed events.
// It stores up to MaxSize failures, discarding oldest entries when full.
type DeadLetterQueue struct {
	mu       sync.RWMutex
	entries  []DeadLetterEntry
	maxSize  int
	writeIdx int
	count    int
}

// NewDeadLetterQueue creates a new DLQ with the specified capacity.
func NewDeadLetterQueue(maxSize int) *DeadLetterQueue {
	if maxSize <= 0 {
		maxSize = 100 // Default capacity
	}
	return &DeadLetterQueue{
		entries: make([]DeadLetterEntry, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a new failure entry into the circular buffer.
func (q *DeadLetterQueue) Add(evt event.Event, errStr string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.entries[q.writeIdx] = DeadLetterEntry{
		Event:     evt,
		Error:     errStr,
		Timestamp: time.Now(),
	}

	q.writeIdx = (q.writeIdx + 1) % q.maxSize
	if q.count < q.maxSize {
		q.count++
	}
}

// Entries returns a copy of all active entries in chronological order (oldest first).
func (q *DeadLetterQueue) Entries() []DeadLetterEntry {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]DeadLetterEntry, q.count)
	if q.count == 0 {
		return result
	}

	// Calculate starting index of the oldest entry in circular buffer
	startIdx := 0
	if q.count == q.maxSize {
		startIdx = q.writeIdx
	}

	for i := 0; i < q.count; i++ {
		readIdx := (startIdx + i) % q.maxSize
		result[i] = q.entries[readIdx]
	}

	return result
}

// Len returns the current number of failures in the queue.
func (q *DeadLetterQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.count
}

// Clear resets the queue.
func (q *DeadLetterQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.entries = make([]DeadLetterEntry, q.maxSize)
	q.writeIdx = 0
	q.count = 0
}
```

---

### Phần 2: Cập nhật `kernel/eventbus/subscriber.go` và `bus.go`

Cập nhật `safeHandler` nhận thêm con trỏ `DeadLetterQueue` để tự động ghi nhận lỗi:

```go
// Cập nhật safeHandler trong kernel/eventbus/subscriber.go
func safeHandler(handler func(event.Event), dlq *DeadLetterQueue, logger *slog.Logger) func(event.Event) {
	return func(evt event.Event) {
		defer func() {
			if r := recover(); r != nil {
				errStr := fmt.Sprintf("%v", r)
				if logger != nil {
					logger.Error("event handler panicked",
						"event_type", evt.Type,
						"event_source", evt.Source,
						"panic", errStr,
					)
				}
				// Ghi nhận lỗi vào Dead Letter Queue
				if dlq != nil {
					dlq.Add(evt, errStr)
				}
			}
		}()
		handler(evt)
	}
}
```

> **Lưu ý**: Cập nhật struct `Bus` trong `kernel/eventbus/bus.go` để tích hợp `DeadLetterQueue`:

```go
// Cập nhật struct Bus trong kernel/eventbus/bus.go
type Bus struct {
	subscribers *subscriberMap
	logger      *slog.Logger
	dlq         *DeadLetterQueue // Thêm trường dlq
	wg          sync.WaitGroup
	closed      bool
	mu          sync.RWMutex
}

// Cập nhật hàm khởi tạo New() để tự động init DLQ:
func New(logger *slog.Logger) *Bus {
	return &Bus{
		subscribers: newSubscriberMap(),
		logger:      logger,
		dlq:         NewDeadLetterQueue(100), // Mặc định chứa 100 lỗi
	}
}

// Cập nhật Subscribe() để truyền dlq vào safeHandler:
func (b *Bus) Subscribe(pattern string, handler func(event.Event)) (func(), error) {
	// ... validation ...
	safe := safeHandler(handler, b.dlq, b.logger) // Truyền b.dlq
	sub := b.subscribers.add(pattern, safe)
	// ...
}

// Thêm phương thức getter cho DLQ:
func (b *Bus) DLQ() *DeadLetterQueue {
	return b.dlq
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Rò rỉ bộ nhớ từ Dead Letter Queue (Unbounded Queue Growth)
```go
// ❌ SAI:
type DeadLetterQueue struct {
    entries []DeadLetterEntry // Không giới hạn kích thước -> append liên tục gây OOM trong long-running.
}
```
DLQ bắt buộc phải dùng cơ chế **Circular Buffer** hoặc giới hạn dung lượng (`maxSize`). Khi đạt giới hạn, các lỗi cũ nhất sẽ bị đẩy ra ngoài để bảo vệ hệ thống khỏi tràn bộ nhớ.

### Pitfall 2: Copy nhầm trật tự ghi (Circular indexing calculation)
Khi trích xuất phần tử từ ring buffer, trật tự ghi tăng dần không trùng với index mảng nếu buffer đã bị lặp vòng. Bắt buộc phải tính toán chỉ mục bắt đầu `startIdx = writeIdx` khi `count == maxSize` để đảm bảo kết quả xuất ra theo thứ tự thời gian cũ nhất đến mới nhất.

## Checklist
- [ ] File `kernel/eventbus/dlq.go` tồn tại
- [ ] Tích hợp `DeadLetterQueue` vào struct `Bus` của `kernel/eventbus/bus.go`
- [ ] `safeHandler` ghi nhận panic vào DLQ
- [ ] `DeadLetterQueue` được giới hạn kích thước tối đa (circular ring buffer)
- [ ] Hàm `Entries()` trả về bản sao mảng theo trật tự thời gian chính xác
- [ ] Phương thức `DLQ()` có getter truy cập
- [ ] `go build ./kernel/eventbus/...` không lỗi
