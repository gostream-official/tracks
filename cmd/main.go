package main

import (
	"fmt"
	"strconv"

	"github.com/gostream-official/tracks/impl/funcs/createtrack"
	"github.com/gostream-official/tracks/impl/funcs/deletetrack"
	"github.com/gostream-official/tracks/impl/funcs/gettrack"
	"github.com/gostream-official/tracks/impl/funcs/gettracks"
	"github.com/gostream-official/tracks/impl/funcs/updatetrack"
	"github.com/gostream-official/tracks/impl/inject"
	"github.com/gostream-official/tracks/pkg/env"
	"github.com/gostream-official/tracks/pkg/router"
	"github.com/gostream-official/tracks/pkg/store"

	"github.com/revx-official/output/log"
)

// Description:
//
//	The package initializer function.
//	Initializes the log level to info.
func init() {
	log.Level = log.LevelInfo
}

// Description:
//
//	The main function.
//	Represents the entry point of the application.
func main() {
	log.Infof("booting service instance ...")

	executionPortEnvVar := env.GetEnvironmentVariableWithFallback("PORT", "9871")
	executionPort, err := strconv.Atoi(executionPortEnvVar)

	if err != nil {
		log.Fatalf("Received invalid execution port")
	}

	if executionPort < 0 || executionPort > 65535 {
		log.Fatalf("Received invalid execution port")
	}

	mongoUsername, err := env.GetEnvironmentVariable("MONGO_USERNAME")
	if err != nil {
		log.Fatalf("Cannot retrieve mongo username via environment variable")
	}

	mongoPassword, err := env.GetEnvironmentVariable("MONGO_PASSWORD")
	if err != nil {
		log.Fatalf("Cannot retrieve mongo password via environment variable")
	}

	mongoHost := env.GetEnvironmentVariableWithFallback("MONGO_HOST", "127.0.0.1:27017")

	connectionURI := fmt.Sprintf("mongodb://%s:%s@%s", mongoUsername, mongoPassword, mongoHost)
	instance, err := store.NewMongoInstance(connectionURI)

	log.Infof("establishing database connection ...")
	if err != nil {
		log.Fatalf("failed to connect to mongo instance: %s", err)
	}

	log.Infof("successfully established database connection")

	injector := inject.Injector{
		MongoInstance: instance,
	}

	log.Infof("launching router engine ...")
	engine := router.Default()

	engine.HandleWith("GET", "/tracks", gettracks.Handler).Inject(injector)
	engine.HandleWith("GET", "/tracks/:id", gettrack.Handler).Inject(injector)
	engine.HandleWith("POST", "/tracks", createtrack.Handler).Inject(injector)
	engine.HandleWith("PUT", "/tracks/:id", updatetrack.Handler).Inject(injector)
	engine.HandleWith("DELETE", "/tracks/:id", deletetrack.Handler).Inject(injector)

	err = engine.Run(uint16(executionPort))
	if err != nil {
		log.Fatalf("failed to launch router engine: %s", err)
	}
}
