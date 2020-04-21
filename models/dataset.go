package models

import "time"

// DatasetUpdate represents an evolving dataset with the current dataset and the updated dataset
type DatasetUpdate struct {
	ID      string   `bson:"_id,omitempty"         json:"id,omitempty"`
	Current *Dataset `bson:"current,omitempty"     json:"current,omitempty"`
	Next    *Dataset `bson:"next,omitempty"        json:"next,omitempty"`
}

// Dataset represents information related to a single dataset
type Dataset struct {
	CollectionID      string           `bson:"collection_id,omitempty"          json:"collection_id,omitempty"`
	Contacts          []ContactDetails `bson:"contacts,omitempty"               json:"contacts,omitempty"`
	Description       string           `bson:"description,omitempty"            json:"description,omitempty"`
	Keywords          []string         `bson:"keywords,omitempty"               json:"keywords,omitempty"`
	ID                string           `bson:"_id,omitempty"                    json:"id,omitempty"`
	LastUpdated       time.Time        `bson:"last_updated,omitempty"           json:"-"`
	License           string           `bson:"license,omitempty"                json:"license,omitempty"`
	Links             *DatasetLinks    `bson:"links,omitempty"                  json:"links,omitempty"`
	Methodologies     []GeneralDetails `bson:"methodologies,omitempty"          json:"methodologies,omitempty"`
	NationalStatistic *bool            `bson:"national_statistic,omitempty"     json:"national_statistic,omitempty"`
	NextRelease       string           `bson:"next_release,omitempty"           json:"next_release,omitempty"`
	Publications      []GeneralDetails `bson:"publications,omitempty"           json:"publications,omitempty"`
	Publisher         *Publisher       `bson:"publisher,omitempty"              json:"publisher,omitempty"`
	QMI               *GeneralDetails  `bson:"qmi,omitempty"                    json:"qmi,omitempty"`
	RelatedDatasets   []GeneralDetails `bson:"related_datasets,omitempty"       json:"related_datasets,omitempty"`
	ReleaseFrequency  string           `bson:"release_frequency,omitempty"      json:"release_frequency,omitempty"`
	State             string           `bson:"state,omitempty"                  json:"state,omitempty"`
	Theme             string           `bson:"theme,omitempty"                  json:"theme,omitempty"`
	Title             string           `bson:"title,omitempty"                  json:"title,omitempty"`
	UnitOfMeasure     string           `bson:"unit_of_measure,omitempty"        json:"unit_of_measure,omitempty"`
	URI               string           `bson:"uri,omitempty"                    json:"uri,omitempty"`
}

// DatasetLinks represents a list of specific links related to the dataset resource
type DatasetLinks struct {
	AccessRights  *LinkObject `bson:"access_rights,omitempty"   json:"access_rights,omitempty"`
	Editions      *LinkObject `bson:"editions,omitempty"        json:"editions,omitempty"`
	LatestVersion *LinkObject `bson:"latest_version,omitempty"  json:"latest_version,omitempty"`
	Self          *LinkObject `bson:"self,omitempty"            json:"self,omitempty"`
	Taxonomy      *LinkObject `bson:"taxonomy,omitempty"        json:"taxonomy,omitempty"`
}

// LinkObject represents a generic structure for all links
type LinkObject struct {
	HRef string `bson:"href,omitempty"  json:"href,omitempty"`
	ID   string `bson:"id,omitempty"    json:"id,omitempty"`
}

// GeneralDetails represents generic fields stored against an object (reused)
type GeneralDetails struct {
	Description string `bson:"description,omitempty"    json:"description,omitempty"`
	HRef        string `bson:"href,omitempty"           json:"href,omitempty"`
	Title       string `bson:"title,omitempty"          json:"title,omitempty"`
}

// ContactDetails represents an object containing information of the contact
type ContactDetails struct {
	Email     string `bson:"email,omitempty"      json:"email,omitempty"`
	Name      string `bson:"name,omitempty"       json:"name,omitempty"`
	Telephone string `bson:"telephone,omitempty"  json:"telephone,omitempty"`
}

