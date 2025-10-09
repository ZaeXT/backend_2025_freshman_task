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

// Init initializes the MongoDB client using configuration values.
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
    // Ping to ensure connection is established
    if err := cli.Ping(c, nil); err != nil {
        return nil, err
    }
    log.Println("MongoDB 已连接")
    client = cli
    return client, nil
}

// Client returns the initialized mongo client.
func Client() *mongo.Client { return client }

// DB returns the default database.
func DB() *mongo.Database {
    return client.Database(config.Get().MongoDBName)
}


