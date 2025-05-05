package models

import (
	"time"
)

type Message struct {
	MessageId   string    `json:"message_id" bson:"message_id"`
	Destination string    `json:"destination" bson:"destination"`
	Message     string    `json:"message" bson:"message"`
	Date        time.Time `json:"date" bson:"date"`
	Name        string    `json:"name" bson:"name"`
	Profile     string    `json:"profile" bson:"profile"`
	Email       string    `json:"email" bson:"email"`
}

type ChatUsers struct {
	SubId       string    `json:"sub_id" bson:"sub_id"`
	Date        time.Time `json:"date" bson:"date"`
	Name        string    `json:"name" bson:"name"`
	Profile     string    `json:"profile" bson:"profile"`
	IsUnread    bool      `json:"is_unread" bson:"is_unread"`
	LastMessage string    `json:"last_message" bson:"last_message"`
}

type ChatSkipLimit struct {
	Skip  int `json:"skip" bson:"-"`
	Limit int `json:"limit" bson:"-"`
}

type MessagesSkitLimit struct {
	Skip  int    `json:"skip" bson:"-"`
	Limit int    `json:"limit" bson:"-"`
	SubID string `json:"sub_id" bson:"sub_id"`
}
