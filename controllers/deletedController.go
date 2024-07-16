package controllers

import (
	"context"
	"net/http"
	"nitiwat/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetAllDeleted() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := deleteCollections.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var deletedData []models.DeleteModal

		if err = cursor.All(ctx, &deletedData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": deletedData})

	}
}

func GetDeletedById() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		del_idParam := c.Param("del_id")
		delId, err := primitive.ObjectIDFromHex(del_idParam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "no ID found"})
			return
		}

		var delData models.DeleteModal

		err = deleteCollections.FindOne(ctx, bson.M{"id": delId}).Decode(&delData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": delId})

	}
}