// Publisher represents an object containing information of the publisher
type Publisher struct {
	HRef string `bson:"href,omitempty" json:"href,omitempty"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	Type string `bson:"type,omitempty" json:"type,omitempty"`
}

// Version represents information related to a single version for an edition of a dataset
type Version struct {
	Alerts        *[]Alert             `bson:"alerts,omitempty"         json:"alerts,omitempty"`
	CollectionID  string               `bson:"collection_id,omitempty"  json:"collection_id,omitempty"`
	Dimensions    []Dimension          `bson:"dimensions,omitempty"     json:"dimensions,omitempty"`
	Downloads     *DownloadList        `bson:"downloads,omitempty"      json:"downloads,omitempty"`
	Edition       string               `bson:"edition,omitempty"        json:"edition,omitempty"`
	Headers       []string             `bson:"headers,omitempty"        json:"-"`
	ID            string               `bson:"id,omitempty"             json:"id,omitempty"`
	LastUpdated   time.Time            `bson:"last_updated,omitempty"   json:"-"`
	LatestChanges *[]LatestChange      `bson:"latest_changes,omitempty" json:"latest_changes,omitempty"`
	Links         *VersionLinks        `bson:"links,omitempty"          json:"links,omitempty"`
	ReleaseDate   string               `bson:"release_date,omitempty"   json:"release_date,omitempty"`
	State         string               `bson:"state,omitempty"          json:"state,omitempty"`
	Temporal      *[]TemporalFrequency `bson:"temporal,omitempty"       json:"temporal,omitempty"`
	UsageNotes    *[]UsageNote         `bson:"usage_notes,omitempty"    json:"usage_notes,omitempty"`
	Version       int                  `bson:"version,omitempty"        json:"version,omitempty"`
}

// Alert represents an object containing information on an alert
type Alert struct {
	Date        string `bson:"date,omitempty"        json:"date,omitempty"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	Type        string `bson:"type,omitempty"        json:"type,omitempty"`
}

// DownloadList represents a list of objects of containing information on the downloadable files
type DownloadList struct {
	CSV  *DownloadObject `bson:"csv,omitempty" json:"csv,omitempty"`
	CSVW *DownloadObject `bson:"csvw,omitempty" json:"csvw,omitempty"`
	XLS  *DownloadObject `bson:"xls,omitempty" json:"xls,omitempty"`
}

// DownloadObject represents information on the downloadable file
type DownloadObject struct {
	HRef    string `bson:"href,omitempty"  json:"href,omitempty"`
	Private string `bson:"private,omitempty" json:"private,omitempty"`
	Public  string `bson:"public,omitempty" json:"public,omitempty"`
	// TODO size is in bytes and probably should be an int64 instead of a string this
	// will have to change for several services (filter API, exporter services and web)
	Size string `bson:"size,omitempty" json:"size,omitempty"`
}

// LatestChange represents an object contining
// information on a single change between versions
type LatestChange struct {
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	Name        string `bson:"name,omitempty"        json:"name,omitempty"`
	Type        string `bson:"type,omitempty"        json:"type,omitempty"`
}

// TemporalFrequency represents a frequency for a particular period of time
type TemporalFrequency struct {
	EndDate   string `bson:"end_date,omitempty"    json:"end_date,omitempty"`
	Frequency string `bson:"frequency,omitempty"   json:"frequency,omitempty"`
	StartDate string `bson:"start_date,omitempty"  json:"start_date,omitempty"`
}

// UsageNote represents a note containing extra information associated to the resource
type UsageNote struct {
	Note  string `bson:"note,omitempty"     json:"note,omitempty"`
	Title string `bson:"title,omitempty"    json:"title,omitempty"`
}

// VersionLinks represents a list of specific links related to the version resource for an edition of a dataset
type VersionLinks struct {
	Dataset    *LinkObject `bson:"dataset,omitempty"     json:"dataset,omitempty"`
	Dimensions *LinkObject `bson:"dimensions,omitempty"  json:"dimensions,omitempty"`
	Edition    *LinkObject `bson:"edition,omitempty"     json:"edition,omitempty"`
	Self       *LinkObject `bson:"self,omitempty"        json:"self,omitempty"`
	Spatial    *LinkObject `bson:"spatial,omitempty"     json:"spatial,omitempty"`
	Version    *LinkObject `bson:"version,omitempty"     json:"-"`
}
