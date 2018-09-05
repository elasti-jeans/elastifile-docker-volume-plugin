# Docker volume plugin for Elastifile's ECFS

This README is intended for plugin developers

## Dependencies

* Fetch plugin's dependencies - useful to bring your project up-to-speed
```bash
dep ensure
```

* Update specific dependency to the newest version of a 3rd party package
```bash
dep ensure -no-vendor -update github.com/elastifile/emanage-go
```

* Update specific dependency to the version in your tree, but don't overwrite the package
Useful if you're updating the dependency while working on the main project (not recommended as you're prone to lose your work)
```bash
dep ensure -no-vendor -update github.com/elastifile/emanage-go
```

* Update all dependencies
```bash
dep ensure -update
```

## Building

* Build locally
```bash
make all
```

* Build and push to Docker Hub, with "next" tag
```bash
make push
```

* Build and push to Docker Hub, with custom tag, e.g. "your_name"
```bash
PLUGIN_TAG=your_name make push
```

* Build stable version
```bash
git tag "RELEASE-TAG-HERE"
git push --tags origin master
PLUGIN_TAG=latest make push
```

* Test
```bash
MGMT_ADDRESS=10.11.209.222 NFS_ADDRESS=172.16.0.1 make all test
```
Note: "test" target is only useful if your machine has access to ECFS' storage address. If not:
- Push the plugin with your custom tag
- Install the plugin on a docker host with access to the storage address
- Test according to README.md

## Troubleshooting
* Examine the plugin logs
```bash
journalctl -u docker
```

* Enable debug logs
```bash
docker plugin disable elastifileio/edvp:next
docker plugin set elastifileio/edvp:next DEBUG=true
docker plugin enable elastifileio/edvp:next
```
