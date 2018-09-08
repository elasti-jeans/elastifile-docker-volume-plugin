package main

import (
	"github.com/go-errors/errors"

	"github.com/elastifile/emanage-go/src/emanage-client"
)

type elastifileVolume struct {
	connections   int
	Mountpoint    string
	MountOpts     []string
	Export        *emanage.Export
	DataContainer *emanage.DataContainer
}

func (v *elastifileVolume) ExportPath() (exportPath string, err error) {
	exportPath, err = Ems.dcExportPath(v.Export)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to get export path from EMS", 0)
	}
	return
}
