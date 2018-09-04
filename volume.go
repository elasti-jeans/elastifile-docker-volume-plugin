package main

import (
	"github.com/elastifile/emanage-go/src/emanage-client"
	"github.com/sirupsen/logrus"
)

type elastifileVolume struct {
	connections   int
	Mountpoint    string
	MountOpts     []string
	Export        *emanage.Export
	DataContainer *emanage.DataContainer
}

func (v *elastifileVolume) ExportPath() (exportPath string) {
	//exportPath, err := dcExportPath(&emanage.Export{})
	exportPath, err := dcExportPath(v.Export)
	if err != nil {
		// TODO: Log error
		// TODO: Check spec if it's ok to return an empty path in this case. Alternatives: "/", panic
		logrus.Error("Failed to get export path", "volume", v, "err", err.Error())
		return ""
	}

	//return exportPath
	return "/export" // TODO: FIXME: Return actual export path
}
