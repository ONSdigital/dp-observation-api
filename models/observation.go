package models

import "github.com/ONSdigital/dp-api-clients-go/dataset"

const wildcard = "*"

// ObservationsDoc represents information (observations) relevant to a version
type ObservationsDoc struct {
	Dimensions        map[string]Option    `json:"dimensions"`
	Limit             int                  `json:"limit"`
	Links             *ObservationLinks    `json:"links"`
	Observations      []Observation        `json:"observations"`
	Offset            int                  `json:"offset"`
	TotalObservations int                  `json:"total_observations"`
	UnitOfMeasure     string               `json:"unit_of_measure,omitempty"`
	UsageNotes        *[]dataset.UsageNote `json:"usage_notes,omitempty"`
}

// Observation represents an object containing a single
// observation and its equivalent metadata
type Observation struct {
	Dimensions  map[string]*DimensionObject `json:"dimensions,omitempty"`
	Metadata    map[string]string           `json:"metadata,omitempty"`
	Observation string                      `json:"observation"`
}

// DimensionObject represents the unique dimension option data relevant to the observation
type DimensionObject struct {
	HRef  string `json:"href"`
	ID    string `json:"id"`
	Label string `json:"label"`
}

// ObservationLinks represents a link object to list of links relevant to the observation
type ObservationLinks struct {
	DatasetMetadata *LinkObject `json:"dataset_metadata,omitempty"`
	Self            *LinkObject `json:"self,omitempty"`
	Version         *LinkObject `json:"version,omitempty"`
}

// Option represents an object containing a list of link objects that refer to the
// code url for that dimension option
type Option struct {
	LinkObject *LinkObject `json:"option,omitempty"`
}

// CreateObservationsDoc manages the creation of metadata across dataset and version docs
func CreateObservationsDoc(rawQuery string, versionDoc *dataset.Version, unitOfMeasure string, usageNotes *[]dataset.UsageNote, observations []Observation, queryParameters map[string]string, offset, limit int) *ObservationsDoc {

	observationsDoc := &ObservationsDoc{
		Limit: limit,
		Links: &ObservationLinks{
			DatasetMetadata: &LinkObject{
				HRef: versionDoc.Links.Version.URL + "/metadata",
			},
			Self: &LinkObject{
				HRef: versionDoc.Links.Version.URL + "/observations?" + rawQuery,
			},
			Version: &LinkObject{
				HRef: versionDoc.Links.Version.URL,
				ID:   versionDoc.Links.Version.ID,
			},
		},
		Observations:      observations,
		Offset:            offset,
		TotalObservations: len(observations),
		UnitOfMeasure:     unitOfMeasure, // TODO dataset api should return this?
		UsageNotes:        usageNotes,    // TODO dataset api should return this?
	}

	var dimensions = make(map[string]Option)

	// add the dimension codes
	for paramKey, paramValue := range queryParameters {
		for _, dimension := range versionDoc.Dimensions {
			var linkObjects []*LinkObject
			if dimension.Name == paramKey && paramValue != wildcard {

				linkObject := &LinkObject{
					HRef: dimension.URL + "/codes/" + paramValue,
					ID:   paramValue,
				}
				linkObjects = append(linkObjects, linkObject)
				dimensions[paramKey] = Option{
					LinkObject: linkObject,
				}
				break
			}
		}
	}
	observationsDoc.Dimensions = dimensions

	return observationsDoc
}
