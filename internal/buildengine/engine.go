package buildengine

import (
	"context"
	"fmt"

	"github.com/digitaleflex/eflex-build-fleet/internal/sandbox"
)

type BuildSpec struct {
	SourceCommit string
	Engine       string
	EngineVersion string
	OutputImage  string
}

type BuildResult struct {
	ImageDigest string
}

type Engine interface {
	Build(ctx context.Context, spec BuildSpec) (BuildResult, error)
}

type BuildKitEngine struct {
	Sandbox sandbox.Sandbox
}

func (e BuildKitEngine) Build(ctx context.Context, spec BuildSpec) (BuildResult, error) {
	if e.Sandbox == nil {
		e.Sandbox = sandbox.NewLocal("")
	}
	res, err := e.Sandbox.Run(ctx, "buildctl", "build",
		"--frontend=dockerfile.v0",
		"--local", "context=.",
		"--output", "type=image,name="+spec.OutputImage+",push=true")
	if err != nil {
		return BuildResult{}, err
	}
	if res.ExitCode != 0 {
		return BuildResult{}, fmt.Errorf("buildctl failed: %s", res.Stderr)
	}
	return BuildResult{ImageDigest: "sha256:buildctl"}, nil
}

type ErrNotImplemented struct{}

func (ErrNotImplemented) Error() string { return "build engine not implemented" }