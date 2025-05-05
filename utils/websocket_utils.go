package utils

import (
	"context"
	"fmt"
	"log"
	"my-work/config"
	"my-work/controllers"
	"my-work/models"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func WatchChatsCollection(app *config.AppConfig, userID string, conn *websocket.Conn) {
	collection := app.Client.Database("talkmore").Collection("chats")
	ctx := context.Background()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "operationType", Value: bson.D{{Key: "$in", Value: bson.A{"insert", "update", "replace"}}}},
			{Key: "fullDocument.user_id", Value: userID}}}}}

	//open the Change stream
	stream, err := collection.Watch(ctx, pipeline, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		log.Printf("Error creating change stream: %v", err)

	}
	defer stream.Close(ctx)
	fmt.Println("Watching for changes on `user_id == ")
	// Process Change Events

	for stream.Next(ctx) {
		var changeDoc bson.M
		if err := stream.Decode(&changeDoc); err != nil {
			log.Println("Error decoding change document:", err)
			continue
		}
		fmt.Println("Change detected:")
		if fullDoc, ok := changeDoc["fullDocument"].(bson.M); ok {
			// Inspecting the "items" array
			if items, found := fullDoc["chats"].(bson.A); found {

				chatMap := (items[len(items)-1]).(bson.M)
				data := bson.M{
					"sub_id":       chatMap["sub_id"],
					"date":         chatMap["date"],
					"name":         chatMap["name"],
					"profile":      chatMap["profile"],
					"is_unread":    chatMap["is_unread"],
					"last_message": chatMap["last_message"],
				}

				// Send the modified data to the client (connection)
				err = conn.WriteJSON(data)
				if err != nil {
					fmt.Println("Error sending data over WebSocket:", err)
				}

			}
		}
	}

	if err := stream.Err(); err != nil {
		log.Fatal("Change stream error:", err)
	}
	fmt.Println("Change stream closed")
}
func WatchMessagesCollection(app *config.AppConfig, userID string, conn *websocket.Conn, done <-chan struct{}) {
	collection := app.Client.Database("talkmore").Collection("wsmessages")
	ctx := context.Background()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "operationType", Value: bson.D{{Key: "$in", Value: bson.A{"insert", "update", "replace"}}}},
			{Key: "fullDocument.user_id", Value: userID}}}},
	}

	stream, err := collection.Watch(ctx, pipeline)
	if err != nil {
		log.Printf("Error creating change stream: %v", err)
		return
	}
	defer stream.Close(ctx)

	for {
		select {
		case <-done:
			log.Printf("Stopping WatchMessagesCollection for user %s due to disconnect", userID)
			return
		case <-ctx.Done():
			log.Printf("Context cancelled for user %s", userID)
			return
		default:
			if !stream.Next(ctx) {
				if err := stream.Err(); err != nil {
					log.Printf("Change stream error for user %s: %v", userID, err)
				}
				return
			}
			var changeDoc bson.M
			if err := stream.Decode(&changeDoc); err != nil {
				log.Println("Error decoding change document:", err)
				continue
			}
			if fullDoc, ok := changeDoc["fullDocument"].(bson.M); ok {
				// Check if connection is still open before writing
				select {
				case <-done:
					log.Printf("Aborting write for user %s: connection closed", userID)
					return
				default:
					err = conn.WriteJSON(fullDoc)
					if err != nil {
						log.Printf("Error sending data over WebSocket for user %s: %v", userID, err)
						return
					}
				}
			}
		}
	}
}

func HandleClientMessage(app *config.AppConfig, userDetails models.UserDetails, messageDetails models.Message) {

	// same messages in senders
	// Set the date for the messageDetails
	mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messageDetails.Date = time.Now().UTC()
	messageDetails.MessageId = primitive.NewObjectID().Hex()

	// same messages in senders
	// Set the date for the messageDetails

	messageDetails.Date = time.Now().UTC()
	messageDetails.MessageId = primitive.NewObjectID().Hex()
	err := controllers.SaveMessageByUserId(mctx, messageDetails.Destination, app, userDetails, messageDetails)
	if err != nil {
		log.Printf("Error inserting message into MongoDB: %v", err)
	}

	err = controllers.SaveMessageForWebSocket(mctx, app, *&userDetails, messageDetails)
	if err != nil {
		log.Printf("Error inserting message into MongoDB: %v", err)
	}

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

	err = controllers.SaveMessageByUserId(mctx, userDetails.UserID, app, *&receiverDetails, messageDetails)
	if err != nil {
		log.Printf("Error inserting message into MongoDB: %v", err)

	}
	err = controllers.SaveMessageForWebSocket(mctx, app, *&receiverDetails, messageDetails)
	if err != nil {
		log.Printf("Error inserting message into MongoDB: %v", err)
	}

	log.Printf("Message from user %s saved: %s", userDetails.UserID, messageDetails.Message)
}
