package homepage

import "my-work/models"

type UlalaSkipLimitWithCurrentTime struct {
	Skip     int    `json:"skip" bson:"-"`
	Limit    int    `json:"limit" bson:"-"`
	FromDate string `json:"from_date" bson:"from_date"`
	ToDate   string `json:"to_date" bson:"to_date"`
}

type Ulala struct {
	ID             string                   `json:"id" bson:"id"`
	UserID         string                   `json:"user_id" bson:"user_id"`
	FirstName      string                   `json:"first_name" bson:"first_name"`
	LastName       string                   `json:"last_name" bson:"last_name"`
	Email          string                   `json:"email" bson:"email"`
	Profile        string                   `json:"profile" bson:"profile"`
	PhotoUrl       string                   `json:"photo_url" bson:"photo_url"`
	Liked          int                      `json:"liked" bson:"liked"`
	Disliked       int                      `json:"disliked" bson:"disliked"`
	Post_Date      string                   `json:"post_date" bson:"post_date"`
	Changed_Date   string                   `json:"changed_date" bson:"changed_date"`
	UserInterests  *[]models.UserInterest   `json:"user_interests" bson:"user_interests"`
	UserLookingFor *[]models.UserLookingFor `json:"user_looking_for" bson:"user_looking_for"`
}
