package models

import "github.com/ONSdigital/dp-api-clients-go/v2/dataset"

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
	DatasetMetadata *dataset.Link `json:"dataset_metadata,omitempty"`
	Self            *dataset.Link `json:"self,omitempty"`
	Version         *dataset.Link `json:"version,omitempty"`
}

// Option represents an object containing a list of link objects that refer to the
// code url for that dimension option
type Option struct {
	LinkObject *dataset.Link `json:"option,omitempty"`
}

// FilterSubmitted is the structure of each event consumed.
type FilterSubmitted struct {
	FilterID   string `avro:"filter_output_id"`
	InstanceID string `avro:"instance_id"`
	DatasetID  string `avro:"dataset_id"`
	Edition    string `avro:"edition"`
	Version    string `avro:"version"`
}

// CreateObservationsDoc manages the creation of metadata across dataset and version docs
func CreateObservationsDoc(obsAPIURL, rawQuery string, versionDoc *dataset.Version, datasetDetails dataset.DatasetDetails, observations []Observation, queryParameters map[string]string, offset, limit int) *ObservationsDoc {

	selfLink := generateSelfURL(obsAPIURL, rawQuery, versionDoc)

	observationsDoc := &ObservationsDoc{
		Limit: limit,
		Links: &ObservationLinks{
			DatasetMetadata: &dataset.Link{
				URL: versionDoc.Links.Version.URL + "/metadata",
			},
			Self: &dataset.Link{
				URL: selfLink,
			},
			Version: &dataset.Link{
				URL: versionDoc.Links.Version.URL,
				ID:  versionDoc.Links.Version.ID,
			},
		},
		Observations:      observations,
		Offset:            offset,
		TotalObservations: len(observations),
		UnitOfMeasure:     datasetDetails.UnitOfMeasure,
		UsageNotes:        datasetDetails.UsageNotes,
	}

	var dimensions = make(map[string]Option)

	versionDocDimensions := make(map[string]dataset.VersionDimension)
	for _, dim := range versionDoc.Dimensions {
		versionDocDimensions[dim.Name] = dim
	}

	// add the dimension codes
	for paramKey, paramValue := range queryParameters {
		dimension, found := versionDocDimensions[paramKey]
		if found && paramValue != wildcard {
			var linkObjects []*dataset.Link

			linkObject := &dataset.Link{
				URL: dimension.URL + "/codes/" + paramValue,
				ID:  paramValue,
			}
			linkObjects = append(linkObjects, linkObject)
			dimensions[paramKey] = Option{
				LinkObject: linkObject,
			}
		}
	}
	observationsDoc.Dimensions = dimensions

	return observationsDoc
}

func generateSelfURL(obsAPIURL, rawQuery string, versionDoc *dataset.Version) string {
	return obsAPIURL + "/datasets/" + versionDoc.Links.Dataset.ID + "/editions/" +
		versionDoc.Links.Edition.ID + "/versions/" + versionDoc.Links.Version.ID + "/observations?" + rawQuery
}
