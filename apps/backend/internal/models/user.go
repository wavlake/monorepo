package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          string    `json:"id" firestore:"id"`
	Email       string    `json:"email" firestore:"email"`
	DisplayName string    `json:"displayName" firestore:"displayName"`
	ProfilePic  string    `json:"profilePic" firestore:"profilePic"`
	NostrPubkey string    `json:"nostrPubkey" firestore:"nostrPubkey"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt"`
}

// Track represents a music track
type Track struct {
	ID          string    `json:"id" firestore:"id"`
	Title       string    `json:"title" firestore:"title"`
	Artist      string    `json:"artist" firestore:"artist"`
	Album       string    `json:"album,omitempty" firestore:"album"`
	Duration    int       `json:"duration" firestore:"duration"` // in seconds
	AudioURL    string    `json:"audioUrl" firestore:"audioUrl"`
	ArtworkURL  string    `json:"artworkUrl,omitempty" firestore:"artworkUrl"`
	Genre       string    `json:"genre,omitempty" firestore:"genre"`
	PriceMsat   int64     `json:"priceMsat,omitempty" firestore:"priceMsat"`
	OwnerID     string    `json:"ownerId" firestore:"ownerId"`
	NostrEventID string   `json:"nostrEventId,omitempty" firestore:"nostrEventId"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt"`
}