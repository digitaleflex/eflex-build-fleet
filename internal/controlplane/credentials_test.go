package controlplane

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"log/slog"
	"io"
)

func TestAgentAPIRequiresNodeCredential(t *testing.T) {
	s:=New(nil,"01234567890123456789012345678901",slog.New(slog.NewTextHandler(io.Discard,nil)))
	req:=httptest.NewRequest(http.MethodPost,"/agent/v1/heartbeat",nil)
	rec:=httptest.NewRecorder()
	s.Handler().ServeHTTP(rec,req)
	if rec.Code!=http.StatusUnauthorized { t.Fatalf("got %d, want %d",rec.Code,http.StatusUnauthorized) }
}
