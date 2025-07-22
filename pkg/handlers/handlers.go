package handlers

import (
	"bachelors-battlefield-auth/pkg/models"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler struct to hold database instance
type Handler struct {
	DB *mongo.Database
}

// Request structures for vote operations
type VoteSubmission struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	LocationID string             `json:"locationId" bson:"locationId"`
	Difficulty string             `json:"difficulty" bson:"difficulty"`
	Notes      string             `json:"notes" bson:"notes"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	IsApproved bool               `json:"isApproved" bson:"isApproved"`
	ApprovedAt *time.Time         `json:"approvedAt" bson:"approvedAt"`
}

type VoteApprovalRequest struct {
	LocationID string   `json:"locationId" binding:"required"`
	Difficulty string   `json:"difficulty" binding:"required"`
	Notes      []string `json:"notes"`
}

type VoteRejectRequest struct {
	LocationID string `json:"locationId" binding:"required"`
}

type VoteRequestBody struct {
	LocationID string `json:"locationId" binding:"required"`
	Difficulty string `json:"difficulty" binding:"required"`
	Notes      string `json:"notes"`
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
	if len(req.Notes) > 0 {
		for _, reqNote := range req.Notes {
			note := models.Note{
				ID:         primitive.NewObjectID(),
				Content:    reqNote.Content,
				CreatedAt:  time.Now(),
				IsApproved: false,
			}
			location.PendingNotes = append(location.PendingNotes, note)
		}
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

// Vote-related handlers from the first file

func (h *Handler) AddVote(c *gin.Context) {
	var voteReq VoteRequestBody
	if err := c.ShouldBindJSON(&voteReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create vote submission
	vote := VoteSubmission{
		ID:         primitive.NewObjectID(),
		LocationID: voteReq.LocationID,
		Difficulty: voteReq.Difficulty,
		Notes:      voteReq.Notes,
		CreatedAt:  time.Now(),
		IsApproved: false,
	}

	collection := h.DB.Collection("votes")
	_, err := collection.InsertOne(ctx, vote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add vote"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vote submitted successfully",
		"vote":    vote,
	})
}

func (h *Handler) GetPendingVotes(c *gin.Context) {
	collection := h.DB.Collection("votes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"isApproved": false}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending votes"})
		return
	}
	defer cursor.Close(ctx)

	var votes []VoteSubmission
	if err = cursor.All(ctx, &votes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode votes"})
		return
	}

	c.JSON(http.StatusOK, votes)
}

func (h *Handler) ApproveVotes(c *gin.Context) {
	var request VoteApprovalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// Approve votes for the location in votes collection
	votesCollection := h.DB.Collection("votes")
	filter := bson.M{"locationId": request.LocationID, "isApproved": false}
	update := bson.M{
		"$set": bson.M{
			"isApproved": true,
			"approvedAt": now,
		},
	}

	_, err := votesCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve votes"})
		return
	}

	// Update location with new difficulty and notes
	locationsCollection := h.DB.Collection("locations")
	var locationFilter bson.M

	// Try to convert location ID to ObjectID, otherwise use as string
	if objectID, err := primitive.ObjectIDFromHex(request.LocationID); err == nil {
		locationFilter = bson.M{"_id": objectID}
	} else {
		locationFilter = bson.M{"key": request.LocationID}
	}

	// Create approved notes from the approved notes list
	var approvedNotes []models.Note
	for _, noteContent := range request.Notes {
		if strings.TrimSpace(noteContent) != "" {
			approvedNotes = append(approvedNotes, models.Note{
				ID:         primitive.NewObjectID(),
				Content:    strings.TrimSpace(noteContent),
				IsApproved: true,
				CreatedAt:  now,
			})
		}
	}

	locationUpdate := bson.M{
		"$set": bson.M{
			"difficulty": request.Difficulty,
			"isApproved": true,
			"approvedAt": now,
			"updatedAt":  now,
		},
		"$unset": bson.M{
			"pendingNotes": "",
		},
	}

	// Add notes if any exist
	if len(approvedNotes) > 0 {
		locationUpdate["$push"] = bson.M{
			"notes": bson.M{"$each": approvedNotes},
		}
	}

	result, err := locationsCollection.UpdateOne(ctx, locationFilter, locationUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update location: " + err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "approved",
		"message":    "Votes approved and location updated successfully",
		"updated":    result.ModifiedCount > 0,
		"notesAdded": len(approvedNotes),
	})
}

func (h *Handler) RejectVotes(c *gin.Context) {
	var request VoteRejectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Delete pending votes for the location
	collection := h.DB.Collection("votes")
	filter := bson.M{"locationId": request.LocationID, "isApproved": false}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject votes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "rejected",
		"message":      "Votes rejected successfully",
		"deletedCount": result.DeletedCount,
	})
}

// Existing note-related handlers (preserved from original)

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
