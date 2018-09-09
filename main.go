package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
)

const (
	socketAddress = "/run/docker/plugins/elastifile.sock"
	pluginName    = "Elastifile Docker Volume Plugin"
)

func logErrorAndReturn(format string, args ...interface{}) error {
	logrus.Errorf(format, args...)
	return fmt.Errorf(format, args...)
}

func logFatalAndPanic(format string, args ...interface{}) error {
	logrus.Fatalf(format, args...)
	panic(fmt.Sprintf(format, args...))
}

// initFromEnv initializes the plugin configuration from environment variables defined in config.json
// The variables start with default specified in config.json and can be overridden via docker plugin install/set
func initFromEnv() {
	driverInfo.RestAddr = os.Getenv("MGMT_ADDRESS")
	driverInfo.RestUser = os.Getenv("MGMT_USERNAME")
	driverInfo.RestPass = os.Getenv("MGMT_PASSWORD")
	driverInfo.StorageAddr = os.Getenv("NFS_ADDRESS")

	envVarName := "CRUD_IDEMPOTENT"
	envVarValue := os.Getenv(envVarName)
	volCrudIdempotent, err := strconv.ParseBool(envVarValue)
	if err != nil {
		err = errors.WrapPrefix(err, fmt.Sprintf("Failed to parse environment variable's value. %v='%v'",
			envVarName, envVarValue), 0)
		logFatalAndPanic(err.Error())
	}
	driverInfo.CrudIdempotent = volCrudIdempotent

	envVarName = "DEBUG"
	envVarValue = os.Getenv(envVarName)
	enableDebug, err := strconv.ParseBool(envVarValue)
	if err != nil {
		err = errors.WrapPrefix(err, fmt.Sprintf("Failed to parse environment variable's value. %v='%v'",
			envVarName, envVarValue), 0)
		logFatalAndPanic(err.Error())
	}
	if enableDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	initFromEnv()
	// TODO: logrus.SetFormatter()

	logrus.Infof("Initializing %v", pluginName)
	driver, err := newElastifileDriver(driverInfo)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to initialize plugin", 0)
		logFatalAndPanic(err.Error())
	}

	handler := volume.NewHandler(driver)
	if handler == nil {
		err = errors.WrapPrefix(err, "Received nil volume handler", 0)
		logFatalAndPanic(err.Error())
	}

	logrus.Debugf("Getting ready to listen on %s", socketAddress)
	err = handler.ServeUnix(socketAddress, 0)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to start listener", 0)
		logFatalAndPanic(err.Error())
	}
	logrus.Infof("%v initialized - listening on %v", pluginName, socketAddress)
}
