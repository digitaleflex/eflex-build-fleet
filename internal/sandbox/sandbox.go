package sandbox

import (
	"context"

	"github.com/digitaleflex/eflex-build-fleet/internal/buildrunner"
)

type Sandbox interface {
	Run(ctx context.Context, name string, args ...string) (buildrunner.Result, error)
}

type Local struct {
	Runner *buildrunner.Runner
}

func NewLocal(workRoot string) *Local {
	return &Local{Runner: buildrunner.New(workRoot, 0)}
}

func (l *Local) Run(ctx context.Context, name string, args ...string) (buildrunner.Result, error) {
	return l.Runner.Run(ctx, name, args...)
}