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

## Testing
* Local test
```bash
MGMT_ADDRESS=10.11.209.222 NFS_ADDRESS=172.16.0.1 make all test
```
Note: "test" target is only useful if your machine has access to ECFS' storage address. If not:

* Remote test
```bash
$ PLUGIN_TAG=dev make push
$ scp scripts/test_standalone.sh <user@remote_machine_wth_storage_access>:
$ ssh <user@remote_machine_wth_storage_access>
$ PLUGIN_TAG=dev MGMT_ADDRESS=10.11.209.222 NFS_ADDRESS=172.16.0.1 MGMT_PASSWORD=changeme ./test_standalone.sh 
### Cleanup - it's ok to se errors in this phase ###
vol1
Error response from daemon: plugin "elastifileio/edvp:dev" not found
Error: No such plugin: elastifileio/edvp:dev
### Installing plugin ###
dev: Pulling from elastifileio/edvp
724cb075746b: Download complete 
Digest: sha256:53432ace774cec7aa8f5aac54a022413c62fa51650c7d26a8505adb6b4f6257e
Status: Downloaded newer image for elastifileio/edvp:dev
Installed plugin elastifileio/edvp:dev
>>> Test step 1: Plugin is reported by docker
elastifileio/edvp:dev
>>> [PASS]
>>> Test step 2: Plugin is enabled
true
>>> [PASS]
### Testing ###
vol1
>>> Test step 3: Volume is reported by docker
vol1
>>> [PASS]
-rw-r--r--    1 root     root             0 Sep  8 17:46 /elastifile_mount/HELLO
>>> Test step 4: New container sees the created file
/elastifile_mount/HELLO
>>> [PASS]
>>> Test step 5: Volume uses proper driver, i.e. not the local one
elastifileio/edvp:dev
>>> [PASS]
>>> Test step 6: File can be seen on the export directly
/mnt/test-19266/HELLO
>>> [PASS]
### Teardown ###
vol1
elastifileio/edvp:dev
elastifileio/edvp:dev
### Test completed successfully ###
```
Notes:

The user on the remote machine should be in the sudoers list

Preferably, the user should not require password when doing sudo 

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

* Connect to the plugin's container
```bash
$ docker-runc --root /var/run/docker/plugins/runtime-root/moby-plugins list
$ docker-runc --root /var/run/docker/plugins/runtime-root/moby-plugins exec -t <id> sh
# cat /mnt/state/elastifile-state.json
```
