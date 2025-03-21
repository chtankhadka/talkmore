package token

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"my-work/config"
	"my-work/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GenerateTokenPair creates a new access and refresh token pair
func GenerateTokenPair(email, uid string, app *config.AppConfig) (models.TokenPair, error) {
	// Access token claims (short-lived)
	accessClaims := &models.SigningDetails{
		Email: email,
		UID:   uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)), // 15 minutes
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "myWork",
		},
	}

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString(app.SecretKey)
	if err != nil {
		return models.TokenPair{}, err
	}

	// Refresh token claims (longer-lived)
	refreshID := generateRandomID(16) // Unique ID for revocation
	refreshClaims := &models.SigningDetails{
		UID: uid, // Include UID for validation
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "myWork",
			ID:        refreshID, // Unique identifier (jti)
		},
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString(app.SecretKey)
	if err != nil {
		return models.TokenPair{}, err
	}

	// // Store tokens in database
	// err = updateTokens(signedAccessToken, signedRefreshToken, refreshID, uid, app)
	// if err != nil {
	// 	return models.TokenPair{}, err
	// }

	return models.TokenPair{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
	}, nil
}

func ValidateToken(tokenString string, app *config.AppConfig) (*models.SigningDetails, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.SigningDetails{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return app.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*models.SigningDetails)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshTokens generates a new token pair using a valid refresh token
func RefreshTokens(refreshTokenString string, app *config.AppConfig) (models.TokenPair, error) {
	// Validate the refresh token
	claims, err := ValidateToken(refreshTokenString, app)
	if err != nil {
		return models.TokenPair{}, errors.New("invalid or expired refresh token")
	}

	// Optionally check if the refresh token is revoked in the database
	mctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{
		"user_id":       claims.UID,
		"refresh_token": refreshTokenString,
	}
	count, err := app.Client.Database("talkmore").Collection("users").CountDocuments(mctx, filter)
	if err != nil || count == 0 {
		return models.TokenPair{}, errors.New("refresh token not found or revoked")
	}

	// Generate a new token pair
	// Note: Email isn’t in the refresh token, so fetch it from DB or omit it
	var user models.SetSignUpModel
	err = app.Client.Database("talkmore").Collection("users").FindOne(mctx, bson.M{"user_id": claims.UID}).Decode(&user)
	if err != nil {
		return models.TokenPair{}, errors.New("user not found")
	}

	newTokenPair, err := GenerateTokenPair(user.Email, claims.UID, app)
	if err != nil {
		return models.TokenPair{}, err
	}

	// Update the user's tokens in the database
	_, err = app.Client.Database("talkmore").Collection("users").UpdateOne(
		mctx,
		bson.M{"user_id": claims.UID},
		bson.M{"$set": bson.M{
			"token":         newTokenPair.AccessToken,
			"refresh_token": newTokenPair.RefreshToken,
			"updated_at":    time.Now(),
		}},
	)
	if err != nil {
		log.Printf("Failed to update tokens: %v", err)
		return models.TokenPair{}, errors.New("failed to update tokens")
	}

	return newTokenPair, nil
}

func generateRandomID(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Failed to generate random ID: %v", err)
		return ""
	}
	return hex.EncodeToString(b)
}

// updateTokens stores tokens in MongoDB
func updateTokens(accessToken, refreshToken, refreshID, userID string, app *config.AppConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateObj := bson.D{
		{Key: "token", Value: accessToken},
		{Key: "refresh_token", Value: refreshToken},
		{Key: "refresh_id", Value: refreshID}, // Store for revocation
		{Key: "updated_at", Value: time.Now()},
	}

	upsert := true
	filter := bson.M{"user_id": userID}
	opts := options.UpdateOptions{Upsert: &upsert}

	_, err := app.Client.Database("talkmore").Collection("users").UpdateOne(ctx, filter, bson.D{
		{Key: "$set", Value: updateObj},
	}, &opts)
	if err != nil {
		log.Printf("Failed to update tokens: %v", err)
		return err
	}
	return nil
}
