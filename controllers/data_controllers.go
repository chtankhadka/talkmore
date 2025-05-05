package controllers

import (
	"context"
	"my-work/config"

	"go.mongodb.org/mongo-driver/bson"
)

func GetChatMessages(app *config.AppConfig, userID, subID string) ([]bson.M, error) {
	var chat struct {
		Chats []bson.M `bson:"chats"`
	}
	err := app.Client.Database("chatmore").Collection("chats").
		FindOne(context.TODO(), bson.M{"user_id": userID, "chats.sub_id": subID}).
		Decode(&chat)
	if err != nil {
		return nil, err
	}

	// Extract messages from the correct sub_id chat
	for _, c := range chat.Chats {
		if c["sub_id"] == subID {
			if messages, ok := c["messages"].([]bson.M); ok {
				return messages, nil
			}
		}
	}
	return nil, nil
}
