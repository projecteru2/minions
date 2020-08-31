Minions
=======

![golangci-lint](https://github.com/projecteru2/minions/workflows/golangci-lint/badge.svg)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/d4cf457004844c0f80bb372237159e70)](https://www.codacy.com/app/projecteru2/minions?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=projecteru2/minions&amp;utm_campaign=Badge_Grade)

Since [calico-libnetwork-plugin](https://github.com/projectcalico/libnetwork-plugin) not support docker anymore. I built this for eru project to use calico network with latest version.

### Build

#### build binary

`make build`

#### build rpm

`./make-rpm`

#### build image

`docker build -t minions .`

### Develop

```shell
go get github.com/projecteru2/minions
cd $GOPATH/src/get github.com/projecteru2/minions
make deps
```

### Dockerized Minions manually

```shell
docker run -d \
  --name eru_minions_$HOSTNAME \
  --net host \
  --restart always \
  -v /var/run/docker/plugins/:/var/run/docker/plugins \
  projecteru2/minions \
  /usr/bin/eru-minions
```

### Build and Deploy by Eru itself

After we implemented bootstrap in eru2, now you can build and deploy minions with [cli](https://github.com/projecteru2/cli) tool.

1. Test source code and build image

```shell
<cli_execute_path> --name <image_name> https://bit.ly/EruMinions
```

Make sure you can clone code by ssh protocol because libgit2 ask for it. So you need configure core with github certs. After the fresh image was named and tagged, it will be auto pushed to the remote registry which was defined in core.

2. Deploy minions by eru with specific resource.

```shell
<cli_execute_path> container deploy --pod <pod_name> --entry minions --network <network_name> --deploy-method fill --image <projecteru2/minions>|<your_own_image> --count 1 --env ETCD_ENDPOINTS=${ETCD_ENDPOINTS} [--cpu 0.3 | --mem 1024000000] https://bit.ly/EruMinions
```

Now you will find minions was started in each node.

# Install with github releases
Unarchive and run command with sudo
```shell
sudo ./install.sh
``` 
