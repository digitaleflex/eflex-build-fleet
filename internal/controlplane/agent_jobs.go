package controlplane

import (
 "database/sql"
 "encoding/json"
 "errors"
 "net/http"
 "time"
 "github.com/google/uuid"
)

type assignedJob struct {
 ID string `json:"id"`
 TenantID string `json:"tenant_id"`
 SourceCommit string `json:"source_commit"`
 Engine *string `json:"engine"`
 EngineVersion *string `json:"engine_version"`
 RequestedResources json.RawMessage `json:"requested_resources"`
 AttemptID string `json:"attempt_id"`
 FencingToken int64 `json:"fencing_token"`
 LeaseExpiresAt time.Time `json:"lease_expires_at"`
}
type completeJobInput struct { Status string `json:"status"`; ErrorCode string `json:"error_code"` }

func (s *Server) agentNextJob(w http.ResponseWriter,r *http.Request) {
 nodeID:=nodeIDFromContext(r.Context())
 var out assignedJob
 err:=s.db.QueryRowContext(r.Context(),`SELECT j.id::text,j.tenant_id,j.source_commit,j.engine,j.engine_version,j.requested_resources,a.id::text,l.fencing_token,l.expires_at
 FROM job_leases l JOIN build_attempts a ON a.id=l.attempt_id JOIN build_jobs j ON j.id=l.job_id
 WHERE l.node_id=$1 AND l.expires_at>now() AND a.status='assigned' AND j.status='scheduled'
 ORDER BY l.created_at LIMIT 1`,nodeID).Scan(&out.ID,&out.TenantID,&out.SourceCommit,&out.Engine,&out.EngineVersion,&out.RequestedResources,&out.AttemptID,&out.FencingToken,&out.LeaseExpiresAt)
 if errors.Is(err,sql.ErrNoRows){writeJSON(w,200,map[string]any{"job":nil});return}
 if err!=nil{writeError(w,500,"could not fetch assigned job");return}
 writeJSON(w,200,map[string]any{"job":out})
}
func (s *Server) agentStartJob(w http.ResponseWriter,r *http.Request) {
 nodeID:=nodeIDFromContext(r.Context());jobID:=r.PathValue("id")
 if _,err:=uuid.Parse(jobID);err!=nil{writeError(w,400,"invalid job id");return}
 tx,err:=s.db.BeginTx(r.Context(),nil);if err!=nil{writeError(w,500,"could not start transaction");return};defer tx.Rollback()
 var attemptID string
 err=tx.QueryRowContext(r.Context(),`UPDATE build_attempts a SET status='running',started_at=COALESCE(started_at,now())
 FROM job_leases l WHERE a.id=l.attempt_id AND l.job_id=$1 AND l.node_id=$2 AND l.expires_at>now() AND a.status='assigned'
 RETURNING a.id::text`,jobID,nodeID).Scan(&attemptID)
 if errors.Is(err,sql.ErrNoRows){writeError(w,409,"no active assigned lease for this node");return};if err!=nil{writeError(w,500,"could not start attempt");return}
 if _,err=tx.ExecContext(r.Context(),`UPDATE build_jobs SET status='running',updated_at=now() WHERE id=$1 AND status='scheduled'`,jobID);err!=nil{writeError(w,500,"could not update job");return}
 _,_=tx.ExecContext(r.Context(),`INSERT INTO job_events(job_id,event_type,payload) VALUES($1,'job.started',jsonb_build_object('node_id',$2,'attempt_id',$3))`,jobID,nodeID,attemptID)
 if err=tx.Commit();err!=nil{writeError(w,500,"could not commit start");return}
 writeJSON(w,200,map[string]string{"job_id":jobID,"status":"running"})
}
func (s *Server) agentRenewLease(w http.ResponseWriter,r *http.Request) {
 nodeID:=nodeIDFromContext(r.Context());jobID:=r.PathValue("id")
 if _,err:=uuid.Parse(jobID);err!=nil{writeError(w,400,"invalid job id");return}
 var expires time.Time
 err:=s.db.QueryRowContext(r.Context(),`UPDATE job_leases SET expires_at=now()+interval '60 seconds'
 WHERE job_id=$1 AND node_id=$2 AND expires_at>now() AND attempt_id IN (SELECT id FROM build_attempts WHERE status='running')
 RETURNING expires_at`,jobID,nodeID).Scan(&expires)
 if errors.Is(err,sql.ErrNoRows){writeError(w,409,"active lease not found");return};if err!=nil{writeError(w,500,"could not renew lease");return}
 writeJSON(w,200,map[string]time.Time{"expires_at":expires})
}
func (s *Server) agentCompleteJob(w http.ResponseWriter,r *http.Request) {
 nodeID:=nodeIDFromContext(r.Context());jobID:=r.PathValue("id")
 if _,err:=uuid.Parse(jobID);err!=nil{writeError(w,400,"invalid job id");return}
 var in completeJobInput
 if err:=decodeJSON(w,r,&in);err!=nil{writeError(w,400,err.Error());return}
 if in.Status!="succeeded"&&in.Status!="failed"{writeError(w,400,"status must be succeeded or failed");return}
 if len(in.ErrorCode)>128{writeError(w,400,"error_code is too long");return}
 tx,err:=s.db.BeginTx(r.Context(),nil);if err!=nil{writeError(w,500,"could not start completion transaction");return};defer tx.Rollback()
 var attemptID string
 err=tx.QueryRowContext(r.Context(),`SELECT a.id::text FROM build_attempts a JOIN job_leases l ON l.attempt_id=a.id
 WHERE l.job_id=$1 AND l.node_id=$2 AND l.expires_at>now() AND a.status='running' FOR UPDATE OF a`,jobID,nodeID).Scan(&attemptID)
 if errors.Is(err,sql.ErrNoRows){writeError(w,409,"active running lease not found");return};if err!=nil{writeError(w,500,"could not verify lease");return}
 if _,err=tx.ExecContext(r.Context(),`UPDATE build_attempts SET status=$2,finished_at=now(),error_code=NULLIF($3,'') WHERE id=$1`,attemptID,in.Status,in.ErrorCode);err!=nil{writeError(w,500,"could not finish attempt");return}
 if _,err=tx.ExecContext(r.Context(),`UPDATE build_jobs SET status=$2,updated_at=now() WHERE id=$1`,jobID,in.Status);err!=nil{writeError(w,500,"could not finish job");return}
 if _,err=tx.ExecContext(r.Context(),`UPDATE job_leases SET expires_at=now() WHERE attempt_id=$1`,attemptID);err!=nil{writeError(w,500,"could not close lease");return}
 if _,err=tx.ExecContext(r.Context(),`INSERT INTO job_events(job_id,event_type,payload) VALUES($1,$2,jsonb_build_object('node_id',$3,'attempt_id',$4,'error_code',$5))`,jobID,"job."+in.Status,nodeID,attemptID,in.ErrorCode);err!=nil{writeError(w,500,"could not record event");return}
 if err=tx.Commit();err!=nil{writeError(w,500,"could not commit completion");return}
 writeJSON(w,200,map[string]string{"job_id":jobID,"status":in.Status})
}
