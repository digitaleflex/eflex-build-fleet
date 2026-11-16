package controlplane

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var commitPattern = regexp.MustCompile(`^(?:[a-fA-F0-9]{40}|[a-fA-F0-9]{64})$`)

type createJobInput struct {
	TenantID string `json:"tenant_id"`
	IdempotencyKey string `json:"idempotency_key"`
	SourceCommit string `json:"source_commit"`
	Engine string `json:"engine"`
	EngineVersion string `json:"engine_version"`
	RequestedResources json.RawMessage `json:"requested_resources"`
}
type job struct {
	ID string `json:"id"`
	TenantID string `json:"tenant_id"`
	Status string `json:"status"`
	SourceCommit string `json:"source_commit"`
	Engine *string `json:"engine"`
	EngineVersion *string `json:"engine_version"`
	RequestedResources json.RawMessage `json:"requested_resources"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type dispatchResult struct {
	JobID string `json:"job_id"`
	AttemptID string `json:"attempt_id"`
	NodeID string `json:"node_id"`
	FencingToken int64 `json:"fencing_token"`
	LeaseExpiresAt time.Time `json:"lease_expires_at"`
}

func validateCreateJob(in *createJobInput) string {
	in.TenantID=strings.TrimSpace(in.TenantID)
	in.SourceCommit=strings.ToLower(strings.TrimSpace(in.SourceCommit))
	in.Engine=strings.TrimSpace(in.Engine)
	in.EngineVersion=strings.TrimSpace(in.EngineVersion)
	in.IdempotencyKey=strings.TrimSpace(in.IdempotencyKey)
	if in.TenantID==""||len(in.TenantID)>128{return "tenant_id is required and must be at most 128 characters"}
	if !commitPattern.MatchString(in.SourceCommit){return "source_commit must be a 40- or 64-character hexadecimal commit"}
	if len(in.IdempotencyKey)>200{return "idempotency_key must be at most 200 characters"}
	if len(in.Engine)>64||len(in.EngineVersion)>64{return "engine and engine_version must be at most 64 characters"}
	if len(in.RequestedResources)==0{in.RequestedResources=json.RawMessage(`{}`)}
	if !json.Valid(in.RequestedResources)||in.RequestedResources[0]!='{' {return "requested_resources must be a JSON object"}
	return ""
}

func (s *Server) createJob(w http.ResponseWriter,r *http.Request) {
	var in createJobInput
	if err:=decodeJSON(w,r,&in);err!=nil{writeError(w,400,err.Error());return}
	if msg:=validateCreateJob(&in);msg!=""{writeError(w,400,msg);return}
	var out job
	err:=s.db.QueryRowContext(r.Context(),`INSERT INTO build_jobs(tenant_id,idempotency_key,status,source_commit,engine,engine_version,requested_resources)
		VALUES($1,NULLIF($2,''),'queued',$3,NULLIF($4,''),NULLIF($5,''),$6::jsonb)
		ON CONFLICT (tenant_id,idempotency_key) DO NOTHING
		RETURNING id::text,tenant_id,status,source_commit,engine,engine_version,requested_resources,created_at,updated_at`,
		in.TenantID,in.IdempotencyKey,in.SourceCommit,in.Engine,in.EngineVersion,string(in.RequestedResources)).
		Scan(&out.ID,&out.TenantID,&out.Status,&out.SourceCommit,&out.Engine,&out.EngineVersion,&out.RequestedResources,&out.CreatedAt,&out.UpdatedAt)
	if errors.Is(err,sql.ErrNoRows)&&in.IdempotencyKey!=""{
		err=s.db.QueryRowContext(r.Context(),`SELECT id::text,tenant_id,status,source_commit,engine,engine_version,requested_resources,created_at,updated_at FROM build_jobs WHERE tenant_id=$1 AND idempotency_key=$2`,in.TenantID,in.IdempotencyKey).
			Scan(&out.ID,&out.TenantID,&out.Status,&out.SourceCommit,&out.Engine,&out.EngineVersion,&out.RequestedResources,&out.CreatedAt,&out.UpdatedAt)
		if err==nil{writeJSON(w,200,out);return}
	}
	if err!=nil{writeError(w,500,"could not create build job");return}
	_,_=s.db.ExecContext(r.Context(),`INSERT INTO job_events(job_id,event_type,payload) VALUES($1,'job.queued',jsonb_build_object('tenant_id',$2))`,out.ID,out.TenantID)
	writeJSON(w,201,out)
}

func (s *Server) listJobs(w http.ResponseWriter,r *http.Request) {
	tenant:=strings.TrimSpace(r.URL.Query().Get("tenant_id"))
	if tenant==""{writeError(w,400,"tenant_id query parameter is required");return}
	rows,err:=s.db.QueryContext(r.Context(),`SELECT id::text,tenant_id,status,source_commit,engine,engine_version,requested_resources,created_at,updated_at FROM build_jobs WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT 100`,tenant)
	if err!=nil{writeError(w,500,"could not list jobs");return};defer rows.Close()
	items:=make([]job,0)
	for rows.Next(){var j job;if err:=rows.Scan(&j.ID,&j.TenantID,&j.Status,&j.SourceCommit,&j.Engine,&j.EngineVersion,&j.RequestedResources,&j.CreatedAt,&j.UpdatedAt);err!=nil{writeError(w,500,"could not read jobs");return};items=append(items,j)}
	if err:=rows.Err();err!=nil{writeError(w,500,"could not read jobs");return}
	writeJSON(w,200,map[string]any{"items":items,"count":len(items)})
}

func (s *Server) getJob(w http.ResponseWriter,r *http.Request) {
	id:=r.PathValue("id");if _,err:=uuid.Parse(id);err!=nil{writeError(w,400,"invalid job id");return}
	var j job
	err:=s.db.QueryRowContext(r.Context(),`SELECT id::text,tenant_id,status,source_commit,engine,engine_version,requested_resources,created_at,updated_at FROM build_jobs WHERE id=$1`,id).
		Scan(&j.ID,&j.TenantID,&j.Status,&j.SourceCommit,&j.Engine,&j.EngineVersion,&j.RequestedResources,&j.CreatedAt,&j.UpdatedAt)
	if errors.Is(err,sql.ErrNoRows){writeError(w,404,"job not found");return};if err!=nil{writeError(w,500,"could not read job");return}
	writeJSON(w,200,j)
}

// dispatchNext atomically claims one queued job and one healthy node. SKIP LOCKED
// allows multiple scheduler replicas without double assignment.
func (s *Server) dispatchNext(w http.ResponseWriter,r *http.Request) {
	tx,err:=s.db.BeginTx(r.Context(),nil);if err!=nil{writeError(w,500,"could not start dispatch transaction");return};defer tx.Rollback()
	var jobID string
	err=tx.QueryRowContext(r.Context(),`SELECT id::text FROM build_jobs WHERE status='queued' ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&jobID)
	if errors.Is(err,sql.ErrNoRows){writeJSON(w,200,map[string]any{"dispatched":false});return};if err!=nil{writeError(w,500,"could not select queued job");return}
	var nodeID string
	err=tx.QueryRowContext(r.Context(),`SELECT id::text FROM build_nodes WHERE status='ready' AND last_heartbeat_at>now()-interval '90 seconds' ORDER BY last_heartbeat_at DESC LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&nodeID)
	if errors.Is(err,sql.ErrNoRows){writeJSON(w,200,map[string]any{"dispatched":false,"reason":"no_healthy_build_node"});return};if err!=nil{writeError(w,500,"could not select build node");return}
	var attemptNo int
	if err=tx.QueryRowContext(r.Context(),`SELECT COALESCE(MAX(attempt_number),0)+1 FROM build_attempts WHERE job_id=$1`,jobID).Scan(&attemptNo);err!=nil{writeError(w,500,"could not calculate attempt number");return}
	var fencing int64
	if err=tx.QueryRowContext(r.Context(),`SELECT COALESCE(MAX(fencing_token),0)+1 FROM job_leases WHERE job_id=$1`,jobID).Scan(&fencing);err!=nil{writeError(w,500,"could not allocate fencing token");return}
	var attemptID string
	if err=tx.QueryRowContext(r.Context(),`INSERT INTO build_attempts(job_id,node_id,attempt_number,status,started_at) VALUES($1,$2,$3,'assigned',now()) RETURNING id::text`,jobID,nodeID,attemptNo).Scan(&attemptID);err!=nil{writeError(w,500,"could not create build attempt");return}
	var expires time.Time
	if err=tx.QueryRowContext(r.Context(),`INSERT INTO job_leases(job_id,attempt_id,node_id,fencing_token,expires_at) VALUES($1,$2,$3,$4,now()+interval '60 seconds') RETURNING expires_at`,jobID,attemptID,nodeID,fencing).Scan(&expires);err!=nil{writeError(w,500,"could not create job lease");return}
	if _,err=tx.ExecContext(r.Context(),`UPDATE build_jobs SET status='scheduled',updated_at=now() WHERE id=$1`,jobID);err!=nil{writeError(w,500,"could not update job");return}
	if _,err=tx.ExecContext(r.Context(),`INSERT INTO job_events(job_id,event_type,payload) VALUES($1,'job.scheduled',jsonb_build_object('node_id',$2,'attempt_id',$3,'fencing_token',$4))`,jobID,nodeID,attemptID,fencing);err!=nil{writeError(w,500,"could not record job event");return}
	if _,err=tx.ExecContext(r.Context(),`INSERT INTO outbox_events(aggregate_type,aggregate_id,event_type,payload) VALUES('build_job',$1,'job.scheduled',jsonb_build_object('node_id',$2,'attempt_id',$3,'fencing_token',$4))`,jobID,nodeID,attemptID,fencing);err!=nil{writeError(w,500,"could not record outbox event");return}
	if err=tx.Commit();err!=nil{writeError(w,500,"could not commit dispatch");return}
	writeJSON(w,200,dispatchResult{JobID:jobID,AttemptID:attemptID,NodeID:nodeID,FencingToken:fencing,LeaseExpiresAt:expires})
}
