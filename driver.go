package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"

	"github.com/elastifile/emanage-go/src/emanage-client"
	"github.com/elastifile/emanage-go/src/size"
)

type emsDetails struct {
	RestAddr    string
	RestUser    string
	RestPass    string
	StorageAddr string
	Root        string
}

var driverInfo = emsDetails{
	Root: "/mnt",
}

type elastifileDriver struct {
	sync.RWMutex

	managementAddr     string
	managementUser     string
	managementPassword string
	storageAddr        string
	root               string
	statePath          string
	volumes            map[string]*elastifileVolume
}

func newElastifileDriver(emsInfo emsDetails) (*elastifileDriver, error) {
	logrus.WithField("method", "new driver").Debug(emsInfo.Root)

	driver := &elastifileDriver{
		managementAddr:     emsInfo.RestAddr,
		managementUser:     emsInfo.RestUser,
		managementPassword: emsInfo.RestPass,
		storageAddr:        emsInfo.StorageAddr,
		root:               filepath.Join(emsInfo.Root, "volumes"),
		statePath:          filepath.Join(emsInfo.Root, "state", "elastifile-state.json"),
		volumes:            map[string]*elastifileVolume{},
	}

	data, err := ioutil.ReadFile(driver.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.WithField("statePath", driver.statePath).Debug("State not found")
		} else {
			err = errors.WrapPrefix(err, "Failed to load state", 0)
			return nil, err
		}
	} else {
		logrus.WithField("data", string(data)).Debug("Loaded state")
		err := json.Unmarshal(data, &driver.volumes)
		if err != nil {
			return nil, errors.WrapPrefix(err, "Failed to unmarshal state", 0)
		}
		logrus.Debug("Umarshaled state")
	}

	logrus.Debugf("%v driver created", pluginName)
	return driver, nil
}

func (d *elastifileDriver) saveState() {
	data, err := json.Marshal(d.volumes)
	if err != nil {
		logrus.WithField("statePath", d.statePath).Error(err)
		return
	}

	if err := ioutil.WriteFile(d.statePath, data, 0644); err != nil {
		logrus.WithField("savestate", d.statePath).Error(err)
	}
}

func (d *elastifileDriver) Create(r *volume.CreateRequest) (err error) {
	logrus.WithField("method", "create").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()
	v := &elastifileVolume{}

	dcCreateOpts, exportCreateOpts := Ems.defaultDcExportCreateOpts(r.Name)

	for key, val := range r.Options {
		switch key {
		case optionsSize:
			sizeVal, err := size.Parse(val)
			if err != nil {
				err = errors.WrapPrefix(err, "Failed to parse volume size", 0)
				logErrorAndReturn(err.Error(), "key", optionsSize, "value", val)
			}
			dcCreateOpts.HardQuota = int(sizeVal)
		case optionsUserMappingType:
			switch val {
			case string(emanage.UserMappingAll), string(emanage.UserMappingRoot), string(emanage.UserMappingNone):
				exportCreateOpts.UserMapping = emanage.UserMappingType(val)
			default:
				logErrorAndReturn("Unsupported user mapping type: %v", val)
			}
		case optionsUserMappingUid:
			uid, err := strconv.Atoi(val)
			if err != nil || uid < 0 {
				logErrorAndReturn("Unsupported UID value: %v", val)
			}
			exportCreateOpts.Uid = &uid
		case optionsUserMappingGid:
			gid, err := strconv.Atoi(val)
			if err != nil || gid < 0 {
				logErrorAndReturn("Unsupported GID value: %v", val)
			}
			exportCreateOpts.Gid = &gid
		default:
			// Currently extra arguments ignored. Consider passing to mount verbatim
			if val != "" {
				v.MountOpts = append(v.MountOpts, key+"="+val)
			} else {
				v.MountOpts = append(v.MountOpts, key)
			}
		}
	}

	if dcCreateOpts.HardQuota == 0 {
		dcCreateOpts.HardQuota = int(defaultVolumeSize)
		logrus.WithField("size", dcCreateOpts.HardQuota).Info("Using default volume size")
	}
	dcCreateOpts.SoftQuota = dcCreateOpts.HardQuota // Setting hard quota w/o soft quota fails

	logrus.WithField("name", r.Name).Debug("Creating Data Container and Export")
	exp, dc, err := Ems.maybeCreateDcExport(dcCreateOpts, exportCreateOpts)
	if err != nil {
		err = errors.WrapPrefix(err, "Failed to create Data Container / Export", 0)
		return err
	}

	v.Mountpoint = filepath.Join(d.root, r.Name)
	v.DataContainer = dc
	v.Export = exp

	d.volumes[r.Name] = v

	logrus.Debug("Saving state")
	d.saveState()

	return nil
}

