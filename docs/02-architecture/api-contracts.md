# Contrats API et messages

## Statut
Ce document fixe les champs et sémantiques du contrat v1. Les schémas OpenAPI et Protobuf compilables seront produits dans contracts/ avant la première implémentation. Les contrats ne doivent pas être déduits directement des modèles de base de données.

## API Control Plane consommée par E-FLEX App
Routes logiques initiales :
- POST /v1/builds : soumettre un job ; retourne un identifiant et un statut initial.
- GET /v1/builds/{job_id} : lire l'état et le résultat public du job.
- GET /v1/builds/{job_id}/events : consulter les événements ou utiliser un flux équivalent documenté.
- POST /v1/builds/{job_id}/cancel : demander l'annulation.
- GET /v1/capabilities : exposer les capacités de plateforme compatibles avec l'identité appelante.

L'authentification et l'autorisation sont obligatoires. Chaque endpoint vérifie le tenant et le projet. Les appels de création exigent une clé d'idempotence.

## BuildRequest — champs conceptuels
- request_id : identifiant de requête.
- tenant_id et project_id : contexte d'autorisation.
- source.repository : URL ou identifiant du dépôt autorisé.
- source.commit_sha : commit immuable obligatoire.
- source.context_path : répertoire de contexte relatif.
- source.snapshot_digest : digest du snapshot si disponible.
- strategy.mode : automatic, dockerfile ou explicit.
- strategy.engine_id / engine_version : optionnels à la demande, contrôlés par policy.
- target.platform : OS/architecture.
- resources.profile : profil serveur pré-approuvé.
- cache.policy : disabled, tenant-private ou policy-approved.
- network.policy_id : règle réseau pré-approuvée.
- secrets.refs : références temporaires, jamais valeurs secrètes.
- idempotency_key : déduplication des demandes répétées.
- timeout/deadline : bornes validées côté serveur.

Aucun champ envoyé par un client ne peut accorder un privilège, choisir un nœud précis ou augmenter un quota.

## BuildResult — champs conceptuels
- job_id et attempt_id ;
- status et error_class ;
- source_commit et source_snapshot_digest ;
- selected_engine, version et config_digest ;
- image_ref et image_digest ;
- node_id et failure_domain ;
- started_at, completed_at et duration ;
- logs_ref ;
- resource_usage ;
- sbom_ref et provenance_ref si disponibles ;
- warnings et diagnostics non sensibles.

Les informations internes sensibles du nœud, tokens, chemins locaux et secrets ne sont jamais retournés au client.

## Messages Agent / Control Plane
Opérations logiques :
- RegisterAgent : présenter identité, version et capacités.
- Heartbeat : état, capacité et santé.
- ClaimWork : réclamer une tâche admissible.
- RenewLease : renouveler une autorisation de tâche.
- ReportProgress : étape et progression.
- CompleteAttempt : résultat signé/attesté par l'agent de confiance, digest et métriques.
- FailAttempt : classification d'erreur et références de logs.
- DrainNode : arrêter l'attribution de nouveaux jobs.
- RevokeAgent : révocation et quarantaine.

Toutes les opérations sont authentifiées par mTLS et autorisées selon node identity, job, tenant et lease.

## Compatibilité
Les messages Protobuf sont versionnés et évoluent de façon additive. Un agent ancien ne reçoit jamais un champ ou une commande qu'il ne sait pas interpréter. La politique de compatibilité doit être définie avant le premier déploiement multi-version.

## Erreurs et idempotence
Les erreurs possèdent un code stable, une catégorie et un message opérateur. Les messages client ne divulguent pas de détails internes sensibles.

POST /v1/builds avec la même clé d'idempotence et même payload retourne le même job logique. Une clé réutilisée avec un payload différent retourne un conflit.

ClaimWork, RenewLease et CompleteAttempt sont conçus pour tolérer les répétitions réseau. Un rapport provenant d'un lease expiré est mis en réconciliation et n'est pas automatiquement promouvable.

## Séparation des contrats
API produit, protocole agent, BuildRequest interne, BuildResult et schéma SQL sont des contrats distincts. Aucun type DB n'est sérialisé directement.
