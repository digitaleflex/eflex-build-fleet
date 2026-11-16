package controlplane

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// ReconcileExpiredLeases marks attempts whose leases have expired as lost,
// returns their jobs to the queued state (when not already terminal) and
// records the corresponding events. Returns the number of reconciled leases.
func (s *Server) ReconcileExpiredLeases(ctx context.Context) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT l.job_id::text, l.attempt_id::text, l.node_id::text, l.fencing_token
		FROM job_leases l JOIN build_attempts a ON a.id=l.attempt_id
		WHERE l.expires_at < now() AND a.status IN ('assigned','running') FOR UPDATE OF l`)
	if err != nil {
		return 0, err
	}
	type expired struct {
		jobID, attemptID, nodeID string
		fencing                  int64
	}
	var list []expired
	for rows.Next() {
		var e expired
		if err := rows.Scan(&e.jobID, &e.attemptID, &e.nodeID, &e.fencing); err != nil {
			rows.Close()
			return 0, err
		}
		list = append(list, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	count := 0
	for _, e := range list {
		if _, err := tx.ExecContext(ctx, `UPDATE build_attempts SET status='lost',finished_at=now(),error_code='lease_expired' WHERE id=$1 AND status IN ('assigned','running')`, e.attemptID); err != nil {
			return count, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE build_jobs SET status='queued',updated_at=now() WHERE id=$1 AND status IN ('scheduled','running')`, e.jobID); err != nil {
			return count, err
		}
		_, _ = tx.ExecContext(ctx, `INSERT INTO job_events(job_id,event_type,payload) VALUES($1,'job.lease_expired',jsonb_build_object('node_id',$2::text,'attempt_id',$3::text,'fencing_token',$4::bigint))`, e.jobID, e.nodeID, e.attemptID, e.fencing)
		_, _ = tx.ExecContext(ctx, `UPDATE job_leases SET expires_at=now() WHERE attempt_id=$1`, e.attemptID)
		count++
	}
	if err := tx.Commit(); err != nil {
		return count, err
	}
	return count, nil
}

// RunSchedulerLoop periodically reconciles expired leases and dispatches queued
// jobs until ctx is cancelled.
func (s *Server) RunSchedulerLoop(ctx context.Context, interval time.Duration, logger *slog.Logger) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n, err := s.ReconcileExpiredLeases(ctx); err != nil {
				logger.Warn("lease reconciliation failed", "error", err)
			} else if n > 0 {
				logger.Info("reconciled expired leases", "count", n)
			}
			// dispatch as many queued jobs as possible
			for {
				// reuse dispatchNext logic by calling it via a synthetic request is awkward;
				// instead, call an internal helper.
				ok, err := s.dispatchOne(ctx)
				if err != nil {
					logger.Warn("dispatch failed", "error", err)
					break
				}
				if !ok {
					break
				}
			}
		}
	}
}

// dispatchOne performs a single transactional dispatch. Returns true if a job was dispatched.
func (s *Server) dispatchOne(ctx context.Context) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var jobID string
	err = tx.QueryRowContext(ctx, `SELECT id::text FROM build_jobs WHERE status='queued' ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&jobID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var nodeID string
	err = tx.QueryRowContext(ctx, `SELECT n.id::text FROM build_nodes n WHERE n.status='ready' AND n.last_heartbeat_at>now()-interval '90 seconds' AND NOT EXISTS (SELECT 1 FROM unnest((SELECT required_capabilities FROM build_jobs WHERE id=$1)) AS rc(cap) WHERE rc.cap NOT IN (SELECT capability FROM node_capabilities WHERE node_id=n.id)) ORDER BY n.last_heartbeat_at DESC LIMIT 1 FOR UPDATE SKIP LOCKED`, jobID).Scan(&nodeID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var attemptNo int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(attempt_number),0)+1 FROM build_attempts WHERE job_id=$1`, jobID).Scan(&attemptNo); err != nil {
		return false, err
	}
	var fencing int64
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(fencing_token),0)+1 FROM job_leases WHERE job_id=$1`, jobID).Scan(&fencing); err != nil {
		return false, err
	}
	var attemptID string
	if err = tx.QueryRowContext(ctx, `INSERT INTO build_attempts(job_id,node_id,attempt_number,status,started_at) VALUES($1,$2,$3,'assigned',now()) RETURNING id::text`, jobID, nodeID, attemptNo).Scan(&attemptID); err != nil {
		return false, err
	}
	var expires time.Time
	if err = tx.QueryRowContext(ctx, `INSERT INTO job_leases(job_id,attempt_id,node_id,fencing_token,expires_at) VALUES($1,$2,$3,$4,now()+interval '60 seconds') RETURNING expires_at`, jobID, attemptID, nodeID, fencing).Scan(&expires); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE build_jobs SET status='scheduled',updated_at=now() WHERE id=$1`, jobID); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO job_events(job_id,event_type,payload) VALUES($1,'job.scheduled',jsonb_build_object('node_id',$2::text,'attempt_id',$3::text,'fencing_token',$4::bigint))`, jobID, nodeID, attemptID, fencing); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO outbox_events(aggregate_type,aggregate_id,event_type,payload) VALUES('build_job',$1,'job.scheduled',jsonb_build_object('node_id',$2::text,'attempt_id',$3::text,'fencing_token',$4::bigint))`, jobID, nodeID, attemptID, fencing); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}