package main

import (
	"flag"
	"log"
	"strconv"
	"sync"

	"github.com/dlshle/dockman/internal/deployment"
	"github.com/dlshle/dockman/internal/frontend"
	"github.com/dlshle/dockman/internal/handler"
	"github.com/dlshle/dockman/internal/portforward"
	"github.com/dlshle/dockman/pkg/dockerx"
)

func main() {
	var (
		dockerHost string
		fApiPort   string
		fProxyPort string
	)
	flag.StringVar(&dockerHost, "h", "127.0.0.1:2375", "docker host(e.g. 127.0.0.1:2375) on machine")
	flag.StringVar(&fApiPort, "ap", "6300", "dockman api port")
	flag.StringVar(&fProxyPort, "pp", "14514", "dockman proxy port")
	dockerCli, err := dockerx.NewDockerClient(dockerHost)
	if err != nil {
		log.Fatal(err)
	}

	apiPort, err := strconv.Atoi(fApiPort)
	if err != nil {
		log.Fatalf("failed to parse api port %s: %v", fApiPort, err)
	}
	proxyPort, err := strconv.Atoi(fProxyPort)
	if err != nil {
		log.Fatalf("failed to parse proxy port %s: %v", fProxyPort, err)
	}

	deploymentHandler := deployment.NewDeployment(dockerCli)
	portforwardHandler := portforward.NewPortforwardHandler(proxyPort, dockerCli)
	dmHandler := handler.NewDockmanHandler(dockerCli, deploymentHandler, portforwardHandler)

	var wg sync.WaitGroup

	// dockman http server
	wg.Add(1)
	go func() {
		if err = frontend.ServeHTTP(apiPort, dmHandler); err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()

	// proxy server
	wg.Add(1)
	go func() {
		err := dmHandler.StartPortforwardServer()
		if err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()

	wg.Wait()

	// TODO dockerman daemon(server)
	// inputs:
	//   * deployment server port
	//   * proxy server port
	// modules:
	//   * docker client(ok)
	//   * deployment controller(ok)
	//   * gateway controller(not yet)
	//   * proxy controller(OK)
	//   * probe controller(ok?)
}
