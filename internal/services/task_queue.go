package services

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/o0olele/opendeepwiki-go/internal/database/dao"
	"github.com/o0olele/opendeepwiki-go/internal/database/models"
	"go.uber.org/zap"
)

// TaskQueue manages a queue of documentation generation tasks
type TaskQueue struct {
	repoDir  string
	taskChan chan *Task
	taskDao  *dao.RepositoryTaskDAO
	repoDao  *dao.RepositoryDAO
}

// NewTaskQueue creates a new task queue
func NewTaskQueue(repoDir string) *TaskQueue {
	// Create repos directory if it doesn't exist
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		err := os.MkdirAll(repoDir, 0755)
		if err != nil {
			zap.L().Error("Failed to create repository directory: %v", zap.Error(err))
			return nil
		}
	}

	taskQueue := &TaskQueue{
		repoDir:  repoDir,
		taskChan: make(chan *Task, 1024), // Buffer size of 1024 tasks
		taskDao:  dao.NewRepositoryTaskDAO(),
		repoDao:  dao.NewRepositoryDAO(),
	}

	return taskQueue
}

// recoverPendingTasks 从数据库中恢复未完成的任务
func (tq *TaskQueue) recoverPendingTasks() {
	if len(tq.taskChan) > 0 {
		return
	}

	zap.L().Info("Recovering pending tasks from database...")

	// find all tasks with status not in ["completed", "failed"]
	tasks, err := tq.taskDao.ListRepositoryTasksByStatus(models.RepositoryStatusPending, 10, 0)
	if err != nil {
		zap.L().Error("Failed to list repository tasks: %v", zap.Error(err))
		return
	}

	if len(tasks) == 0 {
		return
	}

	for _, modelTask := range tasks {
		err := tq.AddTask(NewTaskFromModel(modelTask))
		if err != nil {
			zap.L().Error("Failed to add task to queue: %v", zap.Error(err))
			break
		}
	}
}

// AddTask adds a new task to the queue
func (tq *TaskQueue) AddTask(task *Task) error {
	select {
	case tq.taskChan <- task:
		zap.L().Info("Added task to queue", zap.String("task_id", task.GitURL), zap.String("git_url", task.GitURL))
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// ProcessTasks starts processing tasks from the queue
func (tq *TaskQueue) ProcessTasks() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Process Tasks err: %v", err)
			}
		}()

		var ticker = time.NewTicker(10 * time.Second)

		for {
			select {
			case task := <-tq.taskChan:
				task.Process(&TaskProcessParams{
					taskDao: tq.taskDao,
					repoDao: tq.repoDao,
					repoDir: tq.repoDir,
				})
			case <-ticker.C:
				if len(tq.taskChan) > 0 {
					continue
				}
				go tq.recoverPendingTasks()
			}
		}
	}()
}
