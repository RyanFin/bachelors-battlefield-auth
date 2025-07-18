package handlers

import (
	"bachelors-battlefield-auth/pkg/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ✅ Option 1: Pass the database instance to the handlers (preferred for clean architecture)
type Handler struct {
	DB *mongo.Database
}

// Access the database instance through a pointer receiver, tying the func to the struct
func (h *Handler) GetLocations(c *gin.Context) {
	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var locations []models.Location
	if err = cursor.All(ctx, &locations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, locations)
}

func (h *Handler) GetLocation(c *gin.Context) {
	id := c.Param("id")
	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	var location models.Location
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&location)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, location)
}

func (h *Handler) CreateLocation(c *gin.Context) {
	var req models.CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	location := models.Location{
		Key:          req.Key,
		Name:         req.Name,
		Emoji:        req.Emoji,
		Difficulty:   req.Difficulty,
		Color:        "#4ECDC4", // Default color
		LocationData: req.LocationData,
		Votes:        []models.Vote{},
		Notes:        []models.Note{},
		PendingNotes: []models.Note{},
		IsApproved:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Add initial note if provided
	if req.Notes != "" {
		note := models.Note{
			ID:         primitive.NewObjectID(),
			Content:    req.Notes,
			CreatedAt:  time.Now(),
			IsApproved: false,
		}
		location.PendingNotes = append(location.PendingNotes, note)
	}

	result, err := collection.InsertOne(ctx, location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	location.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, location)
}

func (h *Handler) UpdateLocation(c *gin.Context) {
	id := c.Param("id")
	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updatedAt timestamp
	updates["updatedAt"] = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updates}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Location updated successfully"})
}

func (h *Handler) DeleteLocation(c *gin.Context) {
	id := c.Param("id")
	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Location deleted successfully"})
}

func (h *Handler) AddVote(c *gin.Context) {
	id := c.Param("id")
	var req models.VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	// First, try to increment existing vote
	filter := bson.M{
		"_id":              objectID,
		"votes.difficulty": req.Difficulty,
	}
	update := bson.M{
		"$inc": bson.M{"votes.$.count": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If no existing vote found, add new vote
	if result.MatchedCount == 0 {
		newVote := models.Vote{
			ID:         primitive.NewObjectID(),
			Difficulty: req.Difficulty,
			Count:      1,
		}

		filter = bson.M{"_id": objectID}
		update = bson.M{
			"$push": bson.M{"votes": newVote},
			"$set":  bson.M{"updatedAt": time.Now()},
		}

		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Add note if provided
	if req.Notes != "" {
		note := models.Note{
			ID:         primitive.NewObjectID(),
			Content:    req.Notes,
			CreatedAt:  time.Now(),
			IsApproved: false,
		}

		filter = bson.M{"_id": objectID}
		update = bson.M{
			"$push": bson.M{"pendingNotes": note},
			"$set":  bson.M{"updatedAt": time.Now()},
		}

		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote added successfully"})
}

func (h *Handler) GetVotes(c *gin.Context) {
	id := c.Param("id")
	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	var location models.Location
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&location)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, location.Votes)
}

func (h *Handler) AddNote(c *gin.Context) {
	id := c.Param("id")
	var req models.NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	note := models.Note{
		ID:         primitive.NewObjectID(),
		Content:    req.Content,
		CreatedAt:  time.Now(),
		IsApproved: false,
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$push": bson.M{"pendingNotes": note},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func (h *Handler) GetNotes(c *gin.Context) {
	id := c.Param("id")
	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	var location models.Location
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&location)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notes":        location.Notes,
		"pendingNotes": location.PendingNotes,
	})
}

func (h *Handler) DeleteNote(c *gin.Context) {
	id := c.Param("id")
	noteId := c.Param("noteId")
	collection := h.DB.Collection("locations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	noteObjectID, err := primitive.ObjectIDFromHex(noteId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	// Try to remove from notes first
	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$pull": bson.M{"notes": bson.M{"_id": noteObjectID}},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If not found in notes, try pendingNotes
	if result.ModifiedCount == 0 {
		update = bson.M{
			"$pull": bson.M{"pendingNotes": bson.M{"_id": noteObjectID}},
			"$set":  bson.M{"updatedAt": time.Now()},
		}

		result, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}
