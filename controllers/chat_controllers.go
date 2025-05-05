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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// request to chat
//
// Struct for receiving JSON request
type FaceRequest struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

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
			ErrorResponse(ctx, http.StatusBadRequest, "Parsing Error", err.Error())
			return
		}

		// same messages in senders
		// Set the date for the messageDetails

		messageDetails.Date = time.Now().UTC()
		messageDetails.MessageId = primitive.NewObjectID().Hex()
		err := SaveMessageByUserId(mctx, messageDetails.Destination, app, *userDetails, messageDetails)
		if err != nil {
			ctx.JSON(http.StatusOK, bson.M{"error": err.Error()})
			return
		}

		SaveMessageForWebSocket(mctx, app, *userDetails, messageDetails)

		// same messages in receiver
		nameFields := strings.Split(messageDetails.Name, " ")
		var receiverDetails models.UserDetails

		receiverDetails.Email = messageDetails.Email
		receiverDetails.FirstName = nameFields[0]
		receiverDetails.LastName = nameFields[1]
		receiverDetails.Profile = messageDetails.Profile
		receiverDetails.UserID = messageDetails.Destination

		messageDetails.Name = userDetails.FirstName + " " + userDetails.LastName
		messageDetails.Email = userDetails.Email
		messageDetails.Profile = userDetails.Profile

		err = SaveMessageByUserId(mctx, receiverDetails.UserID, app, *&receiverDetails, messageDetails)
		if err != nil {
			ctx.JSON(http.StatusOK, bson.M{"error": err.Error()})
			return
		}
		err = SaveMessageForWebSocket(mctx, app, *&receiverDetails, messageDetails)
		if err != nil {
			ctx.JSON(http.StatusOK, bson.M{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, bson.M{"success": userDetails.UserID})

	}
}

func SaveMessageByUserId(mctx context.Context, subId string, app *config.AppConfig, userDetails models.UserDetails, messageDetails models.Message) error {
	// Step 1: Try to update existing sub_id
	filter := bson.M{
		"user_id":      userDetails.UserID,
		"chats.sub_id": subId,
	}
	update := bson.M{
		"$push": bson.M{
			"chats.$.messages": messageDetails,
		},
		"$set": bson.M{
			"chats.$.date":         messageDetails.Date,
			"chats.$.name":         messageDetails.Name,
			"chats.$.profile":      messageDetails.Profile,
			"chats.$.is_unread":    false,
			"chats.$.last_message": messageDetails.Message,
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
				"sub_id":       messageDetails.Destination,
				"date":         messageDetails.Date,
				"name":         messageDetails.Name,
				"profile":      messageDetails.Profile,
				"is_unread":    false,
				"messages":     []interface{}{messageDetails},
				"last_message": messageDetails.Message,
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
func SaveMessageForWebSocket(mctx context.Context, app *config.AppConfig, userDetails models.UserDetails, messageDetails models.Message) error {
	filter := bson.M{
		"user_id": userDetails.UserID,
	}

	newDoc := bson.M{
		"user_id":     userDetails.UserID,
		"destination": messageDetails.Destination,
		"message_id":  messageDetails.MessageId,
		"message":     messageDetails.Message,
		"date":        messageDetails.Date,
		"name":        messageDetails.Name,
		"profile":     messageDetails.Profile,
		"email":       messageDetails.Email,
	}

	opts := options.Replace().SetUpsert(true)

	result, err := app.Client.Database("talkmore").Collection("wsmessages").ReplaceOne(mctx, filter, newDoc, opts)
	if err != nil {
		log.Printf("Error replacing message for user %s: %v", userDetails.UserID, err)
		return fmt.Errorf("failed to replace message: %w", err)
	}

	if result.UpsertedCount > 0 {
		log.Printf("Created new document for user %s", userDetails.UserID)
	} else {
		log.Printf("Replaced message document for user %s", userDetails.UserID)
	}

	return nil
}

func GetChats(app *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Struct to hold pagination parameters
		var chatListRequest models.ChatSkipLimit
		if err := ctx.ShouldBindJSON(&chatListRequest); err != nil {
			ErrorResponse(ctx, http.StatusBadRequest, "Parsing Error", err.Error())
			return
		}

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

		filter := bson.M{"user_id": userDetails.UserID}

		//MongoDB projection to slice the notification_details array
		projection := bson.M{
			"chats": bson.M{
				"$slice": []int{chatListRequest.Skip, chatListRequest.Limit},
			},
		}

		// Define sorting options (e.g., sort by last_message_date in descending order)
		sortOption := bson.M{
			"chats.date": 1, // -1 for descending (newest first), 1 for ascending
		}
		// Struct to hold the result
		var result struct {
			ChatList []models.ChatUsers `json:"chats" bson:"chats"`
		}
		// Execute Query
		err := app.Client.Database("talkmore").Collection("chats").FindOne(mctx, filter, options.FindOne().SetProjection(projection).SetSort(sortOption)).Decode(&result)

		// handle error
		if err != nil {
			if err == mongo.ErrNoDocuments {
				SuccessResponse(ctx, "No chats", nil)
			} else {
				ErrorResponse(ctx, http.StatusInternalServerError, "Sorry! Server Error", err.Error())
			}
			return
		}
		SuccessResponse(ctx, "Your chat list", result.ChatList)
	}
}

func GetMessages(app *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Struct to hold pagination parameters
		var messageSkipLimit models.MessagesSkitLimit
		if err := ctx.ShouldBindJSON(&messageSkipLimit); err != nil {
			ErrorResponse(ctx, http.StatusBadRequest, "parsing error", err.Error())
			return
		}

		clientToken, tokenError := GetMyToken(ctx)
		if tokenError != "" {
			ErrorResponse(ctx, http.StatusUnauthorized, "token error", tokenError)
			ctx.Abort()
			return
		}
		userDetails, idError := GetMyId(mctx, app, clientToken)
		if idError != "" {
			ErrorResponse(ctx, http.StatusInternalServerError, "User Details Error", idError)
			ctx.Abort()
			return
		}

		pipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{
				"user_id": userDetails.UserID,
			}}},
			bson.D{{Key: "$unwind", Value: "$chats"}},
			bson.D{{Key: "$match", Value: bson.M{
				"chats.sub_id": messageSkipLimit.SubID,
			}}},
			bson.D{{Key: "$project", Value: bson.M{
				"_id": 0,
				"messages": bson.M{
					"$slice": []interface{}{
						bson.M{
							"$reverseArray": bson.M{
								"$sortArray": bson.M{
									"input":  "$chats.messages",
									"sortBy": bson.M{"date": 1},
								},
							},
						},
						messageSkipLimit.Skip, messageSkipLimit.Limit,
					},
				},
			}}},
		}

		cursor, err := app.Client.Database("talkmore").Collection("chats").Aggregate(mctx, pipeline)
		if err != nil {
			ErrorResponse(ctx, http.StatusInternalServerError, "Aggregation error", err.Error())
			return
		}

		var results []struct {
			Messages []models.Message `bson:"messages"`
		}
		if err := cursor.All(mctx, &results); err != nil {
			ErrorResponse(ctx, http.StatusInternalServerError, "Cursor error", err.Error())
			return
		}

		if len(results) == 0 {
			SuccessResponse(ctx, "No messages found", nil)
			return
		}

		// Assuming there's only one matched chat
		SuccessResponse(ctx, "Messages found", results[0].Messages)
	}
}
