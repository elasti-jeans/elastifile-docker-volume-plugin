package main

import (
	"fmt"
	"github.com/elastifile/emanage-go/src/optional"
	"net/url"
	"path"

	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"

	"github.com/elastifile/emanage-go/src/emanage-client"
	"github.com/elastifile/emanage-go/src/size"
)

type EmsWrapper struct {
	client             *emanage.Client // Do not access this field directly, use Client() instead
	sessionInitialized bool
}

// Volume creation arguments
const (
	optionsSize            = "size"
	optionsUserMappingType = "user-mapping-type" // no_mapping, remap_root, remap_all
	optionsUserMappingUid  = "user-mapping-uid"
	optionsUserMappingGid  = "user-mapping-gid"
)

var (
	Ems               EmsWrapper       // Keep global to be reachable from ExportPath()
	defaultVolumeSize = 100 * size.GiB // TODO: Take from env
)

func (ems *EmsWrapper) initSession(details emsDetails) (client *emanage.Client, err error) {
	emsUrl := &url.URL{
		Scheme: "http",
		Host:   details.RestAddr,
	}
	client = emanage.NewClient(emsUrl)
	if client == nil {
		err = errors.New("Failed to create new EMS client")
		return
	}

	err = client.Sessions.Login(details.RestUser, details.RestPass)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to log into EMS", 0)
		return
	}
	logrus.Info("Logged into EMS")
	return
}

// Client is used to cache EMS login
// TODO: Consider removing this function and calling intSession() from newElastifileDriver()
func (ems *EmsWrapper) Client() (*emanage.Client, error) {
	if !ems.sessionInitialized {
		client, err := ems.initSession(driverInfo)
		if err != nil {
			return nil, errors.WrapPrefix(err, "Fatal error - failed to login to EMS", 0)
		}
		ems.client = client
		ems.sessionInitialized = true
	}
	return ems.client, nil
}

func (ems *EmsWrapper) defaultDcCreateOpts(name string) *emanage.DcCreateOpts {
	return &emanage.DcCreateOpts{
		Name:           name,
		DirPermissions: 777,
		Dedup:          0,
		Compression:    1,
	}
}

func (ems *EmsWrapper) defaultExportCreateOpts() *emanage.ExportCreateOpts {
	return &emanage.ExportCreateOpts{
		Path:        "/",
		Access:      emanage.ExportAccessRW,
		UserMapping: emanage.UserMappingAll,
		Uid:         optional.NewInt(0),
		Gid:         optional.NewInt(0),
	}
}

func (ems *EmsWrapper) defaultDcExportCreateOpts(name string) (*emanage.DcCreateOpts, *emanage.ExportCreateOpts) {
	return ems.defaultDcCreateOpts(name), ems.defaultExportCreateOpts()
}

func (ems *EmsWrapper) defaultPolicy() (policy emanage.Policy, err error) {
	if ems == nil {
		err = errors.New("Got nil EMS client")
		return
	}

	emsClient, err := ems.Client()
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to create EMS client", 0)
		return
	}

	var policies []emanage.Policy
	policies, err = emsClient.Policies.GetAll(nil)
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

func (ems *EmsWrapper) CreateDc(opts *emanage.DcCreateOpts) (dcRef *emanage.DataContainer, err error) {
	// TODO: Prune illegal characters
	name := fmt.Sprintf(opts.Name)
	// TODO: Support non-default policies via command line options and only use the default one if no policy was specified
	policy, err := ems.defaultPolicy()
	if err != nil {
		err = errors.WrapPrefix(err, fmt.Sprintf("Failed to get policy for volume %s", opts.Name), 0)
		return
	}

	logrus.WithFields(logrus.Fields{"name": name, "policy id": policy.Id, "opts": opts}).Debug("Creating Data Container")

	emsClient, err := ems.Client()
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to create EMS client", 0)
		return
	}

	dc, err := emsClient.DataContainers.Create(name, policy.Id, opts)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to create Data Container", 0)
		return
	}

	dcRef = &dc
	return
}

