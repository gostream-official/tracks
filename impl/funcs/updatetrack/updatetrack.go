package updatetrack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gostream-official/tracks/impl/inject"
	"github.com/gostream-official/tracks/impl/models"
	"github.com/gostream-official/tracks/pkg/api"
	"github.com/gostream-official/tracks/pkg/arrays"
	"github.com/gostream-official/tracks/pkg/marshal"
	"github.com/gostream-official/tracks/pkg/parallel"
	"github.com/gostream-official/tracks/pkg/store"
	"github.com/gostream-official/tracks/pkg/store/query"
	"github.com/revx-official/output/log"

	"github.com/google/uuid"
)

// Description:
//
//	The request body for the update track endpoint.
type UpdateTrackRequestBody struct {

	// The artist.
	ArtistID string `json:"artistId,omitempty"`

	// A list of featured artists.
	FeaturedArtistIDs []string `json:"featuredArtistIds,omitempty"`

	// The track title.
	Title string `json:"title,omitempty"`

	// The label that published the track.
	Label string `json:"label,omitempty"`

	// The release date of the track.
	ReleaseDate string `json:"releaseDate,omitempty"`

	// The statistics of the track.
	TrackStats UpdateTrackStatsRequestBody `json:"trackStats,omitempty"`

	// The audio features of the track.
	AudioFeatures UpdateTrackAudioFeaturesRequestBody `json:"audioFeatures,omitempty"`
}

// Description:
//
//	The request body for the track statistics.
type UpdateTrackStatsRequestBody struct {

	// The stream count of the track.
	Streams uint32 `json:"streams,omitempty"`

	// The amount of likes of the track.
	Likes uint32 `json:"likes,omitempty"`
}

// Description:
//
//	The request body for the track's audio features.
type UpdateTrackAudioFeaturesRequestBody struct {

	// The key of the track.
	Key string `json:"key,omitempty"`

	// The tempo of the track.
	Tempo float32 `json:"tempo,omitempty"`

	// The duration of the track.
	Duration float32 `json:"duration,omitempty"`

	// The energy level of the track.
	Energy float32 `json:"energy,omitempty"`

	// The danceability level of the track.
	Danceability float32 `json:"danceability,omitempty"`

	// The accousticness level of the track.
	Accousticness float32 `json:"accousticness,omitempty"`

	// The instrumentalness level of the track.
	Instrumentalness float32 `json:"instrumentalness,omitempty"`

	// The liveness level of the track.
	Liveness float32 `json:"liveness,omitempty"`

	// The track's loudness.
	Loudness float32 `json:"loudness,omitempty"`

	// The track's time signature.
	TimeSignature int `json:"timeSignature,omitempty"`
}

// Description:
//
//	The error response body for the create track endpoint.
type UpdateTrackErrorResponseBody struct {

	// The error message.
	Message string `json:"message"`
}

// Description:
//
//	Describes a validation error.
type UpdateTrackValidationError struct {

	// The JSON field which is referenced by the error message.
	FieldRef string `json:"ref"`

	// The error message.
	ErrorMessage string `json:"error"`
}

// Description:
//
//	Describes a path parameter validation error.
type UpdateTrackPathValidationError struct {

	// The JSON field which is referenced by the error message.
	PathRef string `json:"pathRef"`

	// The error message.
	ErrorMessage string `json:"error"`
}

// Description:
//
//	Attempts to cast the input object to the endpoint injector.
//	If this cast fails, we cannot proceed to process this request.
//
// Parameters:
//
//	object 	The injector object.
//
// Returns:
//
//	The injector if the cast is successful, an error otherwise.
func GetSafeInjector(object interface{}) (*inject.Injector, error) {
	injector, ok := object.(inject.Injector)

	if !ok {
		return nil, fmt.Errorf("updatetrack: failed to deduce injector")
	}

	return &injector, nil
}

