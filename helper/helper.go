package helper

import (
	"context"
	"log"
	"my-work/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func IsFieldUsed(
	app *config.AppConfig,
	mctx context.Context,
	ctx *gin.Context,
	fieldName string,
	fieldValue interface{}) bool {
	count, err := app.Client.Database("talkmore").Collection("users").CountDocuments(mctx, bson.M{fieldName: fieldValue})
	if err != nil {
		log.Panic(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return true
	}

	if count > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fieldName + " is already used"})
		return true
	}
	return false
}
