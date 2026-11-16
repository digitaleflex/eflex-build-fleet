package buildrunner

import (
	"context"
	"testing"
	"time"
)

func TestRunEcho(t *testing.T) {
	r := New("", 30*time.Second)
	res, err := r.Run(context.Background(), "cmd", "/c", "echo", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit code %d, stderr=%s", res.ExitCode, res.Stderr)
	}
	if res.Stdout != "hello\r\n" {
		t.Fatalf("stdout = %q", res.Stdout)
	}
}