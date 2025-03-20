package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AppConfig holds application-wide configuration
type AppConfig struct {
	Client         *mongo.Client
	SecretKey      []byte
	RequireDBCheck bool
	Validator      *validator.Validate
}

// Init initializes the application configuration
func Init() (*AppConfig, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Connect to MongoDB
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Verify connection with a ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("MongoDB ping failed: %v", err)
	}

	// Load secret key
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY not set in environment")
	}

	// Initialize validator
	validate := validator.New()

	return &AppConfig{
		Client:         client,
		SecretKey:      []byte(secretKey),
		RequireDBCheck: os.Getenv("REQUIRE_DB_CHECK") == "true",
		Validator:      validate,
	}, nil
}
