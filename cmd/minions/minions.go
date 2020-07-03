package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/pkg/errors"
	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	"github.com/projectcalico/libcalico-go/lib/clientv3"

	etcdV3 "github.com/coreos/etcd/clientv3"
	"github.com/projecteru2/minions/internal/driver"
	"github.com/projecteru2/minions/lib"
	"github.com/projecteru2/minions/versioninfo"
	log "github.com/sirupsen/logrus"
)

var (
	cnmName  string
	ipamName string
	debug    bool

	config    *apiconfig.CalicoAPIConfig
	client    clientv3.Interface
	dockerCli *dockerClient.Client
	etcd      *etcdV3.Client
)

func initializeClient() {
	if debug {
		log.SetLevel(log.DebugLevel)
		log.Debugln("Debug logging enabled")
	}

	var err error

	if config, err = apiconfig.LoadClientConfig(""); err != nil {
		panic(err)
	}
	if client, err = clientv3.New(*config); err != nil {
		panic(err)
	}

	if dockerCli, err = dockerClient.NewClientWithOpts(dockerClient.FromEnv); err != nil {
		log.Fatalln(errors.Wrap(err, "Error while attempting to instantiate docker client from env"))
	}

	if etcd, err = lib.NewEtcdClient(strings.Split(config.Spec.EtcdConfig.EtcdEndpoints, ",")); err != nil {
		log.Fatalln(err)
	}
}

func serve() {
	initializeClient()

	errChannel := make(chan error)

	ripam := driver.NewReservedIPManager(etcd)
	networkHandler := network.NewHandler(driver.NewNetworkDriver(client, dockerCli, ripam))
	ipamHandler := ipam.NewHandler(driver.NewIpamDriver(client, ripam))

	go func(c chan error) {
		log.Infoln("calico-net has started.")
		err := networkHandler.ServeUnix(cnmName, 0)
		log.Infoln("calico-net has stopped working.")
		c <- err
	}(errChannel)

	go func(c chan error) {
		log.Infoln("calico-ipam has started.")
		err := ipamHandler.ServeUnix(ipamName, 0)
		log.Infoln("calico-ipam has stopped working.")
		c <- err
	}(errChannel)

	err := <-errChannel

	log.Fatalln(err)
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(versioninfo.VersionString())
	}

	app := cli.NewApp()
	app.Name = versioninfo.NAME
	app.Usage = "Run eru minions"
	app.Version = versioninfo.VERSION
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "cnm",
			Value:       "calico",
			Usage:       "CNM name",
			Destination: &cnmName,
			EnvVar:      "CALICO_CNM",
		},
		cli.StringFlag{
			Name:        "ipam",
			Value:       "calico-ipam",
			Usage:       "ipam name",
			Destination: &ipamName,
			EnvVar:      "CALICO_IPAM",
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "debug or not",
			Destination: &debug,
			EnvVar:      "CALICO_DEBUG",
		},
	}
	app.Action = func(c *cli.Context) error {
		serve()
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
