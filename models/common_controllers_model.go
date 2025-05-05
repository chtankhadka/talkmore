package models

type UserDetails struct {
	UserID    string `json:"user_id" bson:"user_id"`
	FirstName string `json:"first_name" bson:"first_name"`
	LastName  string `json:"last_name" bson:"last_name"`
	Email     string `json:"email" bson:"email"`
	Profile   string `json:"profile" bson:"profile"`
}

type UserInterest struct {
	Interest string `json:"interest" bson:"interest"`
}

type UserLookingFor struct {
	LookingFor string `json:"looking_for" bson:"looking_for"`
}

type UserHistory struct {
	History string `json:"history" bson:"history"`
}

type UserMoreDetails struct {
	UserID         string            `json:"user_id" bson:"user_id"`
	UserInterests  *[]UserInterest   `json:"user_interests" bson:"user_interests"`
	UserLookingFor *[]UserLookingFor `json:"user_looking_for" bson:"user_looking_for"`
	UserHistories  *[]UserHistory    `json:"user_history" bson:"user_history"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
}
