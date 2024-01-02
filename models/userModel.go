package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id           primitive.ObjectID `bson:"_id"`
	FirstName    string             `bson:"first_name" json:"first_name,omitempty"`
	LastName     string             `bson:"last_name" json:"last_name,omitempty"`
	Password     string             `bson:"password" json:"password,omitempty"`
	Email        string             `bson:"email" validate:"email" json:"email,omitempty"`
	Phone        string             `bson:"phone" json:"phone,omitempty"`
	Token        string             `bson:"token" json:"token,omitempty"`
	UserType     string             `bson:"user_type" json:"user_type,omitempty"`
	RefreshToken string             `bson:"refresh_token" json:"refresh_token,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
	UserId       string             `bson:"user_id" json:"user_id,omitempty"`
}
