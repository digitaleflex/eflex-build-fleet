# Roadmap de référence

## Règle
Une phase est terminée uniquement lorsque ses critères sont démontrés par tests, mesures ou procédures répétables. Une API ou un écran terminé ne constitue pas une preuve opérationnelle.

## Phase 0 — Baseline et gouvernance
- Architecture de référence et ADR acceptés.
- Invariants documentés.
- Contrats Job / Result définis.
- Threat model et frontières de confiance établis.
- Modèle de licences et inventaire des dépendances défini.

**Sortie :** décisions centralisées et responsabilités claires.

## Phase 1 — Control Plane et agent minimal
- Service Go Control Plane.
- Binaire Go Build Agent.
- PostgreSQL et migrations.
- Enregistrement et révocation d'agent.
- mTLS, heartbeat, claim, lease et report.
- Job state machine et outbox.
- Logs structurés et health checks.

**Sortie :** un agent distant reçoit une tâche de test, l'exécute et renvoie un résultat durable ; arrêt et reprise contrôlés.

## Phase 2 — Première compilation
- Intégration BuildKit.
- Adaptateur Dockerfile.
- Adaptateur Railpack avec version épinglée.
- Snapshot source au commit exact.
- Limites CPU/RAM/disque/durée.
- Export OCI.
- Résultat par digest et métadonnées.

**Sortie :** plusieurs applications représentatives compilées en staging avec résultats reproductibles au niveau des entrées, diagnostics et artefacts vérifiés.

## Phase 3 — Flotte multi-nœuds
- Plusieurs agents et nœuds.
- Scheduler d'admissibilité et score.
- Réservation atomique de capacité.
- Quotas tenant et concurrence.
- Draining, offline et quarantined.
- Retry, expiration lease et réconciliation.
- Nœuds sur au moins deux domaines de panne.

**Sortie :** test de perte d'un nœud, redistribution et aucune perte d'état durable.

## Phase 4 — Isolation client
- Sandbox et isolation réseau.
- Secrets temporaires.
- Séparation des caches.
- Tests d'accès réseau interdit.
- Profil Firecracker/gVisor qualifié sur les fournisseurs retenus ou interdiction du multi-tenant partagé lorsque non disponible.

**Sortie :** revue de sécurité, tests et preuves d'isolation avant builds clients non fiables.

## Phase 5 — Registre et supply chain
- Qualification de Zot.
- RegistryProvider OCI.
- Politique de rétention et réplication.
- SBOM, scan, signature et provenance.
- Vérification digest au déploiement.

**Sortie :** chaque image promouvable est vérifiable et récupérable depuis un registre indépendant.

## Phase 6 — Intégration E-FLEX App
- API de soumission et statut.
- Webhooks/événements versionnés.
- Mapping projets et tenants.
- Transmission des images par digest à Coolify.
- Chemin Compose multi-images.
- Migration contrôlée depuis le build Coolify direct.

**Sortie :** E-FLEX App consomme Build Fleet sans base de données partagée et sans dépendre des Build Nodes.

## Phase 7 — Résilience et économie
- Backups hors site et restore tests.
- Exercices de perte de fournisseur.
- Tests de panne Registry, PostgreSQL, réseau et Control Plane.
- Benchmarks de concurrence.
- Coût par build réussi.
- SLO/RPO/RTO définis à partir des résultats.

**Sortie :** limites et niveaux de service publiables, démontrés par les tests.

## Phase 8 — Extensions justifiées
- Nixpacks pour compatibilité.
- CNB si cas client pertinent.
- ko/Jib si gain mesuré.
- Autoscaling de machines.
- Headscale ou autre réseau si besoin établi.
- Séparation de modules du Control Plane si charge mesurée.

**Sortie :** chaque extension possède une preuve de valeur et un ADR.

## Hors séquence
Réécriture de BuildKit, fork préventif de moteurs, Kubernetes obligatoire dès v1, Nomad ou Temporal avant preuve de besoin, multi-région actif-actif avant maîtrise de la reprise, console indépendante avant besoin opérateur réel.
