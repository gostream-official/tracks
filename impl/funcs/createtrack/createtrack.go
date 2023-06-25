package createtrack

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
//	The request body for the create track endpoint.
type CreateTrackRequestBody struct {
	// The artist.
	ArtistID string `json:"artistId"`

	// A list of featured artists.
	FeaturedArtistIDs []string `json:"featuredArtistIds"`

	// The track title.
	Title string `json:"title"`

	// The label that published the track.
	Label string `json:"label"`

	// The release date of the track.
	ReleaseDate string `json:"releaseDate"`

	// The statistics of the track.
	TrackStats CreateTrackStatsRequestBody `json:"trackStats"`

	// The audio features of the track.
	AudioFeatures CreateTrackAudioFeaturesRequestBody `json:"audioFeatures"`
}

// Description:
//
//	The request body for the track statistics.
type CreateTrackStatsRequestBody struct {

	// The stream count of the track.
	Streams uint32 `json:"streams"`

	// The amount of likes of the track.
	Likes uint32 `json:"likes"`
}

// Description:
//
//	The request body for the track's audio features.
type CreateTrackAudioFeaturesRequestBody struct {

	// The key of the track.
	Key string `json:"key"`

	// The tempo of the track.
	Tempo float32 `json:"tempo"`

	// The duration of the track.
	Duration float32 `json:"duration"`

	// The energy level of the track.
	Energy float32 `json:"energy"`

	// The danceability level of the track.
	Danceability float32 `json:"danceability"`

	// The accousticness level of the track.
	Accousticness float32 `json:"accousticness"`

	// The instrumentalness level of the track.
	Instrumentalness float32 `json:"instrumentalness"`

	// The liveness level of the track.
	Liveness float32 `json:"liveness"`

	// The track's loudness.
	Loudness float32 `json:"loudness"`

	// The track's time signature.
	TimeSignature int `json:"timeSignature"`
}

// Description:
//
//	The error response body for the create track endpoint.
type CreateTrackErrorResponseBody struct {

	// The error message.
	Message string `json:"message"`
}

// Description:
//
//	Describes a validation error.
type CreateTrackValidationError struct {

	// The JSON field which is referenced by the error message.
	FieldRef string `json:"ref"`

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
		return nil, fmt.Errorf("createtrack: failed to deduce injector")
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
func ExtractRequestBody(request *api.APIRequest) (*CreateTrackRequestBody, error) {
	body := &CreateTrackRequestBody{}

	bytes := []byte(request.Body)
	err := json.Unmarshal(bytes, body)

	if err != nil {
		return nil, err
	}

	return body, nil
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
func ValidateRequestBody(request *CreateTrackRequestBody) *CreateTrackValidationError {
	artistID := strings.TrimSpace(request.ArtistID)
	featuredArtistIDs := arrays.Map(request.FeaturedArtistIDs, func(artist string) string {
		return strings.TrimSpace(artist)
	})

	title := strings.TrimSpace(request.Title)
	releaseDate := strings.TrimSpace(request.ReleaseDate)

	_, err := uuid.Parse(artistID)
	if err != nil {
		return &CreateTrackValidationError{
			FieldRef:     "artistId",
			ErrorMessage: "value is not a valid uuid",
		}
	}

	for _, artist := range featuredArtistIDs {
		_, err := uuid.Parse(artist)
		if err != nil {
			return &CreateTrackValidationError{
				FieldRef:     "featuredArtistIds",
				ErrorMessage: "array contains invalid uuid",
			}
		}
	}

	if len(title) == 0 {
		return &CreateTrackValidationError{
			FieldRef:     "title",
			ErrorMessage: "value must not be empty",
		}
	}

	_, err = time.Parse("2006-01-02", releaseDate)
	if err != nil {
		return &CreateTrackValidationError{
			FieldRef:     "releaseDate",
			ErrorMessage: "expected following format: yyyy-MM-dd",
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
func ValidateStatsRequestBody(request *CreateTrackStatsRequestBody) *CreateTrackValidationError {
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
func ValidateAudioFeaturesRequestBody(request *CreateTrackAudioFeaturesRequestBody) *CreateTrackValidationError {
	return nil
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

	requestBody, err := ExtractRequestBody(request)
	if err != nil {
		log.Warnf("[%s] failed to extract request body: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusBadRequest,
			Body: CreateTrackErrorResponseBody{
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

	trackStore := store.NewMongoStore[models.TrackInfo](injector.MongoInstance, "gostream", "tracks")
	artistStore := store.NewMongoStore[models.ArtistInfo](injector.MongoInstance, "gostream", "artists")

	err = CheckIfArtistExists(artistStore, requestBody.ArtistID)
	if err != nil {
		log.Warnf("[%s] artist does not exist: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusBadRequest,
			Body: CreateTrackErrorResponseBody{
				Message: "artist does not exist",
			},
		}
	}

	for _, featuredArtist := range requestBody.FeaturedArtistIDs {
		err = CheckIfArtistExists(artistStore, featuredArtist)
		if err != nil {
			log.Warnf("[%s] featured artist does not exist: %s", context.ID, err)
			return &api.APIResponse{
				StatusCode: http.StatusBadRequest,
				Body: CreateTrackErrorResponseBody{
					Message: "featured artist does not exist",
				},
			}
		}
	}

	releaseDate, _ := time.Parse("2006-01-02", requestBody.ReleaseDate)
	track := models.TrackInfo{
		ID:                uuid.New().String(),
		ArtistID:          requestBody.ArtistID,
		FeaturedArtistIDs: requestBody.FeaturedArtistIDs,
		Title:             requestBody.Title,
		Label:             requestBody.Label,
		ReleaseDate:       releaseDate,
		TrackStats: models.TrackStats{
			Streams: requestBody.TrackStats.Streams,
			Likes:   requestBody.TrackStats.Likes,
		},
		AudioFeatures: models.AudioFeatures{
			Key:              requestBody.AudioFeatures.Key,
			Tempo:            requestBody.AudioFeatures.Tempo,
			Duration:         requestBody.AudioFeatures.Duration,
			Energy:           requestBody.AudioFeatures.Energy,
			Danceability:     requestBody.AudioFeatures.Danceability,
			Accousticness:    requestBody.AudioFeatures.Accousticness,
			Instrumentalness: requestBody.AudioFeatures.Instrumentalness,
			Liveness:         requestBody.AudioFeatures.Liveness,
			Loudness:         requestBody.AudioFeatures.Loudness,
			TimeSignature:    requestBody.AudioFeatures.TimeSignature,
		},
	}

	log.Tracef("[%s] attempting to create database item ...", context.ID)
	err = trackStore.CreateItem(track)

	if err != nil {
		log.Errorf("[%s] failed to create database item: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	log.Tracef("[%s] successfully completed request", context.ID)
	return &api.APIResponse{
		StatusCode: http.StatusOK,
		Body:       track,
	}
}
