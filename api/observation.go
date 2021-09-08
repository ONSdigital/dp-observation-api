package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ONSdigital/dp-net/request"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-graph/v2/observation"
	"github.com/ONSdigital/dp-observation-api/apierrors"
	errs "github.com/ONSdigital/dp-observation-api/apierrors"
	"github.com/ONSdigital/dp-observation-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const (
	defaultOffset = 0
)

var (
	observationNotFound = map[error]bool{
		errs.ErrDatasetNotFound:      true,
		errs.ErrEditionNotFound:      true,
		errs.ErrVersionNotFound:      true,
		errs.ErrObservationsNotFound: true,
	}

	observationBadRequest = map[error]bool{
		errs.ErrTooManyWildcards: true,
	}
)

func (api *API) getObservations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	datasetID := vars["dataset_id"]
	edition := vars["edition"]
	version := vars["version"]

	// TODO call audit (attempt) once it has its own library
	logData := log.Data{"dataset_id": datasetID, "edition": edition, "version": version}

	observationsDoc, err := api.doGetObservations(ctx, datasetID, edition, version, r, logData)
	if err != nil {
		// TODO call audit (unsuccessful) once it has its own library
		handleObservationsErrorType(ctx, w, err, logData)
		return
	}

	// TODO call audit (successful) once it has its own library

	setJSONContentType(w)

	// The ampersand "&" is escaped to "\u0026" to keep some browsers from
	// misinterpreting JSON output as HTML. This escaping can be disabled using
	// an Encoder that had SetEscapeHTML(false) called on it.
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	if err = enc.Encode(observationsDoc); err != nil {
		handleObservationsErrorType(ctx, w, errors.WithMessage(err, "failed to marshal metadata resource into bytes"), logData)
		return
	}

	log.Info(ctx, "get observations endpoint: successfully retrieved observations relative to a selected set of dimension options for a version", logData)
}

func (api *API) doGetObservations(ctx context.Context, datasetID, edition, version string, r *http.Request, logData log.Data) (*models.ObservationsDoc, error) {

	var authorised bool
	if api.cfg.EnablePrivateEndpoints {
		authorised = api.checkIfAuthorised(r, logData)
	}

	userAuthToken := getUserAuthToken(r.Context())

	datasetDoc, err := api.getDataset(ctx, authorised, userAuthToken, datasetID, logData)
	if err != nil {
		log.Error(ctx, "failed to retrieve dataset doc", err)
		return nil, err
	}

	versionDoc, err := api.getVersion(ctx, authorised, userAuthToken, datasetID, edition, version, logData)
	if err != nil {
		return nil, err
	}

	if err = models.CheckState(models.Version, versionDoc.State); err != nil {
		logData["state"] = versionDoc.State
		log.Error(ctx, "get observations: version has an invalid state", err, logData)
		return nil, err
	}

	if versionDoc.Dimensions == nil {
		logData["version_doc"] = versionDoc
		log.Error(ctx, "get observations: missing dimensions in versio doc", errs.ErrMissingVersionDimensions, logData)
		return nil, errs.ErrMissingVersionDimensions
	}

	// loop through version dimensions to retrieve list of dimension names
	validDimensionNames := GetListOfValidDimensionNames(versionDoc.Dimensions)
	logData["version_dimensions"] = validDimensionNames

	// check query parameters match the version dimensions
	queryParameters, err := ExtractQueryParameters(r.URL.Query(), validDimensionNames)
	if err != nil {
		log.Error(ctx, "get observations: error extracting query parameters", err, logData)
		return nil, err
	}
	logData["query_parameters"] = queryParameters

	event := models.FilterSubmitted{
		DatasetID: datasetID,
		Edition:   edition,
		Version:   version,
	}

	// retrieve observations
	observations, err := api.getObservationList(ctx, &versionDoc, queryParameters, api.cfg.DefaultObservationLimit, logData, &event, userAuthToken)
	if err != nil {
		log.Error(ctx, "get observations: unable to retrieve observations", err, logData)
		return nil, err
	}

	return models.CreateObservationsDoc(api.cfg.ObservationAPIURL, r.URL.RawQuery, &versionDoc, datasetDoc, observations, queryParameters, defaultOffset, api.cfg.DefaultObservationLimit), nil
}

// GetDimensionOffsetInHeaderRow splits the first item of the provided headers by '_', and returns the second item as integer
func GetDimensionOffsetInHeaderRow(headerRow []string) (int, error) {
	metaData := strings.Split(headerRow[0], "_")

	if len(metaData) < 2 {
		return 0, errs.ErrIndexOutOfRange
	}

	dimensionOffset, err := strconv.Atoi(metaData[1])
	if err != nil {
		return 0, err
	}

	return dimensionOffset, nil
}

