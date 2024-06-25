package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/alerty-ai/alerty-go/pkg/alerty"
)

func main() {
	alerty.Start(alerty.AlertyServiceConfig{
		Name:           "crash-example",
		Version:        "1.0.0",
		Environment:    "local",
		OrganizationID: os.Getenv("ALERTY_ORG_ID"),
		IngestURL:      os.Getenv("ALERTY_INGEST_URL"),
		Debug:          true,
	})
	defer alerty.Stop()

	log.Info().Msg("Hello, Alerty!")

	err := errors.New("test error")
	alerty.CaptureError(err)

	log.Info().Msg("Goodbye, Alerty!")
}
