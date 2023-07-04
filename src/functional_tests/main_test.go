package main

import (
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"os"
	"testing"
)

var testerApp *TesterApp

func TestMain(m *testing.M) {

	logger := logging.NewLogger(logging.LoggerOptions{
		IsProduction: false,
		AppName:      "Confluent Gateway Test App",
	})

	logger.Information("Setting up tester app")
	var err error
	testerApp, err = CreateAndSetupTester(logger)
	if err != nil {
		panic(fmt.Errorf("TesterApp setup error: %s", err))
	}

	logger.Information("Starting tests")
	testRunCode := m.Run()

	logger.Information(fmt.Sprintf("Finished tests with code %d", testRunCode))
	os.Exit(testRunCode)

}
