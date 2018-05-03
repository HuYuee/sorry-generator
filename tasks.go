package main

import (
	"runtime"
	"sync"
)

const (
	// StateWaiting 等待状态（添加后默认）
	StateWaiting = "waitting"
	// StateCompleted 完成状态
	StateCompleted = "completed"
	// StateError 失败状态
	StateError = "failed"
	// StateNone 空状态（没有构建任务）
	StateNone = "none"
)

// 更新任务状态的互斥锁
var updateTaskStateMutex sync.Mutex

// 读取任务状态的互斥锁
var loadTaskStateMutex sync.Mutex

// 执行任务的缓冲通道
var taskChan = make(chan Task, runtime.NumCPU())

// 储存任务状态的 map
var taskState = make(map[string]string)

// Task 添加到队列的任务结构体
type Task struct {
	TplKey       string
	Subs         Subs
	RunnableList []makeFunc
}

//
type makeFunc func(string, Subs) (string, error)

// addMakeTask 添加一个生成任务
func addMakeTask(task Task) string {
	taskChan <- task
	hash := task.Subs.Hash(task.TplKey)
	updateTaskState(hash, StateWaiting)
	return hash
}

// updateTaskState 更新任务状态
// 状态更新操作加锁
func updateTaskState(hash, state string) {
	updateTaskStateMutex.Lock()
	{
		taskState[hash] = state
	}
	updateTaskStateMutex.Unlock()
}

// loadTaskState 读取任务状态
// 状态更新操作加锁
func loadTaskState(hash string) (state string) {
	loadTaskStateMutex.Lock()
	{
		resultState, exists := taskState[hash]
		if !exists {
			state = StateNone
		} else {
			state = resultState
		}
	}
	loadTaskStateMutex.Unlock()
	return
}

// asyncMakeAction 异步生成任务启动
// goroutine 函数
func asyncMakeAction() {
next:
	for {
		task := <-taskChan
		var curTaskHash string
		for _, f := range task.RunnableList {
			if hash, err := f(task.TplKey, task.Subs); err != nil {
				updateTaskState(hash, StateError)
				continue next
			}else{
				curTaskHash = hash
			}
		}
		updateTaskState(curTaskHash, StateCompleted)
	}
}
