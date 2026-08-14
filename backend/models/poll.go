package models

import "time"

type Poll struct {
	ID       string   `json:"id" bson:"_id,omitempty"`
	Question string   `json:"question" bson:"question"`
	Options  []Option `json:"options" bson:"options"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type Option struct {
	ID   string `json:"id" bson:"id"`
	Text string `json:"text" bson:"text"`
	Votes int    `json:"votes" bson:"votes"`
}