package main

// User represents a user in the system.
type User struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Item represents an item in the system.
type Item struct {
	Id         string   `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	Categories []string `json:"categories,omitempty"`
}

// Link represents the relationship between a user and an item.
type Link struct {
	Id     string `json:"id,omitempty"`
	UserId string `json:"userId,omitempty"`
	ItemId string `json:"itemId,omitempty"`
	Type   string `json:"type,omitempty"`  // Buy, Rate, View, etc. Makes sense to have a fixed set of values?
	Score  int    `json:"score,omitempty"` // Relationship score
	// TODO: Create user/item support. Optional fields.
}

// Recommendation represents the item being recommended to a user.
type Recommendation struct {
	Item     Item `json:"item,omitempty"`
	Strength int  `json:"strength,omitempty"` // Item frequency
}
