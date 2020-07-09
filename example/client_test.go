package example

import (
	"github.com/feixiaobo/go-xxl-job-client"
	"github.com/feixiaobo/go-xxl-job-client/option"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestXxlClient(t *testing.T) {
	client := xxl.NewXxlClient(
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
