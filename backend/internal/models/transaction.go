package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"  json:"id"`
	UserID      primitive.ObjectID `bson:"user_id"        json:"user_id"`
	CategoryID  primitive.ObjectID `bson:"category_id"    json:"category_id"`
	Type        string             `bson:"type"           json:"type"`
	Amount      float64            `bson:"amount"         json:"amount"`
	Description string             `bson:"description"    json:"description"`
	Date        time.Time          `bson:"date"           json:"date"`
	CreatedAt   time.Time          `bson:"created_at"     json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"     json:"updated_at"`
}
