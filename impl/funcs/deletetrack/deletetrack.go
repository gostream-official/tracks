package deletetrack

import (
	"fmt"
	"net/http"

	"github.com/gostream-official/tracks/impl/inject"
	"github.com/gostream-official/tracks/impl/models"
	"github.com/gostream-official/tracks/pkg/api"
	"github.com/gostream-official/tracks/pkg/marshal"
	"github.com/gostream-official/tracks/pkg/parallel"
	"github.com/gostream-official/tracks/pkg/store"
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

	idToDelete := request.PathParameters["id"]

	store := store.NewMongoStore[models.TrackInfo](injector.MongoInstance, "tracks", "tracks")
	count, err := store.DeleteItem(idToDelete)

	if err != nil {
		log.Errorf("[%s] failed to delete database items: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	if count == 0 {
		return &api.APIResponse{
			StatusCode: http.StatusNoContent,
		}
	}

	return &api.APIResponse{
		StatusCode: http.StatusAccepted,
	}
}
