package handler

import (
	"context"
	"sync"
)

type JobHandlerFunc func(ctx context.Context) error

var JobCancelMap sync.Map

var JobMap map[string]JobHandlerFunc

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

//xxl-job-admin的实现在新job线程放入队列，老的线程就kill，所以维护最新取消函数
func RegisterCancelFunc(jobId int32, cancelFun context.CancelFunc) {
	canFun, ok := JobCancelMap.Load(jobId)
	if ok {
		canFun.(context.CancelFunc)()
	}
	JobCancelMap.Store(jobId, cancelFun)
}

func CancelJob(jobId int32) {
	cancelFun, ok := JobCancelMap.Load(jobId)
	if ok {
		cancelFun.(context.CancelFunc)()
		RemoveCancelFun(jobId)
	}
}

func RemoveCancelFun(jobId int32) {
	JobCancelMap.Delete(jobId)
}

func ClearJob() {
	JobMap = make(map[string]JobHandlerFunc)
	JobCancelMap = sync.Map{}
}
