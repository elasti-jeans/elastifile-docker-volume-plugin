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
```
$ PLUGIN_TAG=dev make push
$ scp scripts/test_standalone.sh <user@remote_machine_wth_storage_access>:
$ ssh <user@remote_machine_wth_storage_access>
$ PLUGIN_TAG=dev MGMT_ADDRESS=10.11.209.222 NFS_ADDRESS=172.16.0.1 MGMT_PASSWORD=changeme ./test_standalone.sh 
### Cleanup - it's ok to see errors in this phase ###
vol1
Error response from daemon: plugin "elastifileio/edvp:dev" not found
Error: No such plugin: elastifileio/edvp:dev
##### Testing CRUD_IDEMPOTENT=true #####
### Installing plugin ###
dev: Pulling from elastifileio/edvp
0bd45ee9b496: Download complete
Digest: sha256:73cc962ffe1497e94b41975952eac56c8162b73a7d8ce1fc08e1bb1c4941ffc2
Status: Downloaded newer image for elastifileio/edvp:dev
Installed plugin elastifileio/edvp:dev
>>> Test step 1: Installed plugin is reported by docker
elastifileio/edvp:dev
>>> [PASS]
>>> Test step 2: Installed plugin is enabled
true
>>> [PASS]
>>> Test step 3: Create new volume
vol1
vol1
>>> [PASS]
>>> Test step 4: Create existing volume
vol1
vol1
>>> [PASS]
>>> Test step 5: Create a file and list it from a different container, i.e. the file persists across container
/elastifile_mount/HELLO
>>> [PASS]
>>> Test step 6: New volume uses the requested driver, i.e. there was no fallback to local volumes
elastifileio/edvp:dev
>>> [PASS]
>>> Test step 7: File created earlier can be seen on the export directly, i.e. the file is located on the export
/mnt/test-822/HELLO
>>> [PASS]
...
### Teardown ###
>>> Test step 17: Delete the volume
vol1
DRIVER                  VOLUME NAME
local                   6f2a8c7d2b2dc9abcc4b2ca8081d0329e229faa229b43034e9f9e667dc6cd74d
local                   d75d07bc4917318a4d243890aaa96c787b28883d16a9bb3137a072e40ceb703b
elastifileio/edvp:dev   myvolume1
>>> [PASS]
>>> Test step 18: Disable the plugin
elastifileio/edvp:dev
false
>>> [PASS]
>>> Test step 19: Remove the plugin
elastifileio/edvp:dev
ID                  NAME                DESCRIPTION         ENABLED
>>> [PASS]
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
ssh$ cat /mnt/state/elastifile-state.json
```
