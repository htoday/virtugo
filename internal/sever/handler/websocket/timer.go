package websocket

import (
	"fmt"
	"sync"
	"time"
)

type ResettableTimer struct {
	duration time.Duration
	timer    *time.Timer
	mu       sync.Mutex
	callback func()
	active   bool
}

// 构造函数
func NewResettableTimer(d time.Duration, callback func()) *ResettableTimer {
	return &ResettableTimer{
		duration: d,
		callback: callback,
	}
}

// 启动计时器（首次或停止后）
func (rt *ResettableTimer) Start() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.active {
		return // 已经在运行，无需重复启动
	}

	rt.active = true
	rt.timer = time.AfterFunc(rt.duration, rt.wrapCallback)
	fmt.Println("▶️ 计时器已启动")
}

// 重置计时器（如果未启动，会自动启动）
func (rt *ResettableTimer) Reset() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if !rt.active {
		rt.Start()
		return
	}

	if !rt.timer.Stop() {
		select {
		case <-rt.timer.C:
		default:
		}
	}
	rt.timer.Reset(rt.duration)
	fmt.Println("🔄 计时器已重置")
}

// 停止计时器
func (rt *ResettableTimer) Stop() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if !rt.active {
		return
	}

	rt.active = false
	rt.timer.Stop()
	fmt.Println("⏹️ 计时器已停止")
}

// 包装回调，确保只在 active 时执行
func (rt *ResettableTimer) wrapCallback() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if !rt.active {
		return
	}
	rt.active = false // 自动变为非活跃状态
	rt.callback()
}
