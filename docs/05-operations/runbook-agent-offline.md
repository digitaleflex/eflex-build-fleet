# Runbook — agent offline et job bloqué

1. Vérifier `/readyz` et PostgreSQL.
2. Lister les nœuds : `GET /api/v1/nodes` ; un nœud `ready` dont `last_heartbeat_at` est ancien est marqué `offline` par le scheduler.
3. Pour un job bloqué en `scheduled`/`running`, vérifier les leases : `job_leases` et `build_attempts`.
4. Expirer manuellement un lease de test : `UPDATE job_leases SET expires_at=now()-interval '1 second' WHERE ...`, puis déclencher `ReconcileExpiredLeases` via le scheduler ou un redémarrage.
5. Révoquer un agent : `POST /api/v1/nodes/{id}/credentials` émet un nouveau token et révoque l'ancien ; l'ancien token est refusé.
6. Aucun secret ne doit apparaître dans les logs ; vérifier `LOG_LEVEL`.