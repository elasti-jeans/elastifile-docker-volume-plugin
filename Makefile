PLUGIN_NAME = elastifileio/edvp
PLUGIN_TAG ?= next
REGISTRY=hub.docker.com

TEST_VOLUME_NAME=myvolume1
TEST_MOUNT_POINT=/elastifile_data
TEST_FILE_NAME=testfile

all: clean rootfs create

clean:
	@echo "### rm ./plugin"
	@rm -rf ./plugin

rootfs:
	@echo "### docker build: rootfs image with docker-volume-elastifile"
	@docker build -q -t ${PLUGIN_NAME}:rootfs .
	@echo "### create rootfs directory in ./plugin/rootfs"
	@mkdir -p ./plugin/rootfs
	@docker create --name tmp ${PLUGIN_NAME}:rootfs
	@docker export tmp | tar -x -C ./plugin/rootfs
	@echo "### copy config.json to ./plugin/"
	@cp config.json ./plugin/
	@docker rm -vf tmp

create:
	@echo "### remove existing plugin ${PLUGIN_NAME}:${PLUGIN_TAG} if exists"
	@docker plugin rm -f ${PLUGIN_NAME}:${PLUGIN_TAG} || true
	@echo "### create new plugin ${PLUGIN_NAME}:${PLUGIN_TAG} from ./plugin"
	@docker plugin create ${PLUGIN_NAME}:${PLUGIN_TAG} ./plugin

enable:		
	@echo "### enable plugin ${PLUGIN_NAME}:${PLUGIN_TAG}"		
	@docker plugin enable ${PLUGIN_NAME}:${PLUGIN_TAG}

test:
	@echo "### test"
	@echo "### remove volume"
	@docker volume rm ${TEST_VOLUME_NAME} || true
	@echo "### create volume"
	@docker volume create -d ${PLUGIN_NAME}:${PLUGIN_TAG} --name ${TEST_VOLUME_NAME} -o size=12GiB
	@echo "### list volumes #1"
	@docker volume ls
	@echo "### create file ${TEST_FILE_NAME} in one container"
	@docker run --rm -it -v ${TEST_VOLUME_NAME}:${TEST_MOUNT_POINT} busybox touch ${TEST_MOUNT_POINT}/${TEST_FILE_NAME}
	@echo "### list file ${TEST_FILE_NAME} in another container"
	@docker run --rm -it -v ${TEST_VOLUME_NAME}:${TEST_MOUNT_POINT} busybox ls -l ${TEST_MOUNT_POINT}/${TEST_FILE_NAME}
	@echo "### list volumes #2"
	@docker volume ls
	@echo "Now check that file ${TEST_FILE_NAME} is present on the export when the latter is mounted from another location, and the volume is NOT local in #2 above"

local: clean rootfs create enable test

push:  clean rootfs create enable
	@echo "### push plugin ${PLUGIN_NAME}:${PLUGIN_TAG}"
	@docker plugin push ${PLUGIN_NAME}:${PLUGIN_TAG}
