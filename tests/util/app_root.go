package testutil

import (
	"log"
	"os"
)

func GetAppRoot() string {
	appRoot, found := os.LookupEnv("TEST_APP_ROOT")

	if !found {
		log.Fatalf("TEST_APP_ROOT must be set!")
	}

	return appRoot
}
