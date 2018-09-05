# Docker volume plugin for Elastifile's ECFS

The plugin allows you to create and mount Elastifile's ECFS volumes in your docker environments

## Usage

* Install the plugin

```
$ docker plugin install --grant-all-permissions elastifileio/edvp MGMT_ADDRESS=10.11.209.222 NFS_ADDRESS=172.16.0.1 MGMT_USERNAME=myuser MGMT_PASSWORD=mypassword
latest: Pulling from elastifileio/edvp
dbef3f5c7798: Download complete 
Digest: sha256:bc0ef95b076b15ac35f3b89c771754e9eb4b1692cdfc5369e0f9dc6a2ea1566a
Status: Downloaded newer image for elastifileio/edvp:latest
Installed plugin elastifileio/edvp
$ docker plugin ls
ID                  NAME                     DESCRIPTION                           ENABLED
1ac8c5c1ae71        elastifileio/edvp:latest   Elastifile volume plugin for Docker   true
```

* Create a volume

```
$ docker volume create -d elastifileio/edvp --name myvolume1
myvolume1
$ docker volume ls
  DRIVER                     VOLUME NAME
  elastifileio/edvp:latest   myvolume1
```

* Use the volume

```
$ docker run --rm -it -v myvolume1:/elastifile_mount busybox touch /elastifile_mount/hello_world
$ docker run --rm -it -v myvolume1:/elastifile_mount busybox ls -l /elastifile_mount/hello_world
-rw-r--r--    1 root     root             0 Sep  5 15:04 /elastifile_mount/hello_world
```

* Create both the container and the volume in one command
```
docker run -it -d -v myvolume1:/elastifile_mount --volume-driver=elastifileio/edvp busybox touch /elastifile_mount/file1
```

* Delete the volume
```
$ docker volume rm myvolume1
myvolume1
```

* Upgrade the plugin
```
$ docker plugin disable elastifileio/edvp
elastifileio/edvp
$ docker plugin upgrade --grant-all-permissions elastifileio/edvp
Upgrading plugin elastifileio/edvp:latest from elastifileio/edvp:latest to elastifileio/edvp:latest
latest: Pulling from elastifileio/edvp
dbef3f5c7798: Download complete 
Digest: sha256:bc0ef95b076b15ac35f3b89c771754e9eb4b1692cdfc5369e0f9dc6a2ea1566a
Status: Downloaded newer image for elastifileio/edvp:latest
Upgraded plugin elastifileio/edvp:latest to docker.io/elastifileio/edvp:latest
$ docker plugin enable elastifileio/edvp
elastifileio/edvp
```

* Reconfigure the plugin
```
$ docker plugin disable elastifileio/edvp
elastifileio/edvp
$ docker plugin set elastifileio/edvp MGMT_ADDRESS=10.11.209.111
$ docker plugin enable elastifileio/edvp
elastifileio/edvp
```

* Uninstall the plugin
```
$ docker plugin disable elastifileio/edvp
elastifileio/edvp
$ docker plugin rm elastifileio/edvp
elastifileio/edvp
```

## LICENSE

MIT
