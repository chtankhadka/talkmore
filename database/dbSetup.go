package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DatabaseSetup() *mongo.Client {
	// Create a context with a 10 second timeout for the connection.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB using the URI provided

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Error connection to MongoDB: %v", err)
	}

	// Ping the MongoDB server to ensure the connection is established
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	fmt.Println("Successfully connected to the database")
	return client
}
