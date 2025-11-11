package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Stock struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ProductID string             `bson:"product_id"`
	Quantity  int32              `bson:"quantity"`
	Reserved  int32              `bson:"reserved"`
}

type ReserveRequest struct {
	ProductID string
	Quantity  int32
}
