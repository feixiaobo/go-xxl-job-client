package example

import (
	"github.com/feixiaobo/go-xxl-job-client/v2"
	"github.com/feixiaobo/go-xxl-job-client/v2/option"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestXxlClient(t *testing.T) {
	client := xxl.NewXxlClient(
		option.WithAccessToken("edqedewfrqdrfrfr"),
		option.WithEnableHttp(true), //xxl_job v2.2.0版本
		option.WithClientPort(8083),
		option.WithAdminAddress("http://localhost:8080/xxl-job-admin"),
	)
	client.SetLogger(&logrus.Entry{
		Logger: logrus.New(),
		Level:  logrus.InfoLevel,
	})
	client.RegisterJob("testJob", JobTest)
	client.Run()
}
