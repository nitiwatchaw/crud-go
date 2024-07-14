package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"nitiwat/database"
	helper "nitiwat/helpers"
	"nitiwat/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollections *mongo.Collection = database.OpenCollection(database.Client, "users")
var validate = validator.New()

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(hashedPassword)

}

func VerifyPassword(userPassWord string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassWord))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("Password is incorrect")
		check = false
	}
	return check, msg

}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		countEmail, err := userCollections.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while checking for the email"})
		}

		if countEmail > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Email or phone number already exists"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		countPhone, err := userCollections.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while checking for the phone number"})
		}

		if countPhone > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Email or phone number already exists"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken
		resultInsertionNumber, insertErr := userCollections.InsertOne(ctx, user)
		if insertErr != nil {
			msg := "User not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return

		}
		defer cancel()

		c.JSON(http.StatusOK, gin.H{"data": resultInsertionNumber})

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollections.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect "})
			return
		}
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		defer cancel()

		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not found"})
		}
		// helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		userCollections.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": foundUser})
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerpage, err := strconv.Atoi(c.Query("recordPerpage"))

		if err != nil || recordPerpage < 1 {
			recordPerpage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1

		}

		startIndex := (page - 1) * recordPerpage

		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{primitive.E{
			Key:   "$match",
			Value: bson.D{},
		},
		}
		groupStage := bson.D{primitive.E{Key: "$group", Value: bson.D{
			primitive.E{
				Key: "_id",
				Value: bson.D{
					primitive.E{
						Key:   "_id",
						Value: "null",
					},
				},
			},
			primitive.E{
				Key: "total_count",
				Value: bson.D{
					primitive.E{
						Key:   "$sum",
						Value: 1,
					},
				},
			},
			primitive.E{
				Key: "data",
				Value: bson.D{
					primitive.E{
						Key:   "$push",
						Value: "$$ROOT",
					},
				},
			},
		}}}
		projectStage := bson.D{
			primitive.E{
				Key: "$project",
				Value: bson.D{
					primitive.E{
						Key:   "_id",
						Value: 0},
					primitive.E{
						Key:   "total_count",
						Value: 1},
					primitive.E{
						Key: "user_items",
						Value: bson.D{
							primitive.E{
								Key:   "$slice",
								Value: []interface{}{"$data", startIndex, recordPerpage},
							},
						},
					},
				},
			},
		}
		result, err := userCollections.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while fetching users"})
		}

		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)

		}
		c.JSON(http.StatusOK, allUsers[0])

	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// c.JSON(200, gin.H{
		// 	"message": "userId",
		// })
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := userCollections.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": user})

	}

}
