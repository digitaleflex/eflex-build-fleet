package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type agentConfig struct {
	ControlPlaneURL string
	AdminToken      string
	NodeName       string
	AgentVersion   string
	Credential     string
	HeartbeatEvery time.Duration
	PollEvery      time.Duration
}

func loadAgentConfig() (agentConfig, error) {
	cp := getenv("CONTROL_PLANE_URL", "http://localhost:8080")
	admin := os.Getenv("FLEET_API_TOKEN")
	name := getenv("NODE_NAME", "agent-1")
	ver := getenv("AGENT_VERSION", "0.1.0")
	cred := os.Getenv("NODE_CREDENTIAL")
	hb := parseDur(getenv("HEARTBEAT_INTERVAL", "10s"), 10*time.Second)
	poll := parseDur(getenv("POLL_INTERVAL", "2s"), 2*time.Second)
	if cp == "" {
		return agentConfig{}, fmt.Errorf("CONTROL_PLANE_URL is required")
	}
	return agentConfig{ControlPlaneURL: cp, AdminToken: admin, NodeName: name, AgentVersion: ver, Credential: cred, HeartbeatEvery: hb, PollEvery: poll}, nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func parseDur(s string, d time.Duration) time.Duration {
	if dd, err := time.ParseDuration(s); err == nil {
		return dd
	}
	return d
}

func doJSON(method, url, token string, body any, headers map[string]string) (int, []byte, error) {
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return 0, nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data, nil
}

type assignedJob struct {
	ID                 string          `json:"id"`
	TenantID           string          `json:"tenant_id"`
	SourceCommit       string          `json:"source_commit"`
	Engine             *string         `json:"engine"`
	EngineVersion      *string         `json:"engine_version"`
	RequestedResources json.RawMessage `json:"requested_resources"`
	AttemptID          string          `json:"attempt_id"`
	FencingToken       int64           `json:"fencing_token"`
	LeaseExpiresAt     time.Time       `json:"lease_expires_at"`
}

func enroll(cfg *agentConfig, logger *slog.Logger) error {
	if cfg.Credential != "" {
		return nil
	}
	if cfg.AdminToken == "" {
		return fmt.Errorf("NODE_CREDENTIAL or FLEET_API_TOKEN (admin) is required for enrollment")
	}
	status, data, err := doJSON("POST", cfg.ControlPlaneURL+"/api/v1/nodes", cfg.AdminToken, map[string]any{
		"node_name": cfg.NodeName, "agent_version": cfg.AgentVersion,
	}, nil)
	if err != nil || status != 201 {
		return fmt.Errorf("register node failed: %d %s %v", status, string(data), err)
	}
	var node struct {
		ID string `json:"id"`
	}
	json.Unmarshal(data, &node)
	status, data, err = doJSON("POST", cfg.ControlPlaneURL+"/api/v1/nodes/"+node.ID+"/credentials", cfg.AdminToken, nil, nil)
	if err != nil || status != 201 {
		return fmt.Errorf("issue credential failed: %d %s %v", status, string(data), err)
	}
	var cred struct {
		Token string `json:"token"`
	}
	json.Unmarshal(data, &cred)
	cfg.Credential = cred.Token
	logger.Info("enrolled node", "node_id", node.ID, "credential", cred.Token)
	fmt.Println("NODE_CREDENTIAL=" + cred.Token)
	return nil
}

func runAgent(ctx context.Context, cfg agentConfig, logger *slog.Logger) {
	cred := cfg.Credential
	// heartbeat loop
	hbTicker := time.NewTicker(cfg.HeartbeatEvery)
	defer hbTicker.Stop()
	pollTicker := time.NewTicker(cfg.PollEvery)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hbTicker.C:
			status, data, err := doJSON("POST", cfg.ControlPlaneURL+"/agent/v1/heartbeat", cred, map[string]any{"agent_version": cfg.AgentVersion}, nil)
			if err != nil || status != 200 {
				logger.Warn("heartbeat failed", "status", status, "err", err, "body", string(data))
			}
		case <-pollTicker.C:
			status, data, err := doJSON("GET", cfg.ControlPlaneURL+"/agent/v1/jobs/next", cred, nil, nil)
			if err != nil || status != 200 {
				logger.Warn("poll failed", "status", status, "err", err)
				continue
			}
			var resp struct {
				Job *assignedJob `json:"job"`
			}
			json.Unmarshal(data, &resp)
			if resp.Job == nil {
				continue
			}
			job := resp.Job
			logger.Info("received job", "job_id", job.ID, "fencing", job.FencingToken)
			ft := map[string]string{"X-Fencing-Token": fmt.Sprintf("%d", job.FencingToken)}
			// start
			s, d, e := doJSON("POST", cfg.ControlPlaneURL+"/agent/v1/jobs/"+job.ID+"/started", cred, nil, ft)
			if e != nil || s != 200 {
				logger.Warn("start failed", "status", s, "err", e, "body", string(d))
				continue
			}
			// simulate work
			time.Sleep(500 * time.Millisecond)
			// renew lease
			s, d, e = doJSON("POST", cfg.ControlPlaneURL+"/agent/v1/jobs/"+job.ID+"/lease", cred, nil, ft)
			if e != nil || s != 200 {
				logger.Warn("renew failed", "status", s, "err", e, "body", string(d))
			}
			// complete
			s, d, e = doJSON("POST", cfg.ControlPlaneURL+"/agent/v1/jobs/"+job.ID+"/complete", cred, map[string]any{"status": "succeeded", "error_code": ""}, ft)
			if e != nil || s != 200 {
				logger.Warn("complete failed", "status", s, "err", e, "body", string(d))
			} else {
				logger.Info("job completed", "job_id", job.ID)
			}
		}
	}
}

func main() {
	cfg, err := loadAgentConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := enroll(&cfg, logger); err != nil {
		logger.Error("enrollment failed", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	runAgent(ctx, cfg, logger)
}