// Description:
//
//	Unmarshals the request body for this endpoint.
//
// Parameters:
//
//	request The original request.
//
// Returns:
//
//	The unmarshalled request body, or an error when unmarshalling fails.
func ExtractRequestBody(request *api.APIRequest) (*UpdateTrackRequestBody, error) {
	body := &UpdateTrackRequestBody{}

	bytes := []byte(request.Body)
	err := json.Unmarshal(bytes, body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

// Description:
//
//	Gets and validates the id path parameter.
//
// Parameters:
//
//	request The http request.
//
// Returns:
//
//	The id path parameter.
//	A validatior error if the id is not a valid uuid.
func GetAndValidateID(request *api.APIRequest) (string, *UpdateTrackPathValidationError) {
	id := request.PathParameters["id"]

	_, err := uuid.Parse(id)
	if err != nil {
		return "", &UpdateTrackPathValidationError{
			PathRef:      ":id",
			ErrorMessage: "value is not a valid uuid",
		}
	}

	return id, nil
}

// Description:
//
//	Validates the request body for this endpoint.
//
// Parameters:
//
//	request The request body.
//
// Returns:
//
//	An error if the validation fails.
func ValidateRequestBody(request *UpdateTrackRequestBody) *UpdateTrackValidationError {

	if request.ArtistID != "" {
		artistID := strings.TrimSpace(request.ArtistID)

		_, err := uuid.Parse(artistID)
		if err != nil {
			return &UpdateTrackValidationError{
				FieldRef:     "artistId",
				ErrorMessage: "value is not a valid uuid",
			}
		}

	}

	if len(request.FeaturedArtistIDs) > 0 {
		featuredArtistIDs := arrays.Map(request.FeaturedArtistIDs, func(artist string) string {
			return strings.TrimSpace(artist)
		})

		for _, artist := range featuredArtistIDs {
			_, err := uuid.Parse(artist)
			if err != nil {
				return &UpdateTrackValidationError{
					FieldRef:     "featuredArtistIds",
					ErrorMessage: "array contains invalid uuid",
				}
			}
		}
	}

	if request.Title != "" {
		title := strings.TrimSpace(request.Title)

		if len(title) == 0 {
			return &UpdateTrackValidationError{
				FieldRef:     "title",
				ErrorMessage: "value must not be empty",
			}
		}
	}

	if request.ReleaseDate != "" {
		releaseDate := strings.TrimSpace(request.ReleaseDate)

		_, err := time.Parse("2006-01-02", releaseDate)
		if err != nil {
			return &UpdateTrackValidationError{
				FieldRef:     "releaseDate",
				ErrorMessage: "expected following format: yyyy-MM-dd",
			}
		}
	}

	result := ValidateStatsRequestBody(&request.TrackStats)
	if result != nil {
		return result
	}

	return ValidateAudioFeaturesRequestBody(&request.AudioFeatures)
}

// Description:
//
//	Validates the statistics request body for this endpoint.
//
// Parameters:
//
//	request The statistics request body.
//
// Returns:
//
//	An error if the validation fails.
func ValidateStatsRequestBody(request *UpdateTrackStatsRequestBody) *UpdateTrackValidationError {
	return nil
}

// Description:
//
//	Validates the audio features request body for this endpoint.
//
// Parameters:
//
//	request The audio features request body.
//
// Returns:
//
//	An error if the validation fails.
func ValidateAudioFeaturesRequestBody(request *UpdateTrackAudioFeaturesRequestBody) *UpdateTrackValidationError {
	return nil
}

// Description:
//
//	Searches a track with the given id in the database.
//
// Parameters:
//
//	store 	The store to search through.
//	id 		The id to search for.
//
// Returns:
//
//	The first matched track.
//	An error if the query fails.
func FindTrackByID(store *store.MongoStore[models.TrackInfo], id string) (*models.TrackInfo, error) {
	filter := query.Filter{
		Root: query.FilterOperatorEq{
			Key:   "_id",
			Value: id,
		},
		Limit: 1,
	}

	items, err := store.FindItems(&filter)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("update: track not found")
	}

	return &items[0], nil
}

// Description:
//
//	Checks whether the given artist id exists in the mongo store.
//
// Parameters:
//
//	store 		The mongo store to search.
//	artistID 	The artist id to search.
//
// Returns:
//
//	An error, if the artist could not be found or an error,
//	if the database request failed, nothing if successful.
func CheckIfArtistExists(store *store.MongoStore[models.ArtistInfo], artistID string) error {
	filter := query.Filter{
		Root: query.FilterOperatorEq{
			Key:   "_id",
			Value: artistID,
		},
		Limit: 1,
	}

	items, err := store.FindItems(&filter)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("artist not found")
	}

	return nil
}

