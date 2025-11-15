//go:build integration

package integration_test_deactivate_team

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	err := HealthCheck(defaultAttempts)
	if err != nil {
		panic(err)
	}
	log.Infof("integration tests: host %s is available", basePath)
	os.Exit(m.Run())
}