func (d *elastifileDriver) Remove(r *volume.RemoveRequest) error {
	logrus.WithField("method", "remove").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return logErrorAndReturn("volume %s not found", r.Name)
	}

	if v.connections != 0 {
		return logErrorAndReturn("volume %s is currently used by a container", r.Name)
	}
	if err := os.RemoveAll(v.Mountpoint); err != nil {
		return logErrorAndReturn(err.Error())
	}

	// Handle DC/export removal
	Ems.DeleteDcExport(v)

	delete(d.volumes, r.Name)
	d.saveState()
	return nil
}

func (d *elastifileDriver) Path(r *volume.PathRequest) (*volume.PathResponse, error) {
	logrus.WithField("method", "path").Debugf("%#v", r)

	d.RLock()
	defer d.RUnlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return &volume.PathResponse{}, logErrorAndReturn("volume %s not found", r.Name)
	}

	return &volume.PathResponse{Mountpoint: v.Mountpoint}, nil
}

func (d *elastifileDriver) Mount(r *volume.MountRequest) (*volume.MountResponse, error) {
	logrus.WithField("method", "mount").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return &volume.MountResponse{}, logErrorAndReturn("volume %s not found", r.Name)
	}

	if v.connections == 0 {
		fi, err := os.Lstat(v.Mountpoint)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(v.Mountpoint, 0755); err != nil {
				return &volume.MountResponse{}, logErrorAndReturn(err.Error())
			}
		} else if err != nil {
			return &volume.MountResponse{}, logErrorAndReturn(err.Error())
		}

		if fi != nil && !fi.IsDir() {
			return &volume.MountResponse{}, logErrorAndReturn("%v already exist and it's not a directory", v.Mountpoint)
		}

		if err := d.mountVolume(v); err != nil {
			return &volume.MountResponse{}, logErrorAndReturn(err.Error())
		}
	}

	v.connections++

	return &volume.MountResponse{Mountpoint: v.Mountpoint}, nil
}

func (d *elastifileDriver) Unmount(r *volume.UnmountRequest) error {
	logrus.WithField("method", "unmount").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()
	v, ok := d.volumes[r.Name]
	if !ok {
		return logErrorAndReturn("volume %s not found", r.Name)
	}

	v.connections--

	if v.connections <= 0 {
		if err := d.unmountVolume(v.Mountpoint); err != nil {
			return logErrorAndReturn(err.Error())
		}
		v.connections = 0
	}

	return nil
}

func (d *elastifileDriver) Get(r *volume.GetRequest) (*volume.GetResponse, error) {
	logrus.WithField("method", "get").Debugf("%#v", r)

	d.Lock()
	defer d.Unlock()

	v, ok := d.volumes[r.Name]
	if !ok {
		return &volume.GetResponse{}, logErrorAndReturn("volume %s not found", r.Name)
	}

	return &volume.GetResponse{Volume: &volume.Volume{Name: r.Name, Mountpoint: v.Mountpoint}}, nil
}

func (d *elastifileDriver) List() (*volume.ListResponse, error) {
	logrus.WithField("method", "list").Debugf("")

	d.Lock()
	defer d.Unlock()

	var vols []*volume.Volume
	for name, v := range d.volumes {
		vols = append(vols, &volume.Volume{Name: name, Mountpoint: v.Mountpoint})
	}
	return &volume.ListResponse{Volumes: vols}, nil
}

func (d *elastifileDriver) Capabilities() *volume.CapabilitiesResponse {
	logrus.WithField("method", "capabilities").Debugf("")

	return &volume.CapabilitiesResponse{Capabilities: volume.Capability{Scope: "local"}}
}

func (d *elastifileDriver) mountVolume(v *elastifileVolume) error {
	logrus.Info("Mounting volume %s on %s", v.ExportPath(), v.Mountpoint)
	cmd := exec.Command("mount", "-o", "nolock", fmt.Sprintf("%v:%v", d.storageAddr, v.ExportPath()), v.Mountpoint)
	//cmd := exec.Command("mount", fmt.Sprintf("%v:%v", d.storageAddr, v.ExportPath()), v.Mountpoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return logErrorAndReturn("mount command execute failed: %v (%s)", err, output)
	}
	logrus.Debug("Mounted", output)
	return nil
}

func (d *elastifileDriver) unmountVolume(target string) error {
	logrus.Info("Unmounting %s", target)
	cmd := fmt.Sprintf("umount %s", target)
	logrus.Debug(cmd)
	return exec.Command("sh", "-c", cmd).Run()
}
