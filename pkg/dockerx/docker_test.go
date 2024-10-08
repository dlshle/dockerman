package dockerx

import (
	"context"
	"testing"
)

func assertError(t *testing.T, msg string, err error) {
	if err != nil {
		t.Errorf("%s due to %v", msg, err)
		t.FailNow()
	}
}

func TestDocker(t *testing.T) {
	dc, err := NewDockerClient("tcp://192.168.0.158:2375")
	if err != nil {
		t.Errorf("failed to create docker client: %v", err)
		t.FailNow()
	}

	t.Run("list images", func(t *testing.T) {
		images, err := dc.ListImages(context.Background(), nil)
		assertError(t, "failed to list images", err)
		t.Logf("images: %v", images)
		if len(images) == 0 {
			t.Errorf("no images found")
			t.FailNow()
		}
	})

	t.Run("list containers", func(t *testing.T) {
		containers, err := dc.ListContainers(context.Background(), nil)
		assertError(t, "failed to list containers", err)
		t.Logf("containers: %v", containers)
		if len(containers) == 0 {
			t.Errorf("no containers found")
			t.FailNow()
		}
	})

	t.Run("list networks", func(t *testing.T) {
		networks, err := dc.ListNetworks(context.Background(), nil)
		assertError(t, "failed to list networks", err)
		t.Logf("networks: %v", networks)
		if len(networks) == 0 {
			t.Errorf("no networks found")
			t.FailNow()
		}
	})
}
