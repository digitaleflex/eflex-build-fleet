package buildengine

import "context"

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

type BuildKitEngine struct{}

func (BuildKitEngine) Build(ctx context.Context, spec BuildSpec) (BuildResult, error) {
	return BuildResult{}, ErrNotImplemented
}

type ErrNotImplemented struct{}

func (ErrNotImplemented) Error() string { return "build engine not implemented" }