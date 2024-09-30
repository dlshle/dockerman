package dockerx

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dlshle/gommon/slices"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerClient struct {
	cli *client.Client
}

func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerClient{cli: cli}, nil
}

func (dc *DockerClient) ListImages(ctx context.Context, params map[string]string) ([]*Image, error) {
	images, err := dc.cli.ImageList(ctx, image.ListOptions{All: true, Filters: dc.fromArgsMap(params)})
	if err != nil {
		return nil, err
	}

	var result []*Image
	for _, img := range images {
		result = append(result, &Image{
			ID:       img.ID,
			RepoTags: img.RepoTags,
			Created:  img.Created,
			Size:     img.Size,
		})
	}
	return result, nil
}

func (dc *DockerClient) ListNetworks(ctx context.Context, params map[string]string) ([]*Network, error) {
	networks, err := dc.cli.NetworkList(ctx, network.ListOptions{Filters: dc.fromArgsMap(params)})
	var resultNetworks []*Network
	if err != nil {
		return nil, err
	}
	for _, network := range networks {
		peers := make(map[string]string)
		for _, peer := range network.Peers {
			peers[peer.Name] = peer.IP
		}
		resultNetworks = append(resultNetworks, &Network{
			Driver:   network.Driver,
			Internal: network.Internal,
			Labels:   network.Labels,
			Name:     network.Name,
			Peers:    peers,
		})
	}
	return resultNetworks, nil
}

func (dc *DockerClient) ListContainers(ctx context.Context, params map[string]string) ([]*Container, error) {
	containers, err := dc.cli.ContainerList(ctx, container.ListOptions{All: true, Filters: dc.fromArgsMap(params)})
	if err != nil {
		return nil, err
	}

	var result []*Container
	for _, container := range containers {
		networkIP := make(map[string]string)
		for networkName, network := range container.NetworkSettings.Networks {
			networkIP[networkName] = network.IPAddress
		}
		result = append(result, &Container{
			ID:           container.ID,
			Names:        container.Names,
			Labels:       container.Labels,
			Image:        container.Image,
			ImageID:      container.ImageID,
			State:        container.State,
			Status:       container.Status,
			IPAddresses:  networkIP,
			ExposedPorts: slices.Map(container.Ports, func(port types.Port) uint16 { return port.PublicPort }),
		})
	}
	return result, nil
}

func (dc *DockerClient) RunImage(ctx context.Context, options *RunOptions) (string, error) {
	config := &container.Config{
		Image:  options.Image,
		Env:    options.Envs,
		Labels: options.Labels,
	}

	hostConfig := &container.HostConfig{
		Binds: options.VolumeMapping,
	}

	if len(options.Networks) > 0 {
		hostConfig.NetworkMode = container.NetworkMode(options.Networks[0])
	}

	portBindings := nat.PortMap{}
	for hostPort, containerPort := range options.PortMapping {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return "", err
		}
		portBindings[port] = []nat.PortBinding{
			{
				HostPort: hostPort,
			},
		}
	}
	hostConfig.PortBindings = portBindings

	resp, err := dc.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, options.ContainerName)
	if err != nil {
		return "", err
	}

	if options.Detached {
		if err := dc.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return "", err
		}
	} else {
		if err := dc.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return "", err
		}
		statusCh, errCh := dc.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				return "", err
			}
		case <-statusCh:
		}
	}

	// Connect the container to additional networks if specified
	for _, network := range options.Networks[1:] {
		err := dc.cli.NetworkConnect(ctx, network, resp.ID, nil)
		if err != nil {
			return "", err
		}
	}

	return resp.ID, nil
}

func (dc *DockerClient) CreateNetwork(ctx context.Context, name string) (string, error) {
	response, err := dc.cli.NetworkCreate(ctx, name, network.CreateOptions{})
	if err != nil {
		return "", err
	}
	return response.ID, nil
}

func (dc *DockerClient) RenameContainer(ctx context.Context, containerID string, newName string) error {
	return dc.cli.ContainerRename(ctx, containerID, newName)
}

func (dc *DockerClient) StopContainer(ctx context.Context, containerID string) error {
	return dc.cli.ContainerStop(ctx, containerID, container.StopOptions{})
}

func (dc *DockerClient) StartContainer(ctc context.Context, containerID string) error {
	return dc.cli.ContainerStart(ctc, containerID, container.StartOptions{})
}

func (dc *DockerClient) RemoveContainer(ctx context.Context, containerID string) error {
	return dc.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})
}

func (dc *DockerClient) ReadFileContentFromContainer(ctx context.Context, containerID, containerFilePath string) ([]byte, error) {
	reader, stat, err := dc.cli.CopyFromContainer(ctx, containerID, containerFilePath)
	defer reader.Close()
	if err != nil {
		return nil, err
	}
	if !stat.Mode.IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", containerFilePath)
	}
	return io.ReadAll(reader)
}

func (dc *DockerClient) CopyFileToContainer(ctx context.Context, containerID, hostFilePath, containerDestPath string) error {
	// Get the file to copy
	file, err := os.Open(hostFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a buffer to write our archive to
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// Create the header
	hdr := &tar.Header{
		Name: filepath.Base(hostFilePath),
		Mode: int64(stat.Mode()),
		Size: stat.Size(),
	}

	// Write the header
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	// Copy the file data to the tar writer
	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	// Close the tar writer
	if err := tw.Close(); err != nil {
		return err
	}

	// Create a read buffer from the tar archive
	tarReader := bytes.NewReader(buf.Bytes())

	// Copy the tar archive to the container's file system
	return dc.cli.CopyToContainer(ctx, containerID, containerDestPath, tarReader, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
}

func (dc *DockerClient) fromArgsMap(args map[string]string) filters.Args {
	if args == nil {
		return filters.NewArgs()
	}
	filterArgs := make([]filters.KeyValuePair, 0)
	for k, v := range args {
		filterArgs = append(filterArgs, filters.Arg(k, v))
	}
	return filters.NewArgs(filterArgs...)
}
func (dc *DockerClient) ExecContainer(ctx context.Context, containerID string, cmd []string) error {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := dc.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return err
	}

	execStartCheck := container.ExecStartOptions{
		Detach: false,
		Tty:    false,
	}

	execResp, err := dc.cli.ContainerExecAttach(ctx, execIDResp.ID, execStartCheck)
	if err != nil {
		return err
	}
	defer execResp.Close()

	execOutput := new(bytes.Buffer)
	io.Copy(execOutput, execResp.Reader)
	fmt.Println(execOutput.String())

	// Check if the command executed successfully
	inspectResp, err := dc.cli.ContainerExecInspect(ctx, execIDResp.ID)
	if err != nil {
		return err
	}
	if inspectResp.ExitCode != 0 {
		return fmt.Errorf("exec failed with exit code %d", inspectResp.ExitCode)
	}

	return nil
}
