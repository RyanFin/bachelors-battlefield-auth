package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Coordinates struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
}

type LocationData struct {
	Type        string      `json:"type" bson:"type"`               // "point" or "polygon"
	Coordinates interface{} `json:"coordinates" bson:"coordinates"` // Coordinates for point, []Coordinates for polygon
}

type Vote struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Difficulty string             `json:"difficulty" bson:"difficulty"`
	Count      int                `json:"count" bson:"count"`
}

type Note struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Content    string             `json:"content" bson:"content"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	IsApproved bool               `json:"isApproved" bson:"isApproved"`
}

type Location struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Key          string             `json:"key" bson:"key"`
	Name         string             `json:"name" bson:"name"`
	Emoji        string             `json:"emoji" bson:"emoji"`
	Difficulty   string             `json:"difficulty" bson:"difficulty"`
	Color        string             `json:"color" bson:"color"`
	LocationData LocationData       `json:"locationData" bson:"locationData"`
	Votes        []Vote             `json:"votes" bson:"votes"`
	Notes        []Note             `json:"notes" bson:"notes"`
	PendingNotes []Note             `json:"pendingNotes" bson:"pendingNotes"`
	IsApproved   bool               `json:"isApproved" bson:"isApproved"`
	ApprovedAt   *time.Time         `json:"approvedAt" bson:"approvedAt"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type CreateLocationRequest struct {
	Key          string       `json:"key" binding:"required"`
	Name         string       `json:"name" binding:"required"`
	Emoji        string       `json:"emoji"`
	Difficulty   string       `json:"difficulty"`
	LocationData LocationData `json:"locationData"`
	Notes        string       `json:"notes"`
}

type VoteRequest struct {
	Difficulty string `json:"difficulty" binding:"required"`
	Notes      string `json:"notes"`
}

type NoteRequest struct {
	Content string `json:"content" binding:"required"`
}
