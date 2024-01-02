package controllers

import (
	"context"
	"demo/database"
	helper "demo/helpers"
	model "demo/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var getCollection *mongo.Collection = database.OpenCollection(database.Client, "data")
var validate = validator.New()

func verifyPassword(incomingPassword, existingPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(existingPassword), []byte(incomingPassword))
	log.Println(existingPassword)
	log.Println(incomingPassword)
	log.Println("************************")
	check := true
	msg := ""
	if err != nil {
		check = false
		msg = "passwords do not match"
	}
	return check, msg
}
func HashPassword(pass string) string {
	password, err := bcrypt.GenerateFromPassword([]byte(pass), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(password)
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
		defer cancel()
		cursor, err := getCollection.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var users []model.User
		for cursor.Next(ctx) {
			var user model.User
			if err := cursor.Decode(&user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
				return
			}
			users = append(users, user)
		}

		if err := cursor.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to iterate over users"})
			return
		}

		c.JSON(http.StatusOK, users)
	}

}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Println("*******************")
		log.Println(userId)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
		defer cancel()
		var user model.User
		err := getCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)

	}

}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		var user model.User
		var existingUser model.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no data found"})
			return
		}
		passwordIsValid, msg := verifyPassword(user.Password, existingUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
		token, refreshToken, err := helper.GenerateAllTokens(existingUser.Email, existingUser.FirstName, existingUser.LastName, existingUser.UserType, existingUser.UserId)
		helper.UpdateAllTokens(token, refreshToken, existingUser.UserId)
		err = userCollection.FindOne(ctx, bson.M{"user_id": existingUser.UserId}).Decode(&existingUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, existingUser)
	}

}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		var user model.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Println("************")
		log.Println(user)
		log.Println("*************")
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			// log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "err while fetching documents"})
		}

		password := HashPassword(user.Password)
		user.Password = password
		log.Println("working fine")
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email already exists"})
			return
		}
		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Id = primitive.NewObjectID()
		user.UserId = user.Id.Hex()
		token, refreshToken, err := helper.GenerateAllTokens(user.Email, user.FirstName, user.LastName, user.UserType, user.UserId)
		user.Token = token
		user.RefreshToken = refreshToken
		insertedId, err := userCollection.InsertOne(ctx, user)
		if err != nil {
			msg := fmt.Sprintf("user item not inserted")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, insertedId)
	}

}

func AddUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		var user model.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		count, err := getCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "err while fetching documents"})
			return
		}

		password := HashPassword(user.Password)
		user.Password = password
		log.Println("working fine")
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email already exists"})
			return
		}
		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Id = primitive.NewObjectID()
		user.UserId = user.Id.Hex()
		insertedId, err := getCollection.InsertOne(ctx, user)
		if err != nil {
			msg := fmt.Sprintf("user item not inserted")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, insertedId)

	}
}
