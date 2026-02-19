package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty"      json:"id"`
	UserID    *primitive.ObjectID `bson:"user_id,omitempty"  json:"user_id,omitempty"`
	Name      string              `bson:"name"               json:"name"`
	Icon      string              `bson:"icon"               json:"icon"`
	Color     string              `bson:"color"              json:"color"`
	IsDefault bool                `bson:"is_default"         json:"is_default"`
	CreatedAt time.Time           `bson:"created_at"         json:"created_at"`
}
