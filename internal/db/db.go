package db

import (
	"context"
	"fmt"
	"log"

	"github.com/LittleAksMax/bids-policy-service/internal/health"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnectionConfig struct {
	Host     string
	Port     int
	User     string
	Passwd   string
	Database string
}

// DSN builds a MongoDB connection string from component parts.
func (connCfg *MongoConnectionConfig) DSN() string {
	return fmt.Sprintf(
		"mongodb://%s:%d/%s",
		connCfg.Host,
		connCfg.Port,
		connCfg.Database,
	)
}

type Config struct {
	Client   *mongo.Client
	Database *mongo.Database
	health.HealthChecker
}

func Connect(ctx context.Context, connCfg *MongoConnectionConfig) (*Config, error) {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	creds := options.Credential{
		Username:      connCfg.User,
		Password:      connCfg.Passwd,
		AuthSource:    connCfg.Database,
		AuthMechanism: "SCRAM-SHA-256",
	}
	opts := options.Client().ApplyURI(connCfg.DSN()).SetServerAPIOptions(serverAPI).SetAuth(creds)

	// Create a new client and connect to the server
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	// database
	db := client.Database(connCfg.Database)

	// Send a ping to confirm a successful connection
	var result bson.M
	if err := db.RunCommand(ctx, bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}
	log.Println("Pinged your deployment. You successfully connected to MongoDB!")

	return &Config{
		Client:   client,
		Database: db,
	}, nil
}

func (cfg *Config) HealthCheck(ctx context.Context) error {
	var result bson.M
	return cfg.Database.RunCommand(ctx, bson.D{{"ping", 1}}).Decode(&result)
}