// getDataset obtains the Dataset document from Dataset API and validates that it is published if the caller is unauthorised to see unpublished datasets.
func (api *API) getDataset(ctx context.Context, authorised bool, userAuthToken, datasetID string, logData log.Data) (dataset.DatasetDetails, error) {
	// Get dataset from dataset API
	datasetDoc, err := api.datasetClient.Get(ctx, userAuthToken, api.cfg.ServiceAuthToken, "", datasetID)
	if err != nil {
		log.Error(ctx, "get observations: dataset api failed to retrieve dataset document", err, logData)

		datasetError, ok := err.(*dataset.ErrInvalidDatasetAPIResponse)
		if !ok {
			return dataset.DatasetDetails{}, err
		}

		switch datasetError.Code() {
		case http.StatusUnauthorized:
			return dataset.DatasetDetails{}, errs.ErrUnauthorised
		case http.StatusNotFound:
			return dataset.DatasetDetails{}, errs.ErrDatasetNotFound
		default:
			return dataset.DatasetDetails{}, errs.ErrInternalServer
		}
	}

	// If not authorised, only published datasets are accessible
	if !authorised {
		if datasetDoc.State != dataset.StatePublished.String() {
			logData["dataset_doc"] = datasetDoc
			log.Error(ctx, "get observations: dataset is not in published state", errs.ErrDatasetNotFound, logData)
			return dataset.DatasetDetails{}, errs.ErrDatasetNotFound
		}
	}

	return datasetDoc, nil
}

func (api *API) getVersion(ctx context.Context, authorised bool, userAuthToken, datasetID, edition, version string, logData log.Data) (dataset.Version, error) {
	// Get Version from dataset API
	versionDoc, err := api.datasetClient.GetVersion(ctx, userAuthToken, api.cfg.ServiceAuthToken, "", "", datasetID, edition, version)
	if err != nil {
		log.Error(ctx, "get observations: dataset api failed to retrieve dataset version", err, logData)

		datasetError, ok := err.(*dataset.ErrInvalidDatasetAPIResponse)
		if !ok {
			return dataset.Version{}, err
		}

		switch datasetError.Code() {
		case http.StatusUnauthorized:
			return dataset.Version{}, errs.ErrUnauthorised
		case http.StatusNotFound:
			return dataset.Version{}, errs.ErrVersionNotFound
		default:
			return dataset.Version{}, errs.ErrInternalServer
		}
	}

	// If not authorised, only published versions of datasets are accessible
	if !authorised {
		if versionDoc.State != dataset.StatePublished.String() {
			logData["version_doc"] = versionDoc
			log.Error(ctx, "get observations: dataset version is not in published state", errs.ErrDatasetNotFound, logData)
			return dataset.Version{}, errs.ErrVersionNotFound
		}
	}
	return versionDoc, err
}

// GetListOfValidDimensionNames iterates the provided dimensions and returns an array with their names
func GetListOfValidDimensionNames(dimensions []dataset.VersionDimension) []string {

	var dimensionNames []string
	for _, dimension := range dimensions {
		dimensionNames = append(dimensionNames, dimension.Name)
	}

	return dimensionNames
}

// ExtractQueryParameters creates a map of query parameters (options) by dimension from the provided urlQuery if they exist in the validDimensions list
func ExtractQueryParameters(urlQuery url.Values, validDimensions []string) (map[string]string, error) {
	queryParameters := make(map[string]string)
	var incorrectQueryParameters, missingQueryParameters, multivaluedQueryParameters []string

	// Map for efficiency
	validDimensionsMap := make(map[string]struct{})
	for _, validDimension := range validDimensions {
		validDimensionsMap[validDimension] = struct{}{}
	}

	// Determine if any request query parameters are invalid dimensions
	// and map the valid dimensions with their equivalent values in map
	for rawDimension, option := range urlQuery {
		// Ignore case sensitivity
		dimension := strings.ToLower(rawDimension)

		queryParamExists := false

		if _, dimFound := validDimensionsMap[dimension]; dimFound {
			queryParamExists = true
			queryParameters[dimension] = option[0]
			if len(option) != 1 {
				multivaluedQueryParameters = append(multivaluedQueryParameters, rawDimension)
			}
		}
		if !queryParamExists {
			incorrectQueryParameters = append(incorrectQueryParameters, rawDimension)
		}
	}

	if len(incorrectQueryParameters) > 0 {
		return nil, apierrors.ErrorIncorrectQueryParameters(incorrectQueryParameters)
	}

	if len(multivaluedQueryParameters) > 0 {
		return nil, apierrors.ErrorMultivaluedQueryParameters(multivaluedQueryParameters)
	}

	// Determine if any dimensions have not been set in request query parameters
	if len(queryParameters) != len(validDimensions) {
		for _, validDimension := range validDimensions {
			if queryParameters[validDimension] == "" {
				missingQueryParameters = append(missingQueryParameters, validDimension)
			}
		}
		return nil, apierrors.ErrorMissingQueryParameters(missingQueryParameters)
	}

	return queryParameters, nil
}

