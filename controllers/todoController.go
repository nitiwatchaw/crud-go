package controllers

import (
	"context"
	"fmt"
	"net/http"
	"nitiwat/database"
	helper "nitiwat/helpers"
	"nitiwat/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var todoCollections *mongo.Collection = database.OpenCollection(database.Client, "todos")

func GetTodo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := todoCollections.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var todos []bson.M
		fmt.Println("cursor", cursor)
		if err = cursor.All(ctx, &todos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": todos})
	}
}

func GetTodoById() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		todoIDParam := c.Param("todo_id")
		todoID, err := primitive.ObjectIDFromHex(todoIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID format"})
			return
		}

		var todo models.Todo
		err = todoCollections.FindOne(ctx, bson.M{"id": todoID}).Decode(&todo)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accessing the database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": todo})
	}
}

func AddTodo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var todo models.Todo
		var foundTodo models.Todo
		var foundUser models.User

		if err := c.BindJSON(&todo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//find todo by title
		filter := bson.M{"title": todo.Title, "user_id": todo.User_id}
		errTodo := todoCollections.FindOne(ctx, filter).Decode(&foundTodo)
		defer cancel()
		if errTodo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "title is exist on database"})
			return
		}

		err := userCollections.FindOne(ctx, bson.M{"user_id": todo.User_id}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ID is not exist on database"})
			return
		}

		todo.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		todo.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		todo.ID = primitive.NewObjectID()
		todo.User_id = foundUser.User_id
		todo.Check = false
		resultInsertionTodo, insertErr := todoCollections.InsertOne(ctx, todo)

		if insertErr != nil {
			msg := "Todo not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return

		}
		defer cancel()

		c.JSON(http.StatusOK, gin.H{"data": resultInsertionTodo})

	}
}

func DeleteTodo() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		todoIDParam := c.Param("todo_id")
		todoID, err := primitive.ObjectIDFromHex(todoIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID format"})
			return
		}

		var todo models.Todo
		//check if have todo id in database
		errTodo := todoCollections.FindOne(ctx, bson.M{"id": todoID}).Decode(&todo)
		if errTodo != nil {
			if errTodo == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accessing the database"})
			}
			return
		}

		result := todoCollections.FindOneAndDelete(ctx, bson.M{"id": todoID})
		if result != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting the todo"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": todoIDParam + "todo deleted successfully"})
	}
}

func UpdateCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var updateTodo models.UpdateTodo
		if err := c.BindJSON(&updateTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		todoIDParam := c.Param("todo_id")
		todoID, err := primitive.ObjectIDFromHex(todoIDParam)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID format"})
			return
		}

		//check if the param and todo id is match

		var todo models.Todo
		err = todoCollections.FindOne(ctx, bson.M{"id": todoID, "user_id": updateTodo.User_id}).Decode(&todo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Todo or user_id is not match"})
			return
		}

		filter := bson.M{"id": todoID}
		update := bson.M{"$set": bson.M{"check": updateTodo.Check}}
		_, err = todoCollections.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the todo"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": "update check successfully"})
	}
}

func UpdateEditTodo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var updateTodo models.Todo

		if err := c.BindJSON(&updateTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID format"})
			return
		}

		todoIDParam := c.Param("todo_id")
		todoID, errParam := primitive.ObjectIDFromHex(todoIDParam)
		if errParam != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params format"})
			return
		}

		// Check if the param and todo id match
		var todo models.Todo
		err := todoCollections.FindOne(ctx, bson.M{"id": todoID, "user_id": updateTodo.User_id}).Decode(&todo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Todo or user_id does not match"})
			return
		}

		// Update the todo item
		update := bson.M{
			"$set": bson.M{
				"title":       updateTodo.Title,
				"description": updateTodo.Description,
				"updated_at":  time.Now(),
			},
		}

		result, err := todoCollections.UpdateOne(ctx, bson.M{"id": todoID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating todo"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No todo found to update"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
	}
}

func GetTodoByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID := c.Param("user_id")
		var todos []bson.M

		// Get the total count of todos for the user
		count, err := todoCollections.CountDocuments(ctx, bson.M{"user_id": userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		//filter that match user_id
		pipeline := mongo.Pipeline{
			bson.D{{"$match", bson.D{{"user_id", userID}}}},
		}

		cursor, err := todoCollections.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err = cursor.All(ctx, &todos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Include the count in the response
		c.JSON(http.StatusOK, gin.H{"data": todos, "total_count": count})
	}
}

func CheckALlTodoActive() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var todos []bson.M

		pipeline := mongo.Pipeline{
			bson.D{
				{"$facet", bson.D{
					{"details", bson.A{bson.D{{"$match", bson.D{{"check", true}}}}}}, // Correctly placed
					{"totalCount", bson.A{bson.D{{"$count", "total_count"}}}},        // Correctly formatted
				}},
			},
		}

		cursor, err := todoCollections.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err = cursor.All(ctx, &todos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": todos})

	}
}

func FindQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var todos []bson.M

		pipeline := bson.A{
			bson.D{{"$match", bson.D{{"check", false}}}},
			bson.D{
				{"$group",
					bson.D{
						{"_id", primitive.Null{}},
						{"check", bson.D{{"$sum", 1}}},
					},
				},
			},
			bson.D{
				{"$project",
					bson.D{
						{"_id", 0},
						{"check", 1},
					},
				},
			},
		}

		cursor, err := todoCollections.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err = cursor.All(ctx, &todos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": todos})
	}
}
