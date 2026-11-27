//go:build integration

package controlplane

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const integrationAPIToken = "01234567890123456789012345678901"

func openIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func resetIntegrationDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE TABLE audit_events, outbox_events, job_events, job_leases, build_attempts, build_jobs, node_credentials, node_capabilities, build_nodes CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
}

func newIntegrationServer(t *testing.T, db *sql.DB) *Server {
	t.Helper()
	return New(db, integrationAPIToken, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func doRequest(t *testing.T, h http.Handler, method, target string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, target, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestIntegrationPostgresPing(t *testing.T) {
	db := openIntegrationDB(t)
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query("SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='build_jobs'")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("no result")
	}
	var n int
	if err := rows.Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("build_jobs table missing (count=%d)", n)
	}
}

func TestIntegrationNodeLifecycle(t *testing.T) {
	db := openIntegrationDB(t)
	resetIntegrationDB(t, db)
	s := newIntegrationServer(t, db)
	h := s.Handler()

	rr := doRequest(t, h, http.MethodPost, "/api/v1/nodes", map[string]any{
		"node_name":     "node-1",
		"agent_version": "0.1.0",
		"capabilities":  map[string]any{"linux-amd64": map[string]any{"arch": "amd64"}},
	}, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr.Code != http.StatusCreated {
		t.Fatalf("register node: %d %s", rr.Code, rr.Body.String())
	}
	var node struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &node); err != nil {
		t.Fatal(err)
	}
	if node.Status != "pending" {
		t.Fatalf("expected pending, got %s", node.Status)
	}

	rr = doRequest(t, h, http.MethodPost, "/api/v1/nodes/"+node.ID+"/heartbeat", map[string]any{"agent_version": "0.2.0"}, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr.Code != http.StatusOK {
		t.Fatalf("heartbeat: %d %s", rr.Code, rr.Body.String())
	}
	var hb struct {
		Status string `json:"status"`
	}
	json.Unmarshal(rr.Body.Bytes(), &hb)
	if hb.Status != "ready" {
		t.Fatalf("expected ready, got %s", hb.Status)
	}

	rr = doRequest(t, h, http.MethodPost, "/api/v1/nodes/"+node.ID+"/credentials", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr.Code != http.StatusCreated {
		t.Fatalf("issue credential: %d %s", rr.Code, rr.Body.String())
	}
	var cred struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &cred); err != nil {
		t.Fatal(err)
	}
	if cred.Token == "" {
		t.Fatal("empty credential")
	}

	rr = doRequest(t, h, http.MethodPost, "/agent/v1/heartbeat", map[string]any{"agent_version": "0.2.0"}, map[string]string{"Authorization": "Bearer " + cred.Token})
	if rr.Code != http.StatusOK {
		t.Fatalf("agent heartbeat: %d %s", rr.Code, rr.Body.String())
	}
}

func TestIntegrationJobDispatchAndAgentFlow(t *testing.T) {
	db := openIntegrationDB(t)
	resetIntegrationDB(t, db)
	s := newIntegrationServer(t, db)
	h := s.Handler()

	// register + ready node
	rr := doRequest(t, h, http.MethodPost, "/api/v1/nodes", map[string]any{
		"node_name": "node-1", "agent_version": "0.1.0",
		"capabilities": map[string]any{"linux-amd64": map[string]any{}},
	}, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr.Code != http.StatusCreated {
		t.Fatal(rr.Body.String())
	}
	var node struct {
		ID string `json:"id"`
	}
	json.Unmarshal(rr.Body.Bytes(), &node)
	doRequest(t, h, http.MethodPost, "/api/v1/nodes/"+node.ID+"/heartbeat", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})

	// issue credential
	rr = doRequest(t, h, http.MethodPost, "/api/v1/nodes/"+node.ID+"/credentials", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	var cred struct {
		Token string `json:"token"`
	}
	json.Unmarshal(rr.Body.Bytes(), &cred)

	// create job
	rr = doRequest(t, h, http.MethodPost, "/api/v1/jobs", map[string]any{
		"tenant_id": "tenant-a", "source_commit": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create job: %d %s", rr.Code, rr.Body.String())
	}
	var job struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	json.Unmarshal(rr.Body.Bytes(), &job)
	if job.Status != "queued" {
		t.Fatalf("expected queued, got %s", job.Status)
	}

	// dispatch
	rr = doRequest(t, h, http.MethodPost, "/api/v1/scheduler/dispatch", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr.Code != http.StatusOK {
		t.Fatalf("dispatch: %d %s", rr.Code, rr.Body.String())
	}
	var disp struct {
		Dispatched   bool   `json:"dispatched"`
		JobID        string `json:"job_id"`
		FencingToken int64  `json:"fencing_token"`
	}
	json.Unmarshal(rr.Body.Bytes(), &disp)
	if !disp.Dispatched {
		t.Fatalf("not dispatched: %s", rr.Body.String())
	}
	if disp.JobID != job.ID {
		t.Fatalf("job mismatch")
	}

	// agent next
	rr = doRequest(t, h, http.MethodGet, "/agent/v1/jobs/next", nil, map[string]string{"Authorization": "Bearer " + cred.Token})
	if rr.Code != http.StatusOK {
		t.Fatalf("next: %d %s", rr.Code, rr.Body.String())
	}
	var next struct {
		Job *struct {
			ID           string `json:"id"`
			FencingToken int64  `json:"fencing_token"`
			AttemptID    string `json:"attempt_id"`
		} `json:"job"`
	}
	json.Unmarshal(rr.Body.Bytes(), &next)
	if next.Job == nil || next.Job.ID != job.ID {
		t.Fatalf("no assigned job: %s", rr.Body.String())
	}
	ft := next.Job.FencingToken
	headers := map[string]string{"Authorization": "Bearer " + cred.Token, "X-Fencing-Token": itoa(ft)}

	// start
	rr = doRequest(t, h, http.MethodPost, "/agent/v1/jobs/"+job.ID+"/started", nil, headers)
	if rr.Code != http.StatusOK {
		t.Fatalf("start: %d %s", rr.Code, rr.Body.String())
	}

	// renew
	rr = doRequest(t, h, http.MethodPost, "/agent/v1/jobs/"+job.ID+"/lease", nil, headers)
	if rr.Code != http.StatusOK {
		t.Fatalf("renew: %d %s", rr.Code, rr.Body.String())
	}

	// complete
	rr = doRequest(t, h, http.MethodPost, "/agent/v1/jobs/"+job.ID+"/complete", map[string]any{"status": "succeeded"}, headers)
	if rr.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", rr.Code, rr.Body.String())
	}

	// verify job status
	rr = doRequest(t, h, http.MethodGet, "/api/v1/jobs/"+job.ID+"?tenant_id=tenant-a", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	var final job
	json.Unmarshal(rr.Body.Bytes(), &final)
	if final.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s", final.Status)
	}
}

func TestIntegrationDoubleClaim(t *testing.T) {
	db := openIntegrationDB(t)
	resetIntegrationDB(t, db)
	s := newIntegrationServer(t, db)
	h := s.Handler()

	rr := doRequest(t, h, http.MethodPost, "/api/v1/nodes", map[string]any{
		"node_name": "node-1", "agent_version": "0.1.0",
	}, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	var node struct {
		ID string `json:"id"`
	}
	json.Unmarshal(rr.Body.Bytes(), &node)
	doRequest(t, h, http.MethodPost, "/api/v1/nodes/"+node.ID+"/heartbeat", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})

	doRequest(t, h, http.MethodPost, "/api/v1/jobs", map[string]any{
		"tenant_id": "tenant-a", "source_commit": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}, map[string]string{"Authorization": "Bearer " + integrationAPIToken})

	rr = doRequest(t, h, http.MethodPost, "/api/v1/scheduler/dispatch", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body.String())
	}
	var first struct {
		Dispatched bool `json:"dispatched"`
	}
	json.Unmarshal(rr.Body.Bytes(), &first)
	if !first.Dispatched {
		t.Fatal("first dispatch failed")
	}

	rr = doRequest(t, h, http.MethodPost, "/api/v1/scheduler/dispatch", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	var second struct {
		Dispatched bool `json:"dispatched"`
	}
	json.Unmarshal(rr.Body.Bytes(), &second)
	if second.Dispatched {
		t.Fatal("second dispatch should not claim a job")
	}
}

func TestIntegrationExpiredLeaseReconciliation(t *testing.T) {
	db := openIntegrationDB(t)
	resetIntegrationDB(t, db)
	s := newIntegrationServer(t, db)
	h := s.Handler()

	rr := doRequest(t, h, http.MethodPost, "/api/v1/nodes", map[string]any{"node_name": "node-1"}, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	var node struct {
		ID string `json:"id"`
	}
	json.Unmarshal(rr.Body.Bytes(), &node)
	doRequest(t, h, http.MethodPost, "/api/v1/nodes/"+node.ID+"/heartbeat", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})

	doRequest(t, h, http.MethodPost, "/api/v1/jobs", map[string]any{"tenant_id": "tenant-a", "source_commit": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, map[string]string{"Authorization": "Bearer " + integrationAPIToken})

	rr = doRequest(t, h, http.MethodPost, "/api/v1/scheduler/dispatch", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	var disp struct {
		JobID string `json:"job_id"`
	}
	json.Unmarshal(rr.Body.Bytes(), &disp)

	// expire the lease
	if _, err := db.Exec(`UPDATE job_leases SET expires_at = now() - interval '1 second' WHERE job_id = $1`, disp.JobID); err != nil {
		t.Fatal(err)
	}

	n, err := s.ReconcileExpiredLeases(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 reconciled, got %d", n)
	}

	rr = doRequest(t, h, http.MethodGet, "/api/v1/jobs/"+disp.JobID+"?tenant_id=tenant-a", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	var job struct {
		Status string `json:"status"`
	}
	json.Unmarshal(rr.Body.Bytes(), &job)
	if job.Status != "queued" {
		t.Fatalf("expected queued after reconciliation, got %s", job.Status)
	}
}

func TestIntegrationTenantIsolation(t *testing.T) {
	db := openIntegrationDB(t)
	resetIntegrationDB(t, db)
	s := newIntegrationServer(t, db)
	h := s.Handler()

	doRequest(t, h, http.MethodPost, "/api/v1/jobs", map[string]any{"tenant_id": "tenant-a", "source_commit": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	// retrieve id via list
	rr := doRequest(t, h, http.MethodGet, "/api/v1/jobs?tenant_id=tenant-a", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	var resp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp.Items) != 1 {
		t.Fatal("expected one job")
	}
	jobID := resp.Items[0].ID

	rr = doRequest(t, h, http.MethodGet, "/api/v1/jobs/"+jobID+"?tenant_id=tenant-b", nil, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for wrong tenant, got %d", rr.Code)
	}
}

func TestIntegrationIdempotencyConflict(t *testing.T) {
	db := openIntegrationDB(t)
	resetIntegrationDB(t, db)
	s := newIntegrationServer(t, db)
	h := s.Handler()

	body := map[string]any{"tenant_id": "tenant-a", "source_commit": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "idempotency_key": "key-1"}
	rr := doRequest(t, h, http.MethodPost, "/api/v1/jobs", body, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr.Code != http.StatusCreated {
		t.Fatal(rr.Body.String())
	}
	// same payload returns same job
	rr2 := doRequest(t, h, http.MethodPost, "/api/v1/jobs", body, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 for same payload, got %d", rr2.Code)
	}
	// different payload conflicts
	body2 := map[string]any{"tenant_id": "tenant-a", "source_commit": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "idempotency_key": "key-1"}
	rr3 := doRequest(t, h, http.MethodPost, "/api/v1/jobs", body2, map[string]string{"Authorization": "Bearer " + integrationAPIToken})
	if rr3.Code != http.StatusConflict {
		t.Fatalf("expected 409 for conflicting payload, got %d", rr3.Code)
	}
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}
