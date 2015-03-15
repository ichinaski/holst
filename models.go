package main

type User struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Item struct {
	Id         string   `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	Categories []string `json:"categories,omitempty"`
}

type Link struct {
	Id     string `json:"id,omitempty"`
	UserId string `json:"userId,omitempty"`
	ItemId string `json:"itemId,omitempty"`
	Type   string `json:"type,omitempty"`  // Buy, Rate, View, etc. Makes sense to have a fixed set of values?
	Score  int    `json:"score,omitempty"` // Relationship score
	// TODO: Create user/item support. Optional fields.
}

type Recommendation struct {
	Item      Item `json:"item,omitempty"`
	Frequency int  `json:"frequency,omitempty"`
	Strength  int  `json:"strength,omitempty"` // Item frequency
}
