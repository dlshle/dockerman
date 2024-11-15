package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dlshle/dockman/api/dmdaemon"
	"github.com/dlshle/dockman/internal/config"
	"gopkg.in/yaml.v2"
)

type Command struct {
	description string
	handler     func(context.Context, *dmdaemon.DmDaemonCli) error
}

var commands = map[string]*Command{
	"rollout":     {description: "Rollout an existing application", handler: handleRollout},
	"delete":      {description: "Delete an existing application", handler: handleDelete},
	"list":        {description: "List all applications", handler: handleList},
	"portforward": {description: "Portforward a local port to a remote port", handler: handlePortForward},
	"inspect":     {description: "Inspect an application", handler: handleInspect},
	"ping":        {description: "Ping the dmdaemon service", handler: handlePing},
}

var (
	host    string
	apiPort int
	pfPort  int

	// command params
	appName         string // appname
	appConfigPath   string // appcfg
	pfContainerName string // cname
	pfContainerID   string // cid
	pfTargetHost    string // th
	pfPortMappings  string // pm
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("no command specified")
	}

	cmd := os.Args[1]

	flag.StringVar(&host, "h", "127.0.0.1", "dockerman host address")
	flag.StringVar(&appName, "appname", "", "name of the application")
	flag.StringVar(&appConfigPath, "appcfg", "", "path to the application config file")
	flag.StringVar(&pfContainerName, "cname", "", "port forward target container name")
	flag.StringVar(&appConfigPath, "cid", "", "port forward target container id")
	flag.StringVar(&appConfigPath, "th", "", "port forward target host address")
	flag.StringVar(&appConfigPath, "pm", "", "port forward port mappings(source:dest), separated by comma")
	flag.IntVar(&apiPort, "p", 6300, "dmdaemon api service port")
	flag.IntVar(&pfPort, "pp", 0, "path to the application config file")

	flag.CommandLine.Parse(os.Args[2:])

	err := handleCommand(cmd)
	if err != nil {
		panic(err)
	}
}

func handleCommand(cmd string) error {
	if cmd == "" {
		return fmt.Errorf("no command specified")
	}
	command, ok := commands[cmd]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd)
	}
	cli, err := dmdaemon.NewCli(fmt.Sprintf("http://%s:%d", host, apiPort))
	if err != nil {
		panic(err)
	}
	return command.handler(context.Background(), cli)
}

func handlePing(ctx context.Context, cli *dmdaemon.DmDaemonCli) error {
	err := cli.Ping(ctx)
	if err != nil {
		return err
	}
	fmt.Println("ping success")
	return nil
}

func handleInspect(ctx context.Context, cli *dmdaemon.DmDaemonCli) error {
	desc, err := cli.GetDeployment(ctx, appName)
	if err != nil {
		return err
	}
	fmt.Println(desc)
	return nil
}

func handleList(ctx context.Context, cli *dmdaemon.DmDaemonCli) error {
	apps, err := cli.ListDeployments(ctx)
	if err != nil {
		return err
	}
	fmt.Println(apps)
	return nil
}

func handleRollout(ctx context.Context, cli *dmdaemon.DmDaemonCli) error {
	if appConfigPath == "" {
		return fmt.Errorf("application config file path(appcfg) must be specified")
	}
	appCfgData, err := os.ReadFile(appConfigPath)
	if err != nil {
		return err
	}
	appCfg := new(config.AppConfig)
	if err = yaml.Unmarshal(appCfgData, appCfg); err != nil {
		return err
	}
	return cli.Rollout(ctx, appCfg)
}

func handleDelete(ctx context.Context, cli *dmdaemon.DmDaemonCli) error {
	if appName == "" {
		return fmt.Errorf("application name(appname) must be specified")
	}
	return cli.DeleteDeployment(ctx, appName)
}

func handlePortForward(ctx context.Context, cli *dmdaemon.DmDaemonCli) error {
	if pfPort == 0 {
		return fmt.Errorf("dmdaemon portforward service port(pp) must be specified")
	}
	if err := validateInput(host, pfContainerName, pfContainerID, pfTargetHost, pfPortMappings); err != nil {
		log.Fatal(err)
	}

	portMappings, err := parsePortMappings(pfPortMappings)
	if err != nil {
		log.Fatal(err)
	}

	backendName := pfTargetHost
	if pfContainerName != "" {
		backendName = "containerName:" + pfContainerName
	} else if pfContainerID != "" {
		backendName = "containerID:" + pfContainerID
	}

	portForward := &PortForward{
		ProxyServerHost: host,
		BackendName:     backendName,
		PortMappings:    portMappings,
	}

	if err = listen(portForward); err != nil {
		return err
	}
	return nil
}
