package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSenderVerifyEmail, opts ...asynq.Option) error
}
type RedisTaksDistributor struct {
	//this to send to task to redis queue
	client *asynq.Client
}

// forcing the redistaksDistributor to implement taskDistributor
func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaksDistributor{
		client: client,
	}
}
