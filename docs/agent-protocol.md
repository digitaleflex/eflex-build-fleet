# Build Agent protocol v1

All agent endpoints require the per-node Bearer credential issued by the Control Plane. The credential is stored only as a SHA-256 hash and is shown once at issuance.

- `POST /agent/v1/heartbeat`: update liveness and agent version.
- `GET /agent/v1/jobs/next`: return the active assignment for this node or `{"job":null}`.
- `POST /agent/v1/jobs/{id}/started`: acknowledge the assigned attempt.
- `POST /agent/v1/jobs/{id}/lease`: renew a running lease for 60 seconds.
- `POST /agent/v1/jobs/{id}/complete`: report `succeeded` or `failed`, optionally with an error code.

Node identity is derived from the credential, never from a request field. Mutations require an unexpired lease. The assignment includes a fencing token. This protocol tracks assignments but does not execute source code, build OCI images, or publish artifacts.
