package example

import (
	"github.com/feixiaobo/go-xxl-job-client"
	"github.com/feixiaobo/go-xxl-job-client/option"
	"testing"
)

func TestXxlClient(t *testing.T) {
	go func() {
		client := xxl.NewXxlClient(
			option.WithAppName("go-example"),
			option.WithClientPort(8084),
		)
		client.Run()
	}()
	client := xxl.NewXxlClient(
		option.WithClientPort(8083),
	)
	client.RegisterJob("testJob", JobTest)
	client.Run()
}
