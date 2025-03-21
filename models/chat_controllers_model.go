package models

import "time"

type Message struct {
	ID          string    `json:"id" bson:"id"`
	Destination string    `json:"destination" bson:"destination"`
	Message     string    `json:"message" bson:"message"`
	Date        time.Time `json:"date" bson:"date"`
	Name        string    `json:"name" bson:"name"`
	Profile     string    `json:"profile" bson:"profile"`
	Email       string    `json:"email" bson:"email"`
}