func (ems *EmsWrapper) CreateExport(name string, opts *emanage.ExportCreateOpts) (export emanage.Export, err error) {
	emsClient, err := ems.Client()
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to create EMS client", 0)
		return
	}

	return emsClient.Exports.Create(name, opts)
}

func (ems *EmsWrapper) dcExists(dcName string) (exists bool, dcRef *emanage.DataContainer, err error) {
	emsClient, err := ems.Client()
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to create EMS client", 0)
		return
	}

	dcs, err := emsClient.DataContainers.GetAll(nil)
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

func (ems *EmsWrapper) exportExists(exportName string, dcId int) (exists bool, exportRef *emanage.Export, err error) {

	emsClient, err := ems.Client()
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to create EMS client", 0)
		return
	}

	exports, err := emsClient.Exports.GetAll(nil)
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
func (ems *EmsWrapper) maybeCreateDc(dcOpts *emanage.DcCreateOpts) (*emanage.DataContainer, error) {
	exists, dc, err := ems.dcExists(dcOpts.Name)
	if err != nil {
		return nil, errors.WrapPrefix(err, "Failed to check if Data Container exists", 0)
	}
	if !exists {
		dc, err = ems.CreateDc(dcOpts)
		if err != nil {
			return nil, errors.WrapPrefix(err, "Failed to create Data Container", 0)
		}
	}
	return dc, nil
}

func (ems *EmsWrapper) maybeCreateExport(exportName string, exportOpts *emanage.ExportCreateOpts) (*emanage.Export, error) {
	exists, export, err := ems.exportExists(exportName, exportOpts.DcId)
	if err != nil {
		return nil, errors.WrapPrefix(err, "Failed to check if Export exists", 0)
	}
	if !exists {
		exp, err := ems.CreateExport(exportName, exportOpts)
		if err != nil {
			return nil, errors.WrapPrefix(err, "Failed to create Export", 0)
		}
		export = &exp
	} else {

	}
	return export, nil
}

func (ems *EmsWrapper) maybeCreateDcExport(dcOpts *emanage.DcCreateOpts, exportOpts *emanage.ExportCreateOpts) (
	export *emanage.Export, dc *emanage.DataContainer, err error) {

	// Create Data Container if it doesn't exist
	dc, err = ems.maybeCreateDc(dcOpts)
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}

	// Create Export if it doesn't exist
	exportName := "e"
	exportOpts.DcId = dc.Id
	export, err = ems.maybeCreateExport(exportName, exportOpts)
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}

	logrus.WithFields(logrus.Fields{"name": dc.Name, "id": dc.Id, "export": export.Name}).Info("Created DC, Export")
	return export, dc, nil
}

func (ems *EmsWrapper) DeleteDc(dc *emanage.DataContainer) (err error) {
	// TODO: Return success if DC doesn't exist

	emsClient, err := ems.Client()
	if err != nil {
		return errors.WrapPrefix(err, "Failed to create EMS client", 0)
	}

	_, err = emsClient.DataContainers.Delete(dc)
	return err
}

func (ems *EmsWrapper) DeleteExport(export *emanage.Export) (err error) {
	// TODO: Return success if export doesn't exist

	emsClient, err := ems.Client()
	if err != nil {
		return errors.WrapPrefix(err, "Failed to create EMS client", 0)
	}

	_, err = emsClient.Exports.Delete(export)
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

func (ems *EmsWrapper) DeleteDcExport(v *elastifileVolume) (err error) {
	err = ems.DeleteExport(v.Export)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to delete Export", 0)
		return
	}
	err = ems.DeleteDc(v.DataContainer)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to delete Data Container", 0)
		return
	}
	return
}

func (ems *EmsWrapper) dcExportPath(export *emanage.Export) (dir string, err error) {
	emsClient, err := ems.Client()
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to create EMS client", 0)
		return
	}

	dc, err := emsClient.DataContainers.GetFull(export.DataContainerId)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to get Data Containers", 0)
		return
	}
	dir = path.Join(dc.Name, export.Name)

	return
}
