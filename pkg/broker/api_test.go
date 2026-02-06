package broker

import (
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
	"testing"
)

// fakeSpawner implements Spawner for compile-time contract testing.
type fakeSpawner struct{}

func (fakeSpawner) Spawn(id, command string, args ...string) (*SpawnResult, error) { return nil, nil }
func (fakeSpawner) SpawnRPC(id, command string, args ...string) (*SpawnResult, error) {
	return nil, nil
}
func (fakeSpawner) SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*SpawnResult, error) {
	return nil, nil
}
func (fakeSpawner) SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*SpawnResult, error) {
	return nil, nil
}

func TestSpawnerInterfaceCompilation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSpawnerInterfaceCompilation", nil, func(t *testing.T, tx *gorm.DB) {
		var _ Spawner = fakeSpawner{}
	})

}

func TestSpawnResultFields(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSpawnResultFields", nil, func(t *testing.T, tx *gorm.DB) {
		res := SpawnResult{ID: "pid", PID: 123, Command: "echo", Args: []string{"hi"}, Status: "running", ExitCode: 0}
		if res.ID != "pid" || res.PID != 123 || res.Command != "echo" || res.Status != "running" || res.ExitCode != 0 {
			t.Fatalf("unexpected spawn result: %+v", res)
		}
		if len(res.Args) != 1 || res.Args[0] != "hi" {
			t.Fatalf("unexpected args: %+v", res.Args)
		}
	})

}
