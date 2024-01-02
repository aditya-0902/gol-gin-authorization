package helpers

import (
	"context"
	"demo/database"
	"fmt"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	FirstName string
	LastName  string
	Email     string
	UId       string
	UserType  string
	jwt.StandardClaims
}

var userCollection = database.OpenCollection(database.Client, "user")

var SECRET_KEY = "my-secret-key"

func GenerateAllTokens(email, firstName, lastName, userType, userId string) (token, refreshToken string, err error) {
	claims := &SignedDetails{
		FirstName: firstName,
		LastName:  lastName,
		UId:       userId,
		UserType:  userType,
		Email:     email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	log.Println(claims)

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}
	log.Println(refreshClaims)
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	log.Println(token)
	log.Println(refreshToken)
	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func UpdateAllTokens(token, refreshToken, userId string) {
	var updatedObj primitive.D
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	updatedObj = append(updatedObj, bson.E{Key: "token", Value: token})
	updatedObj = append(updatedObj, bson.E{Key: "refresh_token", Value: refreshToken})
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updatedObj = append(updatedObj, bson.E{Key: "updated_at", Value: updatedAt})
	// upsert := true
	filter := bson.M{"user_id": userId}
	// opts := options.UpdateOptions{
	// 	Upsert: &upsert,
	// }
	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: updatedObj},
		},
		// &opts,
	)
	if err != nil {
		log.Panic(err)
	}
	return
}

func ValidateToken(signedToken string) (*SignedDetails, string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		msg := err.Error()
		log.Println(msg)
		return &SignedDetails{}, msg
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg := fmt.Sprintf("the token is invalid")
		return &SignedDetails{}, msg
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg := fmt.Sprintf("token is expired")
		return &SignedDetails{}, msg
	}
	return claims, ""
}
