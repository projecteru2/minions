package main

import (
	"fmt"
	"os"

	pluginIPAM "github.com/docker/go-plugins-helpers/ipam"
	pluginNetwork "github.com/docker/go-plugins-helpers/network"
	"github.com/pkg/errors"
	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	calicov3 "github.com/projectcalico/libcalico-go/lib/clientv3"
	cli "github.com/urfave/cli/v2"

	dockerClient "github.com/docker/docker/client"
	"github.com/projecteru2/minions/barrel"
	"github.com/projecteru2/minions/barrel/etcd"
	"github.com/projecteru2/minions/driver"
	"github.com/projecteru2/minions/versioninfo"
	log "github.com/sirupsen/logrus"

	_ "go.uber.org/automaxprocs"
)

func serve(c *cli.Context) error {
	log.SetOutput(os.Stdout)

	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
		log.Debugln("Debug logging enabled")
	}

	var (
		config     *apiconfig.CalicoAPIConfig
		calicoCli  calicov3.Interface
		barrelMeta barrel.Meta
		dockerCli  *dockerClient.Client
		err        error
	)

	if config, err = apiconfig.LoadClientConfig(""); err != nil {
		return err
	}
	if calicoCli, err = calicov3.New(*config); err != nil {
		return err
	}
	if barrelMeta, err = etcd.NewEtcdClient(c.Context, *config); err != nil {
		return err
	}
	if dockerCli, err = dockerClient.NewClientWithOpts(dockerClient.FromEnv); err != nil {
		return errors.Wrap(err, "Error while attempting to instantiate docker client from env")
	}

	errChannel := make(chan error)
	networkHandler := pluginNetwork.NewHandler(driver.NewNetworkDriver(calicoCli, dockerCli, barrelMeta))
	ipamHandler := pluginIPAM.NewHandler(driver.NewIPAMDriver(calicoCli, barrelMeta))

	go func() {
		log.Infoln("calico-net has started.")
		err := networkHandler.ServeUnix(c.String("cnm"), 0)
		log.Infoln("calico-net has stopped working.")
		errChannel <- err
	}()

	go func() {
		log.Infoln("calico-ipam has started.")
		err := ipamHandler.ServeUnix(c.String("ipam"), 0)
		log.Infoln("calico-ipam has stopped working.")
		errChannel <- err
	}()

	return <-errChannel
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
		&cli.StringFlag{
			Name:    "cnm",
			Value:   "calico",
			Usage:   "CNM name",
			EnvVars: []string{"CALICO_CNM"},
		},
		&cli.StringFlag{
			Name:    "ipam",
			Value:   "calico-ipam",
			Usage:   "ipam name",
			EnvVars: []string{"CALICO_IPAM"},
		},
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "debug or not",
			EnvVars: []string{"CALICO_DEBUG"},
		},
	}
	app.Action = serve

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
