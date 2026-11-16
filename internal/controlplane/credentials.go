package controlplane

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type credentialResponse struct {
	NodeID string `json:"node_id"`
	Token string `json:"token"`
	Warning string `json:"warning"`
}

func (s *Server) issueNodeCredential(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid node id")
		return
	}
	var exists bool
	if err := s.db.QueryRowContext(r.Context(), "SELECT EXISTS(SELECT 1 FROM build_nodes WHERE id=$1)", id).Scan(&exists); err != nil {
		writeError(w, http.StatusInternalServerError, "could not verify node")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "node not found")
		return
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate credential")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil { writeError(w, http.StatusInternalServerError, "could not rotate credential"); return }
	defer tx.Rollback()
	if _, err := tx.ExecContext(r.Context(), "UPDATE node_credentials SET revoked_at=now() WHERE node_id=$1 AND revoked_at IS NULL", id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not rotate credential")
		return
	}
	if _, err := tx.ExecContext(r.Context(), "INSERT INTO node_credentials(node_id,token_hash) VALUES($1,$2)", id, hash[:]); err != nil {
		writeError(w, http.StatusInternalServerError, "could not store credential")
		return
	}
	if err := tx.Commit(); err != nil { writeError(w, http.StatusInternalServerError, "could not commit credential rotation"); return }
	writeJSON(w, http.StatusCreated, credentialResponse{NodeID:id, Token:token, Warning:"Store this token securely. It will not be shown again."})
}

func (s *Server) authenticateNode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "node credential required")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			writeError(w, http.StatusUnauthorized, "node credential required")
			return
		}
		hash := sha256.Sum256([]byte(token))
		var nodeID string
		err := s.db.QueryRowContext(r.Context(), `UPDATE node_credentials SET last_used_at=now()
			WHERE token_hash=$1 AND revoked_at IS NULL
			RETURNING node_id::text`, hash[:]).Scan(&nodeID)
		if err != nil {
			if errors.Is(err, errNoRows) {
				writeError(w, http.StatusUnauthorized, "invalid node credential")
				return
			}
			writeError(w, http.StatusInternalServerError, "credential verification failed")
			return
		}
		ctx := withNodeID(r, nodeID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
