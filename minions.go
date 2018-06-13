package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/codegangsta/cli"
	"github.com/projecteru2/minions/versioninfo"
)

//func setupLog(l string) error {
//	level, err := log.ParseLevel(l)
//	if err != nil {
//		return err
//	}
//	log.SetLevel(level)
//
//	formatter := &log.TextFormatter{
//		TimestampFormat: "2006-01-02 15:04:05",
//		FullTimestamp:   true,
//	}
//	log.SetFormatter(formatter)
//	return nil
//}
//
//func serve() {
//	if configPath == "" {
//		log.Fatal("[main] Config path must be set")
//	}
//
//	config, err := utils.LoadConfig(configPath)
//	if err != nil {
//		log.Fatalf("[main] %v", err)
//	}
//
//	logLevel := "INFO"
//	if config.LogLevel != "" {
//		logLevel = config.LogLevel
//	}
//	if err := setupLog(logLevel); err != nil {
//		log.Fatalf("[main] %v", err)
//	}
//
//	stats.NewStatsdClient(config.Statsd)
//
//	cluster, err := calcium.New(config)
//	if err != nil {
//		log.Fatalf("[main] %v", err)
//	}
//
//	vibranium := rpc.New(cluster, config)
//	s, err := net.Listen("tcp", config.Bind)
//	if err != nil {
//		log.Fatalf("[main] %v", err)
//	}
//
//	opts := []grpc.ServerOption{grpc.MaxConcurrentStreams(100)}
//	grpcServer := grpc.NewServer(opts...)
//	pb.RegisterCoreRPCServer(grpcServer, vibranium)
//	go grpcServer.Serve(s)
//	if config.Profile != "" {
//		go http.ListenAndServe(config.Profile, nil)
//	}
//
//	log.Info("[main] Cluster started successfully.")
//
//	// wait for unix signals and try to GracefulStop
//	sigs := make(chan os.Signal, 1)
//	signal.Notify(sigs, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
//	sig := <-sigs
//	log.Infof("[main] Get signal %v.", sig)
//	grpcServer.GracefulStop()
//	log.Info("[main] gRPC server gracefully stopped.")
//
//	log.Info("[main] Check if cluster still have running tasks.")
//	vibranium.Wait()
//	log.Info("[main] cluster gracefully stopped.")
//}

var (
	cniBin      string
	cniConfig   string
	cnmSockFile string
)

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
			Name:        "cni_bin",
			Value:       "/opt/bin/cni",
			Usage:       "CNI plugins path",
			Destination: &cniBin,
			EnvVar:      "MINIONS_CNI_PATH",
		},
		cli.StringFlag{
			Name:        "cni_config",
			Value:       "/etc/cni/net.d",
			Usage:       "CNI configs path",
			Destination: &cniConfig,
			EnvVar:      "MINIONS_CNI_CONFS",
		},
		cli.StringFlag{
			Name:        "cnm_sock",
			Value:       "/run/docker/plugins",
			Usage:       "CNM plugins path",
			Destination: &cnmSockFile,
			EnvVar:      "MINIONS_CNM_SOCK",
		},
	}
	app.Action = func(c *cli.Context) error {
		//serve()
		return nil
	}

	app.Run(os.Args)
}
