#!/bin/bash

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
: ${VOLUME:="vol1"} # Create volume by this name
: ${EXPORT:="root"} # Expect this export to be created by the plugin
: ${MOUNT:="/elastifile_mount"} # Mount the volume on this path inside the container
: ${FILE:="HELLO"} # Create file by this name

HOST_MOUNT=/mnt/test-${RANDOM}
PLUGIN=elastifileio/edvp:${PLUGIN_TAG}
_STEP=1

COLOR_GREEN='\033[0;32m'
COLOR_LIGHT_BLUE='\033[1;34m'
COLOR_RESET='\033[0m'

set -e

function start_test () {
    echo -e ">>> ${COLOR_LIGHT_BLUE}Test step ${_STEP}: $@${COLOR_RESET}"
    ((_STEP++))
}

function log_pass () {
    echo -e ">>> ${COLOR_GREEN}[PASS]${COLOR_RESET}"
}

echo "### Cleanup - it's ok to see errors in this phase ###"
docker volume rm --force ${VOLUME} || true
docker plugin disable --force ${PLUGIN} || true
docker plugin rm --force ${PLUGIN} || true

for CRUD_IDEMPOTENT in true false; do
    echo "##### Testing CRUD_IDEMPOTENT=${CRUD_IDEMPOTENT} #####"

    echo "### Installing plugin ###"
    docker plugin install --grant-all-permissions ${PLUGIN} MGMT_ADDRESS=${MGMT_ADDRESS} NFS_ADDRESS=${NFS_ADDRESS} MGMT_USERNAME=${MGMT_USERNAME} MGMT_PASSWORD=${MGMT_PASSWORD} CRUD_IDEMPOTENT=${CRUD_IDEMPOTENT} DEBUG=true

    start_test "Installed plugin is reported by docker"
    docker plugin inspect -f {{.Name}} ${PLUGIN} | grep ${PLUGIN}
    log_pass

    start_test "Installed plugin is enabled"
    docker plugin inspect -f {{.Enabled}} ${PLUGIN} | grep "true"
    log_pass

    start_test "Create new volume"
    docker volume create -d ${PLUGIN} --name ${VOLUME}
    docker volume inspect -f {{.Name}} ${VOLUME} | grep ${VOLUME}
    log_pass

    if [ ${CRUD_IDEMPOTENT} == true ]; then
        start_test "Create existing volume"
        docker volume create -d ${PLUGIN} --name ${VOLUME}
        docker volume inspect -f {{.Name}} ${VOLUME} | grep ${VOLUME}
        log_pass
    fi

    start_test "Create a file and list it from a different container, i.e. the file persists across container"
    docker run --rm -it -v ${VOLUME}:${MOUNT} busybox touch ${MOUNT}/${FILE}
    docker run --rm -it -v ${VOLUME}:${MOUNT} busybox ls -1 ${MOUNT}/${FILE} | grep ${FILE}
    log_pass

    start_test "New volume uses the requested driver, i.e. there was no fallback to local volumes"
    docker volume inspect -f {{.Driver}} ${VOLUME} | grep ${PLUGIN}
    log_pass

    start_test "File created earlier can be seen on the export directly, i.e. the file is located on the export"
    mkdir ${HOST_MOUNT}
    sudo mount ${NFS_ADDRESS}:${VOLUME}/${EXPORT} ${HOST_MOUNT}
    sudo ls -1 ${HOST_MOUNT}/${FILE} | grep ${FILE}
    log_pass

    echo "### Teardown ###"
    sudo umount ${HOST_MOUNT}
    sudo rmdir ${HOST_MOUNT}
    docker run --rm -it -v ${VOLUME}:${MOUNT} busybox rm -f ${MOUNT}/${FILE}

    start_test "Delete the volume"
    docker volume rm ${VOLUME}
    docker volume ls | grep -v ${VOLUME}
    log_pass

    # docker won't allow removing a non-existent volume, i.e. idempotent volume removal can only be tested from a different host

    start_test "Disable the plugin"
    docker plugin disable ${PLUGIN}
    docker plugin inspect -f {{.Enabled}} ${PLUGIN} | grep "false"
    log_pass

    start_test "Remove the plugin"
    docker plugin rm ${PLUGIN}
    docker plugin ls | grep -v ${PLUGIN}
    log_pass
done

echo -e "### Test completed ${COLOR_GREEN}successfully${COLOR_RESET} ###"
