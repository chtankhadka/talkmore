package middleware

import (
	"context"
	"log"
	"my-work/config"
	"my-work/models"
	"my-work/token"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Authentication is a Gin middleware for JWT validation
func Authentication(app *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Set a short timeout for database operations
		mctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Extract Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			ctx.Abort()
			return
		}

		// Parse Bearer token (case-insensitive)
		fields := strings.Fields(authHeader)
		if len(fields) < 2 || !strings.EqualFold(fields[0], "bearer") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format, expected 'Bearer <token>'"})
			ctx.Abort()
			return
		}

		clientToken := fields[1]

		// Validate token
		claims, err := token.ValidateToken(clientToken, app)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		// Optional database verification
		if app.RequireDBCheck {
			filter := bson.M{"user_id": claims.UID, "access_token": clientToken}
			var user models.SetSignUpModel
			err := app.Client.Database("talkmore").Collection("users").FindOne(mctx, filter).Decode(&user)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token not found or user unauthorized"})
				} else {
					log.Printf("Database error during token check: %v", err)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				}
				ctx.Abort()
				return
			}

			// Check if token is revoked
			if user.Revoked {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
				ctx.Abort()
				return
			}
		}

		// Set claims in context for downstream handlers
		ctx.Set("email", claims.Email)
		ctx.Set("uid", claims.UID)

		// Proceed to the next handler
		ctx.Next()
	}
}

// RequireAuthWithRole extends Authentication to enforce role-based access
func RequireAuthWithRole(app *config.AppConfig, requiredRole string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Run basic authentication first
		Authentication(app)(ctx)
		if ctx.IsAborted() {
			return
		}

		// Example: Check role (assumes roles are stored in DB or token)
		// Here, you'd fetch the user's role from the database or token claims
		uid := ctx.GetString("uid")
		mctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var user models.SetSignUpModel
		err := app.Client.Database("talkmore").Collection("users").FindOne(mctx, bson.M{"user_id": uid}).Decode(&user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user role"})
			ctx.Abort()
			return
		}

		// Placeholder: Assume user.Role exists in your User model
		// if user.Role != requiredRole {
		// 	ctx.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		// 	ctx.Abort()
		// 	return
		// }

		ctx.Next()
	}
}
