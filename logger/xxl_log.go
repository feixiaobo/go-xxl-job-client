package logger

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type LogResult struct {
	FromLineNum int32
	ToLineNum   int32
	LogContent  string
	IsEnd       bool
}

func (LogResult) JavaClassName() string {
	return "com.xxl.job.core.biz.model.LogResult"
}

var dateFormat = "2006-01-02 15:04:05"
var filePathFormat = "2006-01-02"
var logPath = "/data/applogs/xxl-job/jobhandler/"

func Info(ctx context.Context, args ...interface{}) {
	jobMap := ctx.Value("jobParam")
	if jobMap != nil {
		jobParamMap, ok := jobMap.(map[string]map[string]interface{})["logParam"]
		if ok {
			logid, ok := jobParamMap["logId"]
			if ok {
				nowTime := time.Now()

				var buffer bytes.Buffer
				buffer.WriteString(nowTime.Format(dateFormat))
				buffer.WriteString("  [")

				jobName, ok := jobParamMap["jobName"]
				if ok {
					buffer.WriteString(jobName.(string))
				}
				buffer.WriteString("#")
				jobFunc, ok := jobParamMap["jobFunc"]
				if ok {
					buffer.WriteString(jobFunc.(string))
				}
				buffer.WriteString("]-[")

				jobId, ok := jobParamMap["jobId"]
				if ok {
					buffer.WriteString(fmt.Sprintf("jobId:%d", jobId.(int32)))
				}
				buffer.WriteString("]  ")
				if len(args) > 0 {
					for _, arg := range args {
						buffer.WriteString(fmt.Sprintf("%v", arg))
					}
				}
				buffer.WriteString("\r\n")

				logId := logid.(int64)
				writeLog(getLogPath(nowTime), fmt.Sprintf("%d", logId)+".log", buffer.String())
			}
		}
	}
}

func getLogPath(nowTime time.Time) string {
	return logPath + nowTime.Format(filePathFormat)
}

func writeLog(logPath, logFile, log string) error {
	if strings.Trim(logFile, " ") != "" {
		fileFullPath := logPath + "/" + logFile
		file, err := os.OpenFile(fileFullPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil && os.IsNotExist(err) {
			err = os.MkdirAll(logPath, os.ModePerm)
			if err == nil {
				file, err = os.OpenFile(fileFullPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					return err
				}
			}
		}

		if file != nil {
			defer file.Close()
			res, err := file.Write([]byte(log))
			if err != nil {
				return err
			}
			if res <= 0 {
				return errors.New("write log failed")
			}
		}
	}
	return nil
}

func ReadLog(logDateTim, logId int64, fromLineNum int32) (line int32, content string) {
	nowtime := time.Unix(logDateTim/1000, 0)
	fileName := getLogPath(nowtime) + "/" + fmt.Sprintf("%d", logId) + ".log"
	file, err := os.Open(fileName)
	totalLines := int32(1)
	var buffer bytes.Buffer
	if err == nil {
		defer file.Close()

		rd := bufio.NewReader(file)
		for {

			line, err := rd.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			if totalLines >= fromLineNum {
				buffer.WriteString(line)
			}
			totalLines ++
		}
	}
	return totalLines, buffer.String()
}
