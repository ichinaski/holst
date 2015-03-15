package main

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Item struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
}

type Link struct {
	Id     string `json:"id"`
	UserId string `json:"userId"`
	ItemId string `json:"itemId"`
	Type   string `json:"type"` // Buy, Rate, View, etc. Makes sense to have a fixed set of values?
	// TODO: Create user/item support. Optional fields.
}

type Recommendation struct {
	Item      Item `json:"item"`
	Frequency int  `json:"frequency"`
}
