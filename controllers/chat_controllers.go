package controllers

import (
	"context"
	"fmt"
	"log"
	"my-work/config"
	"my-work/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// request to chat
//

func SendText(app *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		clientToken, tokenError := GetMyToken(ctx)
		if tokenError != "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": tokenError})
			ctx.Abort()
			return
		}
		userDetails, idError := GetMyId(mctx, app, clientToken)
		if idError != "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": idError})
			ctx.Abort()
			return
		}

		var messageDetails models.Message
		if err := ctx.BindJSON(&messageDetails); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// same messages in senders
		// Set the date for the messageDetails

		messageDetails.Date = time.Now().UTC()
		messageDetails.ID = primitive.NewObjectID().Hex()
		err := SaveMessageByUserId(mctx, app, *userDetails, messageDetails)
		if err != nil {
			ctx.JSON(http.StatusOK, bson.M{"error": err.Error()})
			return
		}

		// same messages in receiver
		nameFields := strings.Split(messageDetails.Name, " ")
		var receiverDetails models.UserDetails

		receiverDetails.Email = messageDetails.Email
		receiverDetails.FirstName = nameFields[0]
		receiverDetails.LastName = nameFields[1]
		receiverDetails.Profile = messageDetails.Profile
		receiverDetails.UserID = messageDetails.Destination

		messageDetails.Destination = userDetails.UserID
		messageDetails.Name = userDetails.FirstName + " " + userDetails.LastName
		messageDetails.Email = userDetails.Email
		messageDetails.Profile = userDetails.Profile

		err = SaveMessageByUserId(mctx, app, *&receiverDetails, messageDetails)
		if err != nil {
			ctx.JSON(http.StatusOK, bson.M{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, bson.M{"success": userDetails.UserID})

	}
}

func SaveMessageByUserId(mctx context.Context, app *config.AppConfig, userDetails models.UserDetails, messageDetails models.Message) error {
	// Step 1: Try to update existing sub_id
	filter := bson.M{
		"user_id":      userDetails.UserID,
		"chats.sub_id": messageDetails.Destination,
	}
	update := bson.M{
		"$push": bson.M{
			"chats.$.messages": messageDetails,
		},
		"$set": bson.M{
			"chats.$.date":      messageDetails.Date,
			"chats.$.name":      messageDetails.Name,
			"chats.$.profile":   messageDetails.Profile,
			"chats.$.is_unread": false,
		},
	}

	result, err := app.Client.Database("talkmore").Collection("chats").UpdateOne(mctx, filter, update)
	if err != nil {
		log.Printf("Error updating chat for user %s, sub_id %s: %v", userDetails.UserID, messageDetails.Destination, err)
		return fmt.Errorf("failed to update chat: %w", err)
	}

	if result.MatchedCount > 0 {
		log.Printf("Added message to sub_id %s for user %s, updated date to %s", messageDetails.Destination, userDetails.UserID, messageDetails.Date.String())
		return nil
	}

	// Step 2: If no match, add new sub_id or create new document
	filter = bson.M{"user_id": userDetails.UserID}
	updateNewSub := bson.M{
		"$push": bson.M{
			"chats": bson.M{
				"sub_id":    messageDetails.Destination,
				"date":      messageDetails.Date,
				"name":      messageDetails.Name,
				"profile":   messageDetails.Profile,
				"is_unread": false,
				"messages":  []interface{}{messageDetails},
			},
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err = app.Client.Database("talkmore").Collection("chats").UpdateOne(mctx, filter, updateNewSub, opts)
	if err != nil {
		log.Printf("Error adding new sub_id or creating document for user %s: %v", userDetails.UserID, err)
		return fmt.Errorf("failed to add new sub_id or create document: %w", err)
	}

	if result.UpsertedCount > 0 {
		log.Printf("Created new document with main ID %s and sub_id %s", userDetails.UserID, messageDetails.Destination)
	} else {
		log.Printf("Added new sub_id %s to existing main ID %s with date %s", messageDetails.Destination, userDetails.UserID, messageDetails.Date.String())
	}
	return nil
}
