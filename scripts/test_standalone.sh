#!/bin/bash -e

# ------------------------------------------------------------------------------------------
# This script is  intended to be used on non-dev machines to test basic plugin functionality
#
# To override any defaults, use the following form of execution:
# PLUGIN_TAG=dev MGMT_ADDRESS=10.0.0.100 NFS_ADDRESS=10.0.0.200 MGMT_PASSWORD=changeme ./test_standalone.sh
# ------------------------------------------------------------------------------------------

# Set default values
: ${PLUGIN_TAG:="next"}
: ${MGMT_ADDRESS:="10.11.209.222"}
: ${NFS_ADDRESS:="172.16.0.1"}
: ${MGMT_USERNAME:="admin"}
: ${MGMT_PASSWORD:="changeme"}
: ${VOLUME:="vol1"}
: ${MOUNT:="/elastifile_mount"}
: ${FILE:="HELLO"}

HOST_MOUNT=/mnt/test-${RANDOM}
PLUGIN=elastifileio/edvp:${PLUGIN_TAG}
_STEP=1

function start_test () {
    echo ">>> Test step ${_STEP}: $@"
    ((_STEP++))
}

function log_pass () {
    echo ">>> [PASS]"
}

echo "### Cleanup - it's ok to se errors in this phase ###"
docker volume rm ${VOLUME} -f || true
docker plugin disable ${PLUGIN} || true
docker plugin rm ${PLUGIN} || true

echo "### Installing plugin ###"
docker plugin install --grant-all-permissions ${PLUGIN} MGMT_ADDRESS=${MGMT_ADDRESS} NFS_ADDRESS=${NFS_ADDRESS} MGMT_USERNAME=${MGMT_USERNAME} MGMT_PASSWORD=${MGMT_PASSWORD}
start_test "Plugin is reported by docker"
docker plugin inspect -f {{.Name}} ${PLUGIN} | grep ${PLUGIN}
log_pass

start_test "Plugin is enabled"
docker plugin inspect -f {{.Enabled}} ${PLUGIN} | grep "true"
log_pass

echo "### Testing ###"
docker volume create -d ${PLUGIN} --name ${VOLUME}
start_test "Volume is reported by docker"
docker volume inspect -f {{.Name}} ${VOLUME} | grep ${VOLUME}
log_pass

docker run --rm -it -v ${VOLUME}:${MOUNT} busybox touch ${MOUNT}/${FILE}
docker run --rm -it -v ${VOLUME}:${MOUNT} busybox ls -l ${MOUNT}/${FILE}

start_test "New container sees the created file"
docker run --rm -it -v ${VOLUME}:${MOUNT} busybox ls -1 ${MOUNT}/${FILE} | grep ${FILE}
log_pass

start_test "Volume uses proper driver, i.e. not the local one"
docker volume inspect -f {{.Driver}} ${VOLUME} | grep ${PLUGIN}
log_pass

start_test "File can be seen on the export directly"
mkdir ${HOST_MOUNT}
sudo mount ${NFS_ADDRESS}:${VOLUME}/e ${HOST_MOUNT}
sudo ls -1 ${HOST_MOUNT}/${FILE} | grep ${FILE}
log_pass

echo "### Teardown ###"
sudo umount ${HOST_MOUNT}
sudo rmdir ${HOST_MOUNT}
docker run --rm -it -v ${VOLUME}:${MOUNT} busybox rm -f ${MOUNT}/${FILE}
docker volume rm ${VOLUME}
docker plugin disable ${PLUGIN}
docker plugin rm ${PLUGIN}

echo "### Test completed successfully ###"