// getUserAuthToken obtains the user auth token from the context, expected under FlorenceIdentityKey
func getUserAuthToken(ctx context.Context) string {

	if request.IsFlorenceIdentityPresent(ctx) {
		return ctx.Value(request.FlorenceIdentityKey).(string)
	}

	return ""
}

// sortFilter by Dimension size, largest first, to make Neptune searches faster
// The sort is done here because the sizes are retrieved from Mongo and
// its best not to have the dp-graph library acquiring such coupling to its caller.
var SortFilter = func(ctx context.Context, api *API, event *models.FilterSubmitted, dbFilter *observation.DimensionFilters /*, userAuthToken string*/) {
	nofDimensions := len(dbFilter.Dimensions)
	if nofDimensions <= 1 {
		return
	}
	// Create a slice of sorted dimension sizes
	type dim struct {
		index         int
		dimensionSize int
	}

	dimSizes := make([]dim, 0, nofDimensions)
	var dimSizesMutex sync.Mutex

	// get info from mongo
	var getErrorCount int32
	var concurrent = 10 // limit number of go routines so as to not put too much on heap
	var semaphoreChan = make(chan struct{}, concurrent)
	var wg sync.WaitGroup // number of working goroutines

	for i, dimension := range dbFilter.Dimensions {
		if atomic.LoadInt32(&getErrorCount) != 0 {
			break
		}
		semaphoreChan <- struct{}{} // block while full

		wg.Add(1)

		// Get dimension sizes in parallel
		go func(i int, dimension *observation.Dimension) {
			defer func() {
				<-semaphoreChan // read to release a slot
			}()

			defer wg.Done()

			// passing a 'Limit' of 0 makes GetOptions skip getting the documents
			// and to return only what we are interested in: TotalCount
			options, err := api.datasetClient.GetOptions(ctx,
				"", // userAuthToken,
				api.cfg.ServiceAuthToken,
				"", // collectionID
				event.DatasetID, event.Edition, event.Version, dimension.Name,
				&dataset.QueryParams{Offset: 0, Limit: 0})

			if err != nil {
				if atomic.AddInt32(&getErrorCount, 1) <= 2 {
					// only show a few of possibly hundreds of errors, as once someone
					// looks into the one error they may fix all associated errors
					logData := log.Data{"dataset_id": event.DatasetID, "edition": event.Edition, "version": event.Version, "dimension name": dimension.Name}
					log.Info(ctx, "SortFilter: GetOptions failed for dataset and dimension", logData)
				}
			} else {
				d := dim{dimensionSize: options.TotalCount, index: i}
				dimSizesMutex.Lock()
				dimSizes = append(dimSizes, d)
				dimSizesMutex.Unlock()
			}
		}(i, dimension)
	}
	wg.Wait()

	if getErrorCount != 0 {
		logData := log.Data{"dataset_id": event.DatasetID, "edition": event.Edition, "version": event.Version}
		log.Info(ctx, fmt.Sprintf("SortFilter: GetOptions failed for dataset %d times, sorting by default of 'geography' first", getErrorCount), logData)
		// Frig dimension sizes and if geography is present, make it the largest (because it typically is the largest)
		// and to retain compatibility with what the neptune dp-graph library was doing without access to information
		// from mongo.
		dimSizes = dimSizes[:0]
		for i, dimension := range dbFilter.Dimensions {
			if strings.ToLower(dimension.Name) == "geography" {
				d := dim{dimensionSize: 999999, index: i}
				dimSizes = append(dimSizes, d)
			} else {
				// Set sizes of dimensions as largest first to retain list order to improve sort speed
				d := dim{dimensionSize: nofDimensions - i, index: i}
				dimSizes = append(dimSizes, d)
			}
		}
	}

	// sort slice by number of options per dimension, smallest first
	sort.Slice(dimSizes, func(i, j int) bool {
		return dimSizes[i].dimensionSize < dimSizes[j].dimensionSize
	})

	sortedDimensions := make([]observation.Dimension, 0, nofDimensions)

	for i := nofDimensions - 1; i >= 0; i-- { // build required return structure, largest first
		sortedDimensions = append(sortedDimensions, *dbFilter.Dimensions[dimSizes[i].index])
	}

	// Now copy the sorted dimensions back over the original
	for i, dimension := range sortedDimensions {
		*dbFilter.Dimensions[i] = dimension
	}
}

