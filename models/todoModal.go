package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Todo struct {
	ID          primitive.ObjectID `json:"id"`
	Title       string             `json:"title" validate:"required"`
	Description string             `json:"description" validate:"required"`
	User_id     string             `json:"user_id" validate:"required"`
	Check       bool               `json:"check"`
	Created_at  time.Time          `json:"created_at"`
	Updated_at  time.Time          `json:"updated_at"`
}

type UpdateTodo struct {
	Check   bool   `json:"check" validate:"required"`
	User_id string `json:"user_id" validate:"required"`
}


