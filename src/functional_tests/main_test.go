package main

import (
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"os"
	"testing"
)

const (
	nukePrevDataOnErr = false
	prevRunFileName   = "test.run"
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
	defer func() {
		//TODO: make more graceful teardown removing only test data from functional tests
		testerApp.FullTearDown()
	}()

	_, err = os.Stat(prevRunFileName)
	if os.IsNotExist(err) {
		err = os.WriteFile(prevRunFileName, []byte("(=^-ω-^=)"), 0644)
		if err != nil {
			panic(err)
		}
	} else {
		logger.Warning(fmt.Sprintf("found file %s, which indicates that previous test run did not end successfully.", prevRunFileName))
		if !nukePrevDataOnErr {
			logger.Warning("exiting without running tests, set nukePrevDataOnErr=true to nuke previous test data")
			return
		}
		logger.Information("nuking test data in DB")
		testerApp.FullTearDown()
	}

	logger.Information("Starting tests")
	testRunCode := m.Run()
	logger.Information(fmt.Sprintf("Finished tests with code %d", testRunCode))

	logger.Information(fmt.Sprintf("Removing %s file", prevRunFileName))
	err = os.Remove(prevRunFileName)
	if err != nil {
		panic(err)
	}
}
