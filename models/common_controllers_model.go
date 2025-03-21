package models

type UserDetails struct {
	UserID    string `bson:"user_id"`
	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
	Email     string `bson:"email"`
	Profile   string `bson:"profile"`
}
