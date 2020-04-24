package mongo

import (
	"context"
	"errors"
	"strconv"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	mongolib "github.com/ONSdigital/dp-mongodb"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongoHealth "github.com/ONSdigital/dp-mongodb/health"
	errs "github.com/ONSdigital/dp-observation-api/apierrors"
	"github.com/ONSdigital/dp-observation-api/models"
)

// Mongo represents a simplistic MongoDB configuration, and holds the session and healthclient
type Mongo struct {
	Collection   string
	Database     string
	URI          string
	session      *mgo.Session
	healthClient *mongoHealth.CheckMongoClient
}

const (
	editionsCollection = "editions"
)

// Init creates a new mgo.Session for this Mongo object, with a strong consistency and a write mode of "majortiy".
func (m *Mongo) Init() (s *mgo.Session, err error) {
	if m.session != nil {
		return nil, errors.New("session already exists")
	}

	// Create session
	if m.session, err = mgo.Dial(m.URI); err != nil {
		return nil, err
	}
	m.session.EnsureSafe(&mgo.Safe{WMode: "majority"})
	m.session.SetMode(mgo.Strong, true)

	// Create healthclient from client using the created session
	client := mongoHealth.NewClient(m.session)
	m.healthClient = &mongoHealth.CheckMongoClient{
		Client:      *client,
		Healthcheck: client.Healthcheck,
	}

	return m.session, nil
}

// Session is a getter for the MongoDB session
func (m *Mongo) Session() *mgo.Session {
	return m.session
}

// Close closes teh mongoDB session using dp-mongo library
func (m *Mongo) Close(ctx context.Context) error {
	return mongolib.Close(ctx, m.session)
}

// Checker calls the mongo health client Checker
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if m.healthClient == nil {
		return errors.New("client not initialised")
	}
	return m.healthClient.Checker(ctx, state)
}

// GetDataset retrieves a dataset document
func (m *Mongo) GetDataset(id string) (*models.DatasetUpdate, error) {
	s := m.session.Copy()
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
	s := m.session.Copy()
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
	s := m.session.Copy()
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
