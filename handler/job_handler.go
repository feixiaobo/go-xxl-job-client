package handler

import (
	"context"
	"github.com/feixiaobo/go-xxl-job-client/admin"
	"github.com/feixiaobo/go-xxl-job-client/logger"
	"github.com/feixiaobo/go-xxl-job-client/utils"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

type JobHandlerFunc func(ctx context.Context) error

var JobMap map[string]JobHandlerFunc

type JobQueue struct {
	JobId      int32
	CurrentJob *JobRunParam
	Run        int32 //0 stop, 1 run
	Queue      *utils.Queue
}

var JobRun = struct {
	sync.RWMutex
	queueMap map[int32]*JobQueue
}{queueMap: make(map[int32]*JobQueue)}

type JobRunParam struct {
	LogId             int64
	LogDateTime       int64
	JobName           string
	JobFuncName       string
	CurrentCancelFunc context.CancelFunc
	InputParam        string
}

func (jq *JobQueue) StartJob() {
	if atomic.CompareAndSwapInt32(&jq.Run, 0, 1) {
		log.Print("begin syn executor job, jobId:", jq.JobId)
		jq.synRunJob()
	}
}

func (jq *JobQueue) StopJob() bool {
	res := atomic.CompareAndSwapInt32(&jq.Run, 1, 0)
	log.Print("all job run finished, stop goroutine, result:", res, "jobId:", jq.JobId)
	return res
}

func (jq *JobQueue) synRunJob() {
	go func() {
		for {
			has, node := jq.Queue.Poll()
			if has {
				jq.CurrentJob = node.(*JobRunParam)
				RunJob(jq.JobId, jq.CurrentJob)
			} else {
				jq.StopJob()
				break
			}
		}
	}()
}

func AddJob(jobName string, function JobHandlerFunc) {
	if JobMap == nil {
		JobMap = make(map[string]JobHandlerFunc)
	} else {
		_, ok := JobMap[jobName]
		if ok {
			panic("the job had already register, job name can't be repeated:" + jobName)
		}
	}
	JobMap[jobName] = function
}

func CancelJob(jobId int32) {
	queue, has := JobRun.queueMap[jobId]
	if has {
		log.Print("job be canceled, id:", jobId)
		res := queue.StopJob()
		if res {
			if queue.CurrentJob != nil && queue.CurrentJob.CurrentCancelFunc != nil {
				queue.CurrentJob.CurrentCancelFunc()

				go func() {
					jobParam := make(map[string]map[string]interface{})
					logParam := make(map[string]interface{})
					logParam["logId"] = queue.CurrentJob.LogId
					logParam["jobId"] = jobId
					logParam["jobName"] = queue.CurrentJob.JobName
					logParam["jobFunc"] = queue.CurrentJob.JobFuncName
					jobParam["logParam"] = logParam
					ctx := context.WithValue(context.Background(), "jobParam", jobParam)
					logger.Info(ctx, "job canceled by admin!")
				}()
			}
			if queue.Queue != nil {
				queue.Queue.Clear()
			}
		}
	}
}

func ClearJob() {
	JobMap = map[string]JobHandlerFunc{}
	JobRun.RLock()
	JobRun.queueMap = make(map[int32]*JobQueue)
	JobRun.RUnlock()
}

func PushJob(jobId int32, trigger *JobRunParam) bool {
	queue, has := JobRun.queueMap[jobId] //map value是地址，读不加锁
	if has {
		err := queue.Queue.Put(trigger)
		if err == nil {
			queue.StartJob()
			return true
		}
		return false
	} else {
		JobRun.Lock() //任务map初始化锁
		q := utils.NewQueue()
		err := q.Put(trigger)
		if err != nil {
			return false
		}

		jobQueue := &JobQueue{
			JobId: jobId,
			Run:   0,
			Queue: q,
		}
		JobRun.queueMap[jobId] = jobQueue
		jobQueue.StartJob()
		JobRun.Unlock()
		return true
	}
}

func RunJob(jobId int32, trigger *JobRunParam) {
	returnt := xxl.ReturnT{
		Code:    http.StatusOK,
		Content: "success",
	}

	jobParam := make(map[string]map[string]interface{})

	if trigger.InputParam != "" {
		inputParam := make(map[string]interface{})
		params := strings.Split(trigger.InputParam, ",")
		if len(params) > 0 {
			for _, param := range params {
				if param != "" {
					jobP := strings.Split(param, "=")
					if len(jobP) > 0 {
						inputParam[jobP[0]] = jobP[1]
					}
				}
			}
		}
		jobParam["inputParam"] = inputParam
	}

	fun, ok := JobMap[trigger.JobName]
	if ok {
		logParam := make(map[string]interface{})
		logParam["logId"] = trigger.LogId
		logParam["jobId"] = jobId
		logParam["jobName"] = trigger.JobName
		logParam["jobFunc"] = trigger.JobFuncName
		jobParam["logParam"] = logParam

		valueCtx, canFun := context.WithCancel(context.Background())
		defer canFun()

		trigger.CurrentCancelFunc = canFun
		ctx := context.WithValue(valueCtx, "jobParam", jobParam)
		logger.Info(ctx, "job begin start! param:", trigger.InputParam)
		err := fun(ctx)
		if err != nil {
			logger.Info(ctx, "job run failed! msg:", err.Error())
			returnt.Code = http.StatusInternalServerError
			returnt.Content = err.Error()
		} else {
			logger.Info(ctx, "job run success!")
		}
	} else {
		returnt.Code = http.StatusInternalServerError
		returnt.Content = "job handle not found"
	}

	callback := &xxl.HandleCallbackParam{
		LogId:         trigger.LogId,
		LogDateTim:    trigger.LogDateTime,
		ExecuteResult: returnt,
	}
	xxl.CallbackAdmin([]*xxl.HandleCallbackParam{callback})
}
