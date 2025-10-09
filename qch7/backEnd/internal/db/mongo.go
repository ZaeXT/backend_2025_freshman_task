package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"backEnd/internal/config"
)

var client *mongo.Client

// Init 使用配置初始化 MongoDB 客户端。
func Init(ctx context.Context) (*mongo.Client, error) {
	if client != nil {
		return client, nil
	}

	cfg := config.Get()
	cli, err := mongo.NewClient(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, err
	}
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := cli.Connect(c); err != nil {
		return nil, err
	}
	// 通过 Ping 确认连接可用
	if err := cli.Ping(c, nil); err != nil {
		return nil, err
	}
	log.Println("MongoDB 已连接")
	client = cli
	return client, nil
}

// Client 返回已初始化的 Mongo 客户端。
func Client() *mongo.Client { return client }

// DB 返回默认数据库句柄。
func DB() *mongo.Database {
	return client.Database(config.Get().MongoDBName)
}
