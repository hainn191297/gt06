package svc

import (
	"context"
	"gt06/config"
	"gt06/database"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServiceContext struct {
	Config       config.Config
	MongoDBModel database.MongoDBModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	client, err := initMongoClient(c.MongoURI)
	if err != nil {
		logx.Errorf("Failed to initialize MongoClient: %v", err)
		return nil
	}

	return &ServiceContext{
		Config:       c,
		MongoDBModel: database.NewMongoDBModel(client, "gps"), // TODO: use config dbname instead
	}
}

// initMongoClient sets up the MongoDB client
func initMongoClient(mongoURI string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(mongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	logx.Info("Connected to MongoDB")
	return client, nil
}
