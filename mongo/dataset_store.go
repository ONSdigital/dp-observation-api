package mongo

import (
	"errors"
	"strconv"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	errs "github.com/ONSdigital/dp-observation-api/apierrors"
	"github.com/ONSdigital/dp-observation-api/models"
)

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	// CodeListURL    string
	Collection string
	Database   string
	// DatasetURL     string
	Session        *mgo.Session
	URI            string
	lastPingTime   time.Time // TODO add ping?
	lastPingResult error
}

const (
	editionsCollection = "editions"
)

// Init creates a new mgo.Session with a strong consistency and a write mode of "majortiy".
func (m *Mongo) Init() (session *mgo.Session, err error) {
	if session != nil {
		return nil, errors.New("session already exists")
	}

	if session, err = mgo.Dial(m.URI); err != nil {
		return nil, err
	}

	session.EnsureSafe(&mgo.Safe{WMode: "majority"})
	session.SetMode(mgo.Strong, true)
	return session, nil
}

// GetDataset retrieves a dataset document
func (m *Mongo) GetDataset(id string) (*models.DatasetUpdate, error) {
	s := m.Session.Copy()
	defer s.Close()
	var dataset models.DatasetUpdate
	err := s.DB(m.Database).C("datasets").Find(bson.M{"_id": id}).One(&dataset)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errs.ErrDatasetNotFound
		}
		return nil, err
	}

	return &dataset, nil
}

// GetVersion retrieves a version document for a dataset edition
func (m *Mongo) GetVersion(id, editionID, versionID, state string) (*models.Version, error) {
	s := m.Session.Copy()
	defer s.Close()

	versionNumber, err := strconv.Atoi(versionID)
	if err != nil {
		return nil, err
	}
	selector := buildVersionQuery(id, editionID, state, versionNumber)

	var version models.Version
	err = s.DB(m.Database).C("instances").Find(selector).One(&version)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errs.ErrVersionNotFound
		}
		return nil, err
	}
	return &version, nil
}

// CheckEditionExists checks that the edition of a dataset exists
func (m *Mongo) CheckEditionExists(id, editionID, state string) error {
	s := m.Session.Copy()
	defer s.Close()

	var query bson.M
	if state == "" {
		query = bson.M{
			"next.links.dataset.id": id,
			"next.edition":          editionID,
		}
	} else {
		query = bson.M{
			"current.links.dataset.id": id,
			"current.edition":          editionID,
			"current.state":            state,
		}
	}

	count, err := s.DB(m.Database).C(editionsCollection).Find(query).Count()
	if err != nil {
		return err
	}

	if count == 0 {
		return errs.ErrEditionNotFound
	}

	return nil
}

func buildVersionQuery(id, editionID, state string, versionID int) bson.M {
	var selector bson.M
	if state != models.PublishedState {
		selector = bson.M{
			"links.dataset.id": id,
			"version":          versionID,
			"edition":          editionID,
		}
	} else {
		selector = bson.M{
			"links.dataset.id": id,
			"edition":          editionID,
			"version":          versionID,
			"state":            state,
		}
	}

	return selector
}
