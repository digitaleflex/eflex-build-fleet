package sandbox

import (
	"fmt"
	"strings"
)

type Policy struct {
	AllowedCommands map[string]bool
}

func DefaultPolicy() Policy {
	return Policy{AllowedCommands: map[string]bool{
		"buildctl": true,
		"skopeo":   true,
		"oras":     true,
	}}
}

func (p Policy) Check(command string) error {
	if p.AllowedCommands == nil {
		return nil
	}
	base := command
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	if !p.AllowedCommands[base] {
		return fmt.Errorf("command %q is not allowed by sandbox policy", base)
	}
	return nil
}