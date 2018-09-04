package main

import (
	"fmt"
	"github.com/elastifile/emanage-go/src/size"
	"net/url"
	"path"

	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"

	"github.com/elastifile/emanage-go/src/emanage-client"
)

var EMSClient *emanage.Client

// Volume creation arguments
const (
	optionsSize            = "size"
	optionsUserMappingType = "user-mapping-type" // no_mapping, remap_root, remap_all
	optionsUserMappingUid  = "user-mapping-uid"
	optionsUserMappingGid  = "user-mapping-gid"
)

var defaultVolumeSize = 100 * size.GiB

func defaultDcCreateOpts(name string) *emanage.DcCreateOpts {
	return &emanage.DcCreateOpts{
		Name:           name,
		DirPermissions: 777,
		Dedup:          0,
		Compression:    1,
	}
}

func defaultExportCreateOpts() *emanage.ExportCreateOpts {
	return &emanage.ExportCreateOpts{
		Path:        "/",
		Access:      emanage.ExportAccessRW,
		UserMapping: emanage.UserMappingNone,
	}
}

func defaultDcExportCreateOpts(name string) (*emanage.DcCreateOpts, *emanage.ExportCreateOpts) {
	return defaultDcCreateOpts(name), defaultExportCreateOpts()
}

func defaultPolicy(ems *emanage.Client) (policy emanage.Policy, err error) {
	if ems == nil {
		err = errors.New("Got nil EMS client")
		return
	}
	var policies []emanage.Policy
	policies, err = ems.Policies.GetAll(nil)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to get policies from EMS", 0)
		return
	}

	found := false
	for i := range policies {
		if policies[i].IsDefault {
			policy = policies[i]
			found = true
			break
		}
	}

	if !found {
		err = errors.Errorf("Default policy not found")
		return
	}
	return
}

func CreateDc(opts *emanage.DcCreateOpts) (dcref *emanage.DataContainer, err error) {
	// TODO: Prune illegal characters
	name := fmt.Sprintf(opts.Name)
	// TODO: Support non-default policies via command line options and only use the default one if no policy was specified
	policy, err := defaultPolicy(EMSClient)
	if err != nil {
		err = errors.WrapPrefix(err, fmt.Sprintf("Failed to get policy for volume %s", opts.Name), 0)
		return
	}

	logrus.WithFields(logrus.Fields{"name": name, "policy id": policy.Id, "opts": opts}).Debug("Creating Data Container")
	dc, err := EMSClient.DataContainers.Create(name, policy.Id, opts)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to create Data Container", 0)
		return
	}

	dcref = &dc
	return
}

func CreateExport(name string, opts *emanage.ExportCreateOpts) (export emanage.Export, err error) {
	return EMSClient.Exports.Create(name, opts)
}

func dcExists(dcName string) (exists bool, dcRef *emanage.DataContainer, err error) {
	dcs, err := EMSClient.DataContainers.GetAll(nil)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to get Data Containers", 0)
		return
	}
	for _, dc := range dcs {
		if dc.Name == dcName {
			exists = true
			dcRef = &dc
			break
		}
	}
	return
}

func exportExists(exportName string, dcId int) (exists bool, exportRef *emanage.Export, err error) {
	exports, err := EMSClient.Exports.GetAll(nil)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to get Exports", 0)
		return
	}
	for _, export := range exports {
		if export.Name == exportName && export.DataContainerId == dcId {
			exists = true
			exportRef = &export
			break
		}
	}
	return
}

// maybeCreateDc creates DC if it doesn't exist
func maybeCreateDc(dcOpts *emanage.DcCreateOpts) (*emanage.DataContainer, error) {
	exists, dc, err := dcExists(dcOpts.Name)
	if err != nil {
		return nil, errors.WrapPrefix(err, "Failed to check if Data Container exists", 0)
	}
	if !exists {
		dc, err = CreateDc(dcOpts)
		if err != nil {
			return nil, errors.WrapPrefix(err, "Failed to create Data Container", 0)
		}
	}
	return dc, nil
}

func maybeCreateExport(exportName string, exportOpts *emanage.ExportCreateOpts) (*emanage.Export, error) {
	exists, export, err := exportExists(exportName, exportOpts.DcId)
	if err != nil {
		return nil, errors.WrapPrefix(err, "Failed to check if Export exists", 0)
	}
	if !exists {
		exp, err := CreateExport(exportName, exportOpts)
		if err != nil {
			return nil, errors.WrapPrefix(err, "Failed to create Export", 0)
		}
		export = &exp
	} else {

	}
	return export, nil
}

func maybeCreateDcExport(dcOpts *emanage.DcCreateOpts, exportOpts *emanage.ExportCreateOpts) (
	export *emanage.Export, dc *emanage.DataContainer, err error) {

	// Create Data Container if it doesn't exist
	dc, err = maybeCreateDc(dcOpts)
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}

	// Create Export if it doesn't exist
	exportName := "e"
	exportOpts.DcId = dc.Id
	export, err = maybeCreateExport(exportName, exportOpts)
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}

	logrus.WithFields(logrus.Fields{"name": dc.Name, "id": dc.Id, "export": export.Name}).Info("Created DC, Export")
	return export, dc, nil
}

func DeleteDc(dc *emanage.DataContainer) (err error) {
	// TODO: Return success if DC doesn't exist
	_, err = EMSClient.DataContainers.Delete(dc)
	return err
}

func DeleteExport(export *emanage.Export) (err error) {
	// TODO: Return success if export doesn't exist
	_, err = EMSClient.Exports.Delete(export)
	return
}

//func DeleteDcExport(dc *emanage.DataContainer, export *emanage.Export) (err error) {
//	err = DeleteExport(export)
//	if err != nil {
//		// TODO: Wrap error
//		return
//	}
//	err = DeleteDc(dc)
//	if err != nil {
//		// TODO: Wrap error
//		return
//	}
//	return
//}

func DeleteDcExport(v *elastifileVolume) (err error) {
	err = DeleteExport(v.Export)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to delete Export", 0)
		return
	}
	err = DeleteDc(v.DataContainer)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to delete Data Container", 0)
		return
	}
	return
}

func dcExportPath(export *emanage.Export) (dir string, err error) {
	dc, err := EMSClient.DataContainers.GetFull(export.DataContainerId)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to get Data Containers", 0)
		return
	}
	dir = path.Join(dc.Name, export.Name)

	return
}

func getEMSClient(driver *elastifileDriver) (ems *emanage.Client, err error) {
	emsUrl := &url.URL{
		Scheme: "http",
		Host:   driver.managementAddr,
	}
	ems = emanage.NewClient(emsUrl)
	if ems == nil {
		err = errors.New("Failed to create new EMS client")
		return
	}

	err = ems.Sessions.Login(driver.managementUser, driver.managementPassword)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to log into EMS", 0)
		return
	} else {
		logrus.Info("Logged into EMS")
	}
	return
}
