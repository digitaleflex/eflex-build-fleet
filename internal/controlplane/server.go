package controlplane

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var errNoRows = sql.ErrNoRows

type Server struct { db *sql.DB; logger *slog.Logger; token string; metrics *metrics }

func New(db *sql.DB, token string, logger *slog.Logger) *Server { return &Server{db:db,token:token,logger:logger,metrics:newMetrics()} }

func (s *Server) Handler() http.Handler {
	mux:=http.NewServeMux()
	mux.HandleFunc("GET /healthz",func(w http.ResponseWriter,_ *http.Request){writeJSON(w,200,map[string]string{"status":"ok","service":"control-plane"})})
	mux.HandleFunc("GET /readyz",s.readiness)
	mux.HandleFunc("GET /metrics",s.metricsHandler)
	admin:=http.NewServeMux()
	admin.HandleFunc("POST /api/v1/nodes",s.registerNode)
	admin.HandleFunc("GET /api/v1/nodes",s.listNodes)
	admin.HandleFunc("GET /api/v1/nodes/{id}",s.getNode)
	admin.HandleFunc("POST /api/v1/nodes/{id}/credentials",s.issueNodeCredential)
	admin.HandleFunc("POST /api/v1/nodes/{id}/heartbeat",s.heartbeat)
	admin.HandleFunc("POST /api/v1/jobs",s.createJob)
	admin.HandleFunc("GET /api/v1/jobs",s.listJobs)
	admin.HandleFunc("GET /api/v1/jobs/{id}",s.getJob)
	admin.HandleFunc("POST /api/v1/scheduler/dispatch",s.dispatchNext)
	mux.Handle("/api/v1/",s.authenticate(admin))
	agent:=http.NewServeMux()
	agent.HandleFunc("POST /agent/v1/heartbeat",s.agentHeartbeat)
	agent.HandleFunc("GET /agent/v1/jobs/next",s.agentNextJob)
	agent.HandleFunc("POST /agent/v1/jobs/{id}/started",s.agentStartJob)
	agent.HandleFunc("POST /agent/v1/jobs/{id}/lease",s.agentRenewLease)
	agent.HandleFunc("POST /agent/v1/jobs/{id}/complete",s.agentCompleteJob)
	mux.Handle("/agent/v1/",s.authenticateNode(agent))
	return securityHeaders(mux)
}

func (s *Server) Run(ctx context.Context,address string) error {
	server:=&http.Server{Addr:address,Handler:s.Handler(),ReadHeaderTimeout:5*time.Second,IdleTimeout:60*time.Second}
	errCh:=make(chan error,1)
	go func(){s.logger.Info("control plane listening","address",address);errCh<-server.ListenAndServe()}()
	select {
	case <-ctx.Done():
		shutdownCtx,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel()
		if err:=server.Shutdown(shutdownCtx);err!=nil{return err}
		err:=<-errCh;if errors.Is(err,http.ErrServerClosed){return nil};return err
	case err:=<-errCh:
		if errors.Is(err,http.ErrServerClosed){return nil};return err
	}
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		header:=r.Header.Get("Authorization")
		if !strings.HasPrefix(header,"Bearer ") {writeError(w,401,"unauthorized");return}
		got:=strings.TrimSpace(strings.TrimPrefix(header,"Bearer "))
		if got==""||subtle.ConstantTimeCompare([]byte(got),[]byte(s.token))!=1 {writeError(w,401,"unauthorized");return}
		next.ServeHTTP(w,r)
	})
}
func securityHeaders(next http.Handler) http.Handler {return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Header().Set("X-Content-Type-Options","nosniff");w.Header().Set("Cache-Control","no-store");next.ServeHTTP(w,r)})}

type nodeInput struct { NodeName string `json:"node_name"`; AgentVersion string `json:"agent_version"`; Capabilities map[string]json.RawMessage `json:"capabilities"` }
type node struct { ID string `json:"id"`; NodeName string `json:"node_name"`; Status string `json:"status"`; AgentVersion *string `json:"agent_version"`; LastHeartbeatAt *time.Time `json:"last_heartbeat_at"`; CreatedAt time.Time `json:"created_at"`; UpdatedAt time.Time `json:"updated_at"` }
type heartbeatInput struct { AgentVersion string `json:"agent_version"` }

