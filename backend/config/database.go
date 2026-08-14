package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var DB *mongo.Database

func ConnectDatabase() error {
	mongoURI := os.Getenv("MONGO_URI")

	if mongoURI == "" {
		return fmt.Errorf("MONGO_URI is not set")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.Client().
		ApplyURI(mongoURI).
		SetServerAPIOptions(serverAPI)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	client, err := mongo.Connect(opts)
	if err != nil {
		return err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	DB = client.Database("live_polling_db")

	fmt.Println("MongoDB connected successfully!")

	return nil
}