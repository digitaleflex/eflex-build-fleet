# E-FLEX Build Fleet

Architecture baseline v1.0 — 16 November 2026.

E-FLEX Build Fleet is the independent build infrastructure of E-FLEX. It schedules and runs application builds on independent machines, publishes immutable OCI artifacts and reports results to E-FLEX App.

## Read in this order

1. Product charter
2. Reference architecture
3. Build lifecycle
4. Fleet scheduler
5. Engine strategy
6. Artifact and registry strategy
7. Network and identity
8. Threat model and build isolation
9. Reliability and operations
10. Roadmap
11. Architecture decision records

## Accepted baseline

- Go for Control Plane and Build Agent.
- Modular monolith Control Plane plus a separately deployed Go Build Agent.
- E-FLEX-owned scheduler specialized for build jobs.
- PostgreSQL as durable job state, with leases and transactional outbox.
- BuildKit as image build executor.
- Railpack as the initial automatic build planner, pinned and qualified.
- Nixpacks for legacy compatibility, not as the strategic engine.
- Tailscale for private management connectivity, with mTLS and application authorization still mandatory.
- OCI image digests as the immutable artifact reference.
- Zot as the first self-hosted registry candidate to qualify behind an OCI abstraction.
- E-FLEX App remains the product interface; no separate console in the MVP.
- Runtime and deployment remain outside Build Fleet.

## Status discipline

An accepted decision is an architecture choice, not proof that it has been implemented or production-qualified. The roadmap defines evidence required before activation. Decisions may be reopened only under the rules in docs/00-governance/decision-policy.md.

The repository has no LICENSE file yet. Public visibility does not grant permission to reuse or redistribute the E-FLEX code.