func (s *Server) registerNode(w http.ResponseWriter,r *http.Request) {
	var in nodeInput
	if err:=decodeJSON(w,r,&in);err!=nil{writeError(w,400,err.Error());return}
	in.NodeName=strings.TrimSpace(in.NodeName)
	if in.NodeName==""||len(in.NodeName)>128{writeError(w,400,"node_name is required and must be at most 128 characters");return}
	if len(in.AgentVersion)>64{writeError(w,400,"agent_version must be at most 64 characters");return}
	tx,err:=s.db.BeginTx(r.Context(),nil);if err!=nil{writeError(w,500,"database error");return};defer tx.Rollback()
	var out node
	err=tx.QueryRowContext(r.Context(),`INSERT INTO build_nodes(node_name,status,agent_version) VALUES($1,'pending',NULLIF($2,'')) RETURNING id::text,node_name,status,agent_version,last_heartbeat_at,created_at,updated_at`,in.NodeName,in.AgentVersion).Scan(&out.ID,&out.NodeName,&out.Status,&out.AgentVersion,&out.LastHeartbeatAt,&out.CreatedAt,&out.UpdatedAt)
	if err!=nil{if strings.Contains(strings.ToLower(err.Error()),"unique"){writeError(w,409,"node_name already exists")}else{writeError(w,500,"could not register node")};return}
	for name,value:=range in.Capabilities{name=strings.TrimSpace(name);if name==""||len(name)>128{writeError(w,400,"invalid capability name");return};if !json.Valid(value){writeError(w,400,"invalid capability JSON");return};if _,err=tx.ExecContext(r.Context(),`INSERT INTO node_capabilities(node_id,capability,value) VALUES($1,$2,$3::jsonb)`,out.ID,name,string(value));err!=nil{writeError(w,400,"invalid capability value");return}}
	if err=tx.Commit();err!=nil{writeError(w,500,"could not commit node registration");return};writeJSON(w,201,out)
}
func (s *Server) listNodes(w http.ResponseWriter,r *http.Request){rows,err:=s.db.QueryContext(r.Context(),`SELECT id::text,node_name,status,agent_version,last_heartbeat_at,created_at,updated_at FROM build_nodes ORDER BY created_at DESC LIMIT 200`);if err!=nil{writeError(w,500,"could not list nodes");return};defer rows.Close();items:=make([]node,0);for rows.Next(){var n node;if err:=rows.Scan(&n.ID,&n.NodeName,&n.Status,&n.AgentVersion,&n.LastHeartbeatAt,&n.CreatedAt,&n.UpdatedAt);err!=nil{writeError(w,500,"could not read nodes");return};items=append(items,n)};if err:=rows.Err();err!=nil{writeError(w,500,"could not read nodes");return};writeJSON(w,200,map[string]any{"items":items,"count":len(items)})}
func (s *Server) getNode(w http.ResponseWriter,r *http.Request){id:=r.PathValue("id");if _,err:=uuid.Parse(id);err!=nil{writeError(w,400,"invalid node id");return};var n node;err:=s.db.QueryRowContext(r.Context(),`SELECT id::text,node_name,status,agent_version,last_heartbeat_at,created_at,updated_at FROM build_nodes WHERE id=$1`,id).Scan(&n.ID,&n.NodeName,&n.Status,&n.AgentVersion,&n.LastHeartbeatAt,&n.CreatedAt,&n.UpdatedAt);if errors.Is(err,sql.ErrNoRows){writeError(w,404,"node not found");return};if err!=nil{writeError(w,500,"could not read node");return};writeJSON(w,200,n)}
func (s *Server) heartbeat(w http.ResponseWriter,r *http.Request){id:=r.PathValue("id");if _,err:=uuid.Parse(id);err!=nil{writeError(w,400,"invalid node id");return};var in heartbeatInput;if r.Body!=nil&&r.ContentLength!=0{if err:=decodeJSON(w,r,&in);err!=nil{writeError(w,400,err.Error());return}};var n node;err:=s.db.QueryRowContext(r.Context(),`UPDATE build_nodes SET status='ready',last_heartbeat_at=now(),agent_version=COALESCE(NULLIF($2,''),agent_version),updated_at=now() WHERE id=$1 RETURNING id::text,node_name,status,agent_version,last_heartbeat_at,created_at,updated_at`,id,in.AgentVersion).Scan(&n.ID,&n.NodeName,&n.Status,&n.AgentVersion,&n.LastHeartbeatAt,&n.CreatedAt,&n.UpdatedAt);if errors.Is(err,sql.ErrNoRows){writeError(w,404,"node not found");return};if err!=nil{writeError(w,500,"could not update heartbeat");return};writeJSON(w,200,n)}
func (s *Server) agentHeartbeat(w http.ResponseWriter,r *http.Request){id:=nodeIDFromContext(r.Context());var in heartbeatInput;if r.Body!=nil&&r.ContentLength!=0{if err:=decodeJSON(w,r,&in);err!=nil{writeError(w,400,err.Error());return}};_,err:=s.db.ExecContext(r.Context(),`UPDATE build_nodes SET status='ready',last_heartbeat_at=now(),agent_version=COALESCE(NULLIF($2,''),agent_version),updated_at=now() WHERE id=$1`,id,in.AgentVersion);if err!=nil{writeError(w,500,"could not update heartbeat");return};writeJSON(w,200,map[string]string{"status":"accepted","node_id":id})}
func (s *Server) readiness(w http.ResponseWriter,r *http.Request){ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel();if err:=s.db.PingContext(ctx);err!=nil{writeJSON(w,503,map[string]string{"status":"not_ready","database":"unavailable"});return};writeJSON(w,200,map[string]string{"status":"ready","database":"connected"})}
func decodeJSON(w http.ResponseWriter,r *http.Request,dst any)error{r.Body=http.MaxBytesReader(w,r.Body,1<<20);dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields();if err:=dec.Decode(dst);err!=nil{return err};var extra any;if err:=dec.Decode(&extra);!errors.Is(err,io.EOF){if err==nil{return errors.New("request must contain a single JSON value")};return err};return nil}
func writeJSON(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(v)}
func writeError(w http.ResponseWriter,status int,message string){writeJSON(w,status,map[string]string{"error":message})}
