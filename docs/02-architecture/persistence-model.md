# Modèle de persistance logique

## Objectif
Définir les entités durables du Control Plane sans figer prématurément les détails du schéma SQL. Les migrations restent la source de vérité du schéma réellement implémenté.

## Entités principales
| Entité | Rôle |
|---|---|
| tenants | Organisations isolées |
| projects | Projets appartenant à un tenant |
| build_jobs | Demandes logiques de compilation |
| build_attempts | Tentatives individuelles d'exécution |
| build_nodes | Inventaire, identité et état des agents/nœuds |
| node_capabilities | Capacités CPU, RAM, disque, moteurs et plateformes |
| leases | Autorisations temporaires d'exécution |
| job_events | Historique immuable des transitions utiles |
| outbox_events | Événements à publier après commit |
| artifacts | Images, digests, SBOM et provenance |
| tenant_quotas | Limites et compteurs de concurrence |
| audit_events | Actions opérateur et opérations sensibles |

## Relations
- Un tenant possède plusieurs projets.
- Un projet possède plusieurs jobs.
- Un job possède plusieurs attempts.
- Une attempt cible un node et possède au plus un lease actif.
- Un job peut produire plusieurs artefacts.
- Un tenant possède une politique et un quota.
- Les transitions sont référencées par job_events.
- Les publications sont d'abord enregistrées dans outbox_events.

## Contraintes essentielles
- UUID ou identifiants opaques non prédictibles pour les identifiants publics.
- Foreign keys pour les relations de propriété.
- Index tenant_id, project_id, state, created_at, next_attempt_at et lease_expires_at selon les requêtes.
- Unicité de la clé d'idempotence dans son périmètre tenant/projet.
- Unicité des identifiants d'agent et certificats actifs.
- Une seule tentative active par job logique, sauf fonctionnalité explicitement conçue.
- Lease avec expiration et version/fencing token pour éviter les rapports obsolètes.
- Événements et transitions durables dans la même transaction que l'état associé.
- Outbox publiée puis marquée après succès, sans suppression prématurée.

## Transaction scheduler
La réservation de capacité, l'attribution du job et la création du lease sont atomiques. Le scheduler doit empêcher deux transactions concurrentes de réserver la même capacité. Les locks sont courts ; aucune compilation n'est exécutée dans une transaction SQL.

## Rétention
Les logs volumineux et workspaces ne sont pas stockés dans les tables relationnelles. PostgreSQL conserve des références vers leur stockage. Rétention des jobs, tentatives, événements, audits et artefacts doit être définie séparément.

## Données sensibles
Aucun secret brut dans build_jobs, job_events, logs ou provenance. Les références de secret ont un identifiant opaque, une portée, une expiration et un statut. Les données de connexion et clés privées ne sont jamais stockées en clair dans la base.

## Sauvegardes
Les migrations et la version de schéma sont enregistrées. La restauration doit inclure le schéma, les jobs et les métadonnées d'artefacts de manière cohérente.