func (api *API) getObservationList(ctx context.Context, versionDoc *dataset.Version, queryParameters map[string]string, limit int, logData log.Data, event *models.FilterSubmitted, userAuthToken string) ([]models.Observation, error) {

	// Build query (observation.Filter type)
	var dimensionFilters []*observation.Dimension

	// Unable to have more than one wildcard parameter per query
	var wildcardParameter string

	// Build dimension filter object to create queryObject for neo4j
	for dimension, option := range queryParameters {
		if option == "*" {
			if wildcardParameter != "" {
				return nil, errs.ErrTooManyWildcards
			}

			wildcardParameter = dimension
			continue
		}

		dimensionFilter := &observation.Dimension{
			Name:    dimension,
			Options: []string{option},
		}

		dimensionFilters = append(dimensionFilters, dimensionFilter)
	}

	queryObject := observation.DimensionFilters{
		Dimensions: dimensionFilters,
	}

	SortFilter(ctx, api, event, &queryObject)

	logData["query_object"] = queryObject

	log.Info(ctx, "query object built to retrieve observations from db", logData)

	csvRowReader, err := api.graphDB.StreamCSVRows(ctx, versionDoc.ID, "", &queryObject, &limit)
	if err != nil {
		return nil, err
	}
	defer csvRowReader.Close(ctx)

	headerRow, err := csvRowReader.Read()
	if err != nil {
		return nil, err
	}

	headerRowReader := csv.NewReader(strings.NewReader(headerRow))
	headerRowArray, err := headerRowReader.Read()
	if err != nil {
		return nil, err
	}

	dimensionOffset, err := GetDimensionOffsetInHeaderRow(headerRowArray)
	if err != nil {
		log.Error(ctx, "get observations: unable to distinguish headers from version document", err, logData)
		return nil, err
	}

	var observationRow string
	var observations []models.Observation
	// Iterate over observation row reader
	for observationRow, err = csvRowReader.Read(); err != io.EOF; observationRow, err = csvRowReader.Read() {

		if err != nil {
			if strings.Contains(err.Error(), "the filter options created no results") {
				return nil, errs.ErrObservationsNotFound
			}
			return nil, err
		}

		observationRowReader := csv.NewReader(strings.NewReader(observationRow))
		observationRowArray, err := observationRowReader.Read()
		if err != nil {
			return nil, err
		}

		observations = append(observations, createObservation(
			versionDoc,
			observationRowArray,
			headerRowArray,
			dimensionOffset, wildcardParameter))
	}

	// neo4j will always return the same list of observations in the same
	// order as it is deterministic for static data, but this does not
	// necessarily mean we won't want to return observations in a particular
	// order (which may be costly on the services performance)

	return observations, nil
}

func createObservation(versionDoc *dataset.Version, observationRowArray, headerRowArray []string, dimensionOffset int, wildcardParameter string) models.Observation {
	observation := models.Observation{
		Observation: observationRowArray[0],
	}

	// add observation metadata
	if dimensionOffset != 0 {
		observationMetaData := make(map[string]string)

		for i := 1; i < dimensionOffset+1; i++ {
			observationMetaData[headerRowArray[i]] = observationRowArray[i]
		}

		observation.Metadata = observationMetaData
	}

	versionDocDimensions := make(map[string]dataset.VersionDimension)
	for _, dim := range versionDoc.Dimensions {
		versionDocDimensions[dim.Name] = dim
	}

	if wildcardParameter != "" {
		dimensions := make(map[string]*models.DimensionObject)

		for i := dimensionOffset + 2; i < len(observationRowArray); i += 2 {

			if strings.ToLower(headerRowArray[i]) == wildcardParameter {
				versionDimension, found := versionDocDimensions[wildcardParameter]
				if found {
					dimensions[headerRowArray[i]] = &models.DimensionObject{
						ID:    observationRowArray[i-1],
						HRef:  versionDimension.URL + "/codes/" + observationRowArray[i-1],
						Label: observationRowArray[i],
					}
				}
				break
			}
		}
		observation.Dimensions = dimensions
	}
	return observation
}

func handleObservationsErrorType(ctx context.Context, w http.ResponseWriter, err error, data log.Data) {

	_, isObservationErr := err.(apierrors.ObservationQueryError)
	var status int
	resErrMsg := err.Error()

	switch {
	case isObservationErr:
		status = http.StatusBadRequest
	case observationNotFound[err]:
		status = http.StatusNotFound
	case observationBadRequest[err]:
		status = http.StatusBadRequest
	default:
		resErrMsg = errs.ErrInternalServer.Error()
		status = http.StatusInternalServerError
	}

	if data == nil {
		data = log.Data{}
	}

	data["responseStatus"] = status
	log.Error(ctx, "get observation endpoint: request unsuccessful", err, data)
	http.Error(w, resErrMsg, status)
}
