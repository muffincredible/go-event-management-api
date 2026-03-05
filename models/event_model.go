package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	Id           primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Title        string               `json:"title,omitempty" validate:"required"`
	Description  string               `json:"description,omitempty"`
	Date         time.Time            `json:"date,omitempty" validate:"required"`
	Location     string               `json:"location,omitempty"`
	Capacity     int                  `json:"capacity,omitempty" validate:"required"`
	CreatorId    primitive.ObjectID   `json:"creator_id,omitempty" bson:"creator_id,omitempty"`
	Participants []primitive.ObjectID `json:"participants,omitempty" bson:"participants,omitempty"`
}