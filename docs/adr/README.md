# Architecture Decision Records

Les ADR constituent l'historique des décisions techniques et produit structurantes. Les statuts et le processus de modification sont définis dans [la politique de décision](../00-governance/decision-policy.md).

## Baseline v1.0 — 16 novembre 2026

| ADR | Décision | Statut |
|---|---|---|
| [0001](0001-go-and-modular-monolith.md) | Go et monolithe modulaire | Accepted |
| [0002](0002-buildkit-railpack-engine-strategy.md) | BuildKit et Railpack | Accepted with validation gate |
| [0003](0003-postgresql-jobs-and-leases.md) | PostgreSQL, jobs et leases | Accepted |
| [0004](0004-tailscale-and-mutual-tls.md) | Tailscale et mTLS | Accepted with validation gate |
| [0005](0005-oci-registry-and-zot.md) | OCI et qualification de Zot | Accepted with validation gate |
| [0006](0006-untrusted-build-isolation.md) | Isolation des builds non fiables | Accepted with validation gate |
| [0007](0007-product-and-runtime-boundaries.md) | Frontières E-FLEX App / Build Fleet / Runtime | Accepted |
| [0008](0008-no-general-purpose-orchestrator-v1.md) | Pas de Kubernetes, Nomad ou Temporal obligatoires en v1 | Accepted |
| [0009](0009-deferred-start-coolify-build-server.md) | Démarrage différé ; serveur de build Coolify en attendant | Accepted |

Un ADR « Accepted with validation gate » fixe la direction, mais interdit de considérer la capacité comme prête en production avant réussite des tests indiqués.

## Règle d'évolution
Ne jamais modifier silencieusement un ADR accepté. Créer un nouvel ADR qui cite celui qu'il remplace, puis passer l'ancien en Superseded. Conserver les décisions historiques.