// Description:
//
//	The router handler for track creation.
//
// Parameters:
//
//	request The incoming request.
//	object 	The injector. Contains injected dependencies.
//
// Returns:
//
//	An API response object.
func Handler(request *api.APIRequest, object interface{}) *api.APIResponse {
	context := parallel.NewContext()

	log.Infof("[%s] %s: %s", context.ID, request.Method, request.Path)
	log.Tracef("[%s] request: %s", context.ID, marshal.Quick(request))

	injector, err := GetSafeInjector(object)
	if err != nil {
		log.Warnf("[%s] failed to get endpoint injector: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	id, validationErr := GetAndValidateID(request)
	if validationErr != nil {
		log.Warnf("[%s] failed path parameter validation: %s", context.ID, validationErr.ErrorMessage)
		return &api.APIResponse{
			StatusCode: http.StatusBadRequest,
			Body:       validationErr,
		}
	}

	trackStore := store.NewMongoStore[models.TrackInfo](injector.MongoInstance, "gostream", "tracks")
	artistStore := store.NewMongoStore[models.ArtistInfo](injector.MongoInstance, "gostream", "artists")

	trackInfo, err := FindTrackByID(trackStore, id)
	if err != nil {
		log.Warnf("[%s] could not find track: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusNotFound,
		}
	}

	requestBody, err := ExtractRequestBody(request)
	if err != nil {
		log.Warnf("[%s] failed to extract request body: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusBadRequest,
			Body: UpdateTrackErrorResponseBody{
				Message: "invalid request body",
			},
		}
	}

	validationError := ValidateRequestBody(requestBody)
	if validationError != nil {
		log.Warnf("[%s] failed request body validation: %s", context.ID, validationError.ErrorMessage)
		return &api.APIResponse{
			StatusCode: http.StatusBadRequest,
			Body:       validationError,
		}
	}

	if requestBody.ArtistID != "" {
		err = CheckIfArtistExists(artistStore, requestBody.ArtistID)
		if err != nil {
			log.Warnf("[%s] artist does not exist: %s", context.ID, err)
			return &api.APIResponse{
				StatusCode: http.StatusBadRequest,
				Body: UpdateTrackErrorResponseBody{
					Message: "artist does not exist",
				},
			}
		}
	}

	if len(requestBody.FeaturedArtistIDs) > 0 {
		for _, featuredArtist := range requestBody.FeaturedArtistIDs {
			err = CheckIfArtistExists(artistStore, featuredArtist)
			if err != nil {
				log.Warnf("[%s] featured artist does not exist: %s", context.ID, err)
				return &api.APIResponse{
					StatusCode: http.StatusBadRequest,
					Body: UpdateTrackErrorResponseBody{
						Message: "featured artist does not exist",
					},
				}
			}
		}
	}

	if requestBody.ArtistID != "" {
		trackInfo.ArtistID = requestBody.ArtistID
	}

	if len(requestBody.FeaturedArtistIDs) > 0 {
		trackInfo.FeaturedArtistIDs = requestBody.FeaturedArtistIDs
	}

	if requestBody.Title != "" {
		trackInfo.Title = requestBody.Title
	}

	if requestBody.Label != "" {
		trackInfo.Label = requestBody.Label
	}

	if requestBody.ReleaseDate != "" {
		releaseDate, _ := time.Parse("2006-01-02", requestBody.ReleaseDate)
		trackInfo.ReleaseDate = releaseDate
	}

	if requestBody.TrackStats.Streams != 0 {
		trackInfo.TrackStats.Streams = requestBody.TrackStats.Streams
	}

	if requestBody.TrackStats.Likes != 0 {
		trackInfo.TrackStats.Likes = requestBody.TrackStats.Likes
	}

	if requestBody.AudioFeatures.Key != "" {
		trackInfo.AudioFeatures.Key = requestBody.AudioFeatures.Key
	}

	if requestBody.AudioFeatures.Tempo != 0 {
		trackInfo.AudioFeatures.Tempo = requestBody.AudioFeatures.Tempo
	}

	if requestBody.AudioFeatures.Duration != 0 {
		trackInfo.AudioFeatures.Duration = requestBody.AudioFeatures.Duration
	}

	if requestBody.AudioFeatures.Energy != 0 {
		trackInfo.AudioFeatures.Energy = requestBody.AudioFeatures.Energy
	}

	if requestBody.AudioFeatures.Danceability != 0 {
		trackInfo.AudioFeatures.Danceability = requestBody.AudioFeatures.Danceability
	}

	if requestBody.AudioFeatures.Accousticness != 0 {
		trackInfo.AudioFeatures.Accousticness = requestBody.AudioFeatures.Accousticness
	}

	if requestBody.AudioFeatures.Instrumentalness != 0 {
		trackInfo.AudioFeatures.Instrumentalness = requestBody.AudioFeatures.Instrumentalness
	}

	if requestBody.AudioFeatures.Liveness != 0 {
		trackInfo.AudioFeatures.Liveness = requestBody.AudioFeatures.Liveness
	}

	if requestBody.AudioFeatures.Loudness != 0 {
		trackInfo.AudioFeatures.Instrumentalness = requestBody.AudioFeatures.Loudness
	}

	if requestBody.AudioFeatures.TimeSignature != 0 {
		trackInfo.AudioFeatures.TimeSignature = requestBody.AudioFeatures.TimeSignature
	}

	updateFilter := query.Filter{
		Root: query.FilterOperatorEq{
			Key:   "_id",
			Value: id,
		},
	}

	updateOperator := query.Update{
		Root: query.UpdateOperatorSet{
			Set: map[string]interface{}{
				"artistId":                       trackInfo.ArtistID,
				"featuredArtistIds":              trackInfo.FeaturedArtistIDs,
				"title":                          trackInfo.Title,
				"label":                          trackInfo.Label,
				"releaseDate":                    trackInfo.ReleaseDate,
				"trackStats.streams":             trackInfo.TrackStats.Streams,
				"trackStats.likes":               trackInfo.TrackStats.Likes,
				"audioFeatures.key":              trackInfo.AudioFeatures.Key,
				"audioFeatures.tempo":            trackInfo.AudioFeatures.Tempo,
				"audioFeatures.duration":         trackInfo.AudioFeatures.Duration,
				"audioFeatures.energy":           trackInfo.AudioFeatures.Energy,
				"audioFeatures.danceability":     trackInfo.AudioFeatures.Danceability,
				"audioFeatures.accousticness":    trackInfo.AudioFeatures.Accousticness,
				"audioFeatures.instrumentalness": trackInfo.AudioFeatures.Instrumentalness,
				"audioFeatures.liveness":         trackInfo.AudioFeatures.Liveness,
				"audioFeatures.loudness":         trackInfo.AudioFeatures.Loudness,
				"audioFeatures.timeSignature":    trackInfo.AudioFeatures.TimeSignature,
			},
		},
	}

	log.Tracef("[%s] attempting to update database item ...", context.ID)
	count, err := trackStore.UpdateItem(&updateFilter, &updateOperator)

	if err != nil {
		log.Errorf("[%s] failed to update database item: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	if count == 0 {
		log.Warnf("[%s] zero modified items", context.ID)
		return &api.APIResponse{
			StatusCode: http.StatusNoContent,
		}
	}

	log.Tracef("[%s] successfully completed request", context.ID)
	return &api.APIResponse{
		StatusCode: http.StatusNoContent,
	}
}
