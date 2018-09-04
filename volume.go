package main

import (
	"github.com/sirupsen/logrus"

	"github.com/elastifile/emanage-go/src/emanage-client"
)

type elastifileVolume struct {
	connections   int
	Mountpoint    string
	MountOpts     []string
	Export        *emanage.Export
	DataContainer *emanage.DataContainer
}

func (v *elastifileVolume) ExportPath() (exportPath string) {
	exportPath, err := Ems.dcExportPath(v.Export)
	if err != nil {
		// TODO: Check spec if it's ok to return an empty path in this case. Alternatives: "/", panic
		logrus.Error("Failed to get export path", "volume", v, "err", err.Error())
		return ""
	}

	return exportPath
}
