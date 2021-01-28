package middleware

import (
	"context"
	"github.com/cat-go/cat"
	"github.com/cat-go/cat/message"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"log"
)

var (
	cache *redis.ClusterClient
)

func Cache() *redis.ClusterClient {
	return cache
}

type RedisConfig struct {
	Addrs []string `json:"hosts"`
}

func InitRedis(c *RedisConfig) *redis.ClusterClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: c.Addrs,
	})
	pong, err := client.Ping().Result()
	if pong != "PONG" || err != nil {
		log.Fatal("redis conn error")
	}
	client.AddHook(traceHook{})
	cache = client
	return cache
}

func WithContextCache(ctx context.Context) *redis.ClusterClient {
	return cache.WithContext(ctx)
}

type traceHook struct {
}

func (t traceHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	tran := cat.NewTransaction(cat.TypeRedis, cmd.Name())
	tran.AddData(cat.TypeRedisCmd, cmd.String())
	cv := context.WithValue(ctx, CatCtxRedisTran, tran)
	if c, ok := ctx.(*gin.Context); ok {
		if t, ok := c.Get(CatCtxRootTran); ok {
			if t1, ok := t.(message.Transactor); ok {
				cat.SetChildTraceId(t1, tran)
				t1.AddChild(tran)
			}
		}
	}
	return cv, nil
}

func (t traceHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if tran, ok := ctx.Value(CatCtxMysqlTran).(message.Transactor); ok {
		tran.Complete()
	}
	return nil
}

func (t traceHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	tran := cat.NewTransaction(cat.TypeRedis, "redis.pipeline")
	for _, cmd := range cmds {
		tran.AddData(cat.TypeRedisCmd, cmd.String())
	}
	cv := context.WithValue(ctx, CatCtxRedisTran, tran)
	if c, ok := ctx.(*gin.Context); ok {
		if t, ok := c.Get(CatCtxRootTran); ok {
			if t1, ok := t.(message.Transactor); ok {
				t1.AddChild(tran)
			}
		}
	}
	return cv, nil
}

func (t traceHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if tran, ok := ctx.Value(CatCtxMysqlTran).(message.Transactor); ok {
		tran.Complete()
	}
	return nil
}
