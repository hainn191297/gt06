package svc

import (
	"context"
	"gt06/config"
	"gt06/database"
	"time"

	"github.com/gocql/gocql"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServiceContext struct {
	Config        config.Config
	MongoDBModel  database.MongoDBModel
	ScyllaDBModel database.ScyllaDBModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	svc := &ServiceContext{
		Config: c,
	}

	// Initialize MongoDB if configured
	if c.MongoURI != "" {
		client, err := initMongoClient(c.MongoURI)
		if err != nil {
			logx.Errorf("Failed to initialize MongoClient: %v", err)
		} else {
			dbName := c.DBName
			if dbName == "" {
				dbName = "gt06"
			}
			svc.MongoDBModel = database.NewMongoDBModel(client, dbName)
		}
	}

	// Initialize ScyllaDB if configured
	if len(c.ScyllaHosts) > 0 {
		consistency := gocql.LocalOne // Default consistency
		if c.ScyllaConsistency != "" {
			parsed := gocql.ParseConsistency(c.ScyllaConsistency)
			consistency = parsed
		}
		scyllaModel, err := database.NewScyllaDBModel(c.ScyllaHosts, c.ScyllaKeyspace, consistency)
		if err != nil {
			logx.Errorf("Failed to initialize ScyllaDBModel: %v", err)
		} else {
			svc.ScyllaDBModel = scyllaModel
		}
	}

	return svc
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
