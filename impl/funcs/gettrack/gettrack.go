package gettrack

import (
	"fmt"
	"net/http"

	"github.com/gostream-official/tracks/impl/inject"
	"github.com/gostream-official/tracks/impl/models"
	"github.com/gostream-official/tracks/pkg/api"
	"github.com/gostream-official/tracks/pkg/marshal"
	"github.com/gostream-official/tracks/pkg/parallel"
	"github.com/gostream-official/tracks/pkg/store"
	"github.com/gostream-official/tracks/pkg/store/query"
	"github.com/revx-official/output/log"
)

// Description:
//
//	Attempts to cast the input object to the endpoint injector.
//	If this cast fails, we cannot proceed to process this request.
//
// Parameters:
//
//	context The current request context.
//	object 	The injector object.
//
// Returns:
//
//	The injector if the cast is successful, an error otherwise.
func GetSafeInjector(context *parallel.Context, object interface{}) (*inject.Injector, error) {
	injector, ok := object.(inject.Injector)

	if !ok {
		return nil, fmt.Errorf("gettrackbyid: failed to deduce injector")
	}

	return &injector, nil
}

// Description:
//
//	The router handler for: Get Track By ID
//
// Parameters:
//
//	request 	The incoming request.
//	injector 	The injector. Contains injected dependencies.
//
// Returns:
//
//	An API response object.
func Handler(request *api.APIRequest, object interface{}) *api.APIResponse {
	context := parallel.NewContext()

	log.Infof("[%s] %s: %s", context.ID, request.Method, request.Path)
	log.Tracef("[%s] request: %s", context.ID, marshal.Quick(request))

	injector, err := GetSafeInjector(context, object)
	if err != nil {
		log.Errorf("[%s] failed to get endpoint injector: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	store := store.NewMongoStore[models.TrackInfo](injector.MongoInstance, "tracks", "tracks")

	filter := query.Filter{
		Root: query.FilterOperatorEq{
			Key:   "_id",
			Value: request.PathParameters["id"],
		},
		Limit: 10,
	}

	items, err := store.FindItems(&filter)

	if err != nil {
		log.Errorf("[%s] failed to retrieve database items: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	if len(items) == 0 {
		return &api.APIResponse{
			StatusCode: http.StatusNotFound,
		}
	}

	resultItem := items[0]
	return &api.APIResponse{
		StatusCode: http.StatusOK,
		Body:       resultItem,
	}
}
