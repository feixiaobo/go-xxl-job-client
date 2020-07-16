package example

import (
	"context"
	"github.com/feixiaobo/go-xxl-job-client/v2"
	"github.com/feixiaobo/go-xxl-job-client/v2/logger"
	"log"
)

func JobTest(ctx context.Context) error {
	val, _ := xxl.GetParam(ctx, "test")
	log.Print(">>>>>>>>>>>>>>>>", val)
	logger.Info(ctx, "test job!!!!!")
	param, _ := xxl.GetParam(ctx, "name") //获取输入参数
	logger.Info(ctx, "the input param:", param)
	shardingIdx, shardingTotal := xxl.GetSharding(ctx) //获取分片参数
	logger.Info(ctx, "the sharding param: idx:", shardingIdx, ", total:", shardingTotal)
	return nil
}
