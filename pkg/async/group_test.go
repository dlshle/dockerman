package async

import (
	"context"
	"testing"
	"time"
)

func TestRace(t *testing.T) {
	g := NewAsyncGroup[int]()
	t.Run("test", func(t *testing.T) {
		g.Run(func(ctx context.Context) int {
			time.Sleep(50 * time.Millisecond)
			t.Logf("%d", 1)
			return 1
		})
		g.Run(func(ctx context.Context) int {
			time.Sleep(20 * time.Millisecond)
			t.Logf("%d", 2)
			return 2
		})
		g.Run(func(ctx context.Context) int {
			time.Sleep(30 * time.Millisecond)
			t.Logf("%d", 3)
			return 3
		})
		if g.Race() != 2 {
			t.Fail()
		}
	})
}
