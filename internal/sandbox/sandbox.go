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
	Policy Policy
}

func NewLocal(workRoot string) *Local {
	return &Local{Runner: buildrunner.New(workRoot, 0), Policy: DefaultPolicy()}
}

func (l *Local) Run(ctx context.Context, name string, args ...string) (buildrunner.Result, error) {
	if err := l.Policy.Check(name); err != nil {
		return buildrunner.Result{}, err
	}
	return l.Runner.Run(ctx, name, args...)
}