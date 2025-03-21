package controllers

import (
	"context"
	"my-work/config"
	"my-work/models"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func Generate_OTP() *int {

	return nil

}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plain-text password with a stored hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SendMail(userMail string, message string) bool {
	return true
}

func GetMyToken(ctx *gin.Context) (string, string) {
	// Extract Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", "authorization header missing"
	}

	// Parse Bearer token (case-insensitive)
	fields := strings.Fields(authHeader)
	if len(fields) < 2 || !strings.EqualFold(fields[0], "bearer") {
		return "", "invalid authorization format, expected 'Bearer <token>'"
	}
	return fields[1], ""
}

func GetMyId(mctx context.Context, app *config.AppConfig, clientToken string) (*models.UserDetails, string) {
	var userDetails models.UserDetails
	filter := bson.M{
		"access_token": clientToken,
	}

	// Define the projection to return specific fields
	opts := options.FindOne().SetProjection(bson.M{
		"user_id":    1,
		"first_name": 1,
		"last_name":  1,
		"email":      1,
		"_id":        0,
	})

	// Variable to store the result

	// Execute the query
	err := app.Client.Database("talkmore").Collection("users").FindOne(mctx, filter, opts).Decode(&userDetails)
	if err != nil {
		return nil, err.Error()
	}

	return &userDetails, ""
}
