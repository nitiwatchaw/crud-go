package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type DeleteModal struct {
	ID    primitive.ObjectID `json:"id"`
	User  User
	Todos []Todo
}
