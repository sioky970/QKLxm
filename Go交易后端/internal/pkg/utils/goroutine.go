package utils

import (
	"fmt"
	"runtime/debug"

	"exchange-go/internal/pkg/logger"
)

// SafeGo 安全启动goroutine，自动捕获panic
// 用法: utils.SafeGo(func() { ... })
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Goroutine panic recovered: %v\nStack trace:\n%s", r, string(debug.Stack()))
			}
		}()
		fn()
	}()
}

// SafeGoWithName 安全启动goroutine并指定名称（用于日志识别）
// 用法: utils.SafeGoWithName("broadcast", func() { ... })
func SafeGoWithName(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("[%s] Goroutine panic recovered: %v\nStack trace:\n%s", name, r, string(debug.Stack()))
			}
		}()
		fn()
	}()
}

// SafeGoWithRecover 安全启动goroutine并自定义panic处理
// 用法: utils.SafeGoWithRecover(func() { ... }, func(r interface{}) { ... })
func SafeGoWithRecover(fn func(), onPanic func(interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Goroutine panic recovered: %v\nStack trace:\n%s", r, string(debug.Stack()))
				if onPanic != nil {
					// 安全调用自定义处理函数
					func() {
						defer func() {
							if r2 := recover(); r2 != nil {
								logger.Errorf("Panic handler itself panicked: %v", r2)
							}
						}()
						onPanic(r)
					}()
				}
			}
		}()
		fn()
	}()
}

// MustNotPanic 包装函数，如果panic则记录日志并返回error
// 用于需要返回error的场景
func MustNotPanic(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Function panicked: %v\nStack trace:\n%s", r, string(debug.Stack()))
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()
	return fn()
}
