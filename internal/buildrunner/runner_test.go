package buildrunner

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// Le test tournait seulement sous Windows (`cmd /c`, fin de ligne `\r\n`) : la CI Linux échouait.
func TestRunEcho(t *testing.T) {
	r := New("", 30*time.Second)
	name, args, want := "echo", []string{"hello"}, "hello\n"
	if runtime.GOOS == "windows" {
		name, args, want = "cmd", []string{"/c", "echo", "hello"}, "hello\r\n"
	}
	res, err := r.Run(context.Background(), name, args...)
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit code %d, stderr=%s", res.ExitCode, res.Stderr)
	}
	if res.Stdout != want {
		t.Fatalf("stdout = %q, want %q", res.Stdout, want)
	}
}
