package models

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetSignUpModel struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	OTP        int                `json:"otp" bson:"otp"`
	Count      int                `json:"count" bson:"count"`
	First_Name string             `json:"first_name" bson:"first_name" validate:"required,min=2,max=30"`
	Last_Name  string             `json:"last_name" bson:"last_name" validate:"required,min=2,max=30"`
	Password   string             `json:"password" bson:"password" validate:"required,min=6"`
	Email      string             `json:"email" bson:"email" validate:"required"`
	User_ID    string             `json:"user_id" bson:"user_id"`
	Expires_At time.Time          `json:"expires_at" bson:"expires_at"`
}

type SetSignUpModel struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	First_Name    string             `json:"first_name" bson:"first_name"`
	Last_Name     string             `json:"last_name" bson:"last_name"`
	Password      string             `json:"password" bson:"password"`
	Profile_Url   *string            `json:"profile_url" bson:"profile_url"`
	Email         string             `json:"email" bson:"email"`
	Access_Token  string             `json:"access_token" bson:"access_token"`
	Refresh_Token string             `json:"refresh_token" bson:"refresh_token"`
	Refresh_ID    time.Time          `json:"refresh_id" bson:"refresh_id"`
	Created_At    time.Time          `json:"created_at" bson:"created_at"`
	Updated_At    time.Time          `json:"updated_at" bson:"updated_at"`
	User_ID       string             `json:"user_id" bson:"user_id"`
	Revoked       bool               `bson:"revoked" json:"revoked"`
}

type SigningDetails struct {
	Email string `json:"email"`
	UID   string `json:"uid"`
	jwt.RegisteredClaims
}

type TokenVerify struct {
	Token string `json:"token" bson:"token"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
