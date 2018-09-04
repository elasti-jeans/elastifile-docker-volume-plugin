package main

import (
	"fmt"
	"github.com/go-errors/errors"
	"os"
	"strconv"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/sirupsen/logrus"
	//"github.com/elastifile/emanage-go"
	//"github.com/elastifile/emanage-go/src/emanage-client"
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

func main() {
	debug := os.Getenv("DEBUG")

	if ok, _ := strconv.ParseBool(debug); ok {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// TODO: logrus.SetFormatter()

	logrus.Infof("Initializing %v", pluginName)
	driver, ems, err := newElastifileDriver(driverInfo)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to initialize plugin", 0)
		logFatalAndPanic(err.Error())
	}
	if ems == nil {
		panic("Got nil EMS client")
	}
	EMSClient = ems

	handler := volume.NewHandler(driver)
	if handler == nil {
		err = errors.WrapPrefix(err, "Received nil handler", 0)
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
