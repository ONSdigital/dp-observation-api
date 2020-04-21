package models

import "time"

// Dimension represents an overview for a single dimension. This includes a link to the code list API
// which provides metadata about the dimension and all possible values.
type Dimension struct {
	Description string        `bson:"description,omitempty"   json:"description,omitempty"`
	Label       string        `bson:"label,omitempty"         json:"label,omitempty"`
	LastUpdated time.Time     `bson:"last_updated,omitempty"  json:"-"`
	Links       DimensionLink `bson:"links,omitempty"         json:"links,omitempty"`
	HRef        string        `json:"href,omitempty"`
	ID          string        `json:"id,omitempty"`
	Name        string        `bson:"name,omitempty"          json:"name,omitempty"`
}

// DimensionLink contains all links needed for a dimension
type DimensionLink struct {
	CodeList LinkObject `bson:"code_list,omitempty"     json:"code_list,omitempty"`
	Options  LinkObject `bson:"options,omitempty"       json:"options,omitempty"`
	Version  LinkObject `bson:"version,omitempty"       json:"version,omitempty"`
}
