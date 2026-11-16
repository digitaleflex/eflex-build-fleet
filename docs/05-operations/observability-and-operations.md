# Observabilité et exploitation

## Objectif
Chaque job, nœud et artefact doit être traçable. L'observabilité sert au diagnostic, à la sécurité, au dimensionnement et à la mesure économique.

## Identifiants de corrélation
Chaque événement est relié à tenant_id, project_id, job_id, attempt_id, node_id, source_commit, engine_id, engine_version et artifact_digest.

Les logs ne contiennent pas de secrets ni de contenu source non nécessaire.

## Métriques du Control Plane
- jobs reçus, en queue, actifs, terminés et échoués ;
- temps d'attente et dispatch ;
- durée des leases ;
- expirations et retries ;
- jobs par tenant et priorité ;
- erreurs API et latence ;
- état PostgreSQL et backlog outbox.

## Métriques des agents
- âge du heartbeat ;
- CPU/RAM/disque disponibles ;
- concurrence active ;
- espace workspace ;
- temps de préparation, build et export ;
- cache hit/miss ;
- taux d'échec par moteur/version ;
- nettoyage échoué ;
- spool local en attente ;
- erreurs réseau et Registry.

## Métriques économiques
- CPU-seconds et mémoire consommée ;
- octets transférés ;
- stockage et durée de cache ;
- temps total d'occupation du nœud ;
- coût estimé par build ;
- coût par build réussi ;
- coût des retries et échecs.

Les coûts sont d'abord estimés à partir des factures et métriques fournisseurs. Ne pas annoncer une facturation précise avant rapprochement.

## Logs et traces
Prometheus est la référence initiale des métriques. OpenTelemetry est le standard d'instrumentation et de propagation de contexte. Les logs structurés sont corrélés par job et tentative.

Le stockage et la rétention des logs sont dimensionnés et protégés contre l'exfiltration de secrets.

## Alertes initiales
- Control Plane indisponible.
- PostgreSQL indisponible ou retard de sauvegarde.
- Agent offline ou quarantined.
- Expiration de lease en hausse.
- Queue qui vieillit.
- Nœud CPU/RAM/disque saturé.
- Export Registry échoué.
- Nettoyage workspace échoué.
- Certificat ou credential proche expiration.
- Échecs répétés après mise à jour d'un moteur.
- Écart de digest ou signature invalide.

## Runbooks obligatoires avant production
1. Ajouter, enrôler et retirer un Build Node.
2. Révoquer l'identité d'un agent.
3. Drainer un nœud.
4. Diagnostiquer un lease bloqué.
5. Réconcilier un résultat après partition réseau.
6. Restaurer PostgreSQL.
7. Restaurer ou remplacer le Registry.
8. Révoquer un credential compromis.
9. Purger un cache tenant.
10. Mettre un moteur en quarantaine.
11. Réagir à une compromission de Build Node.
12. Perte d'un fournisseur d'hébergement.

## Sauvegardes
Les sauvegardes couvrent la base du Control Plane, configuration critique, inventaire et métadonnées. Images OCI et caches suivent une politique séparée de réplication/rétention.

Les clés de chiffrement et credentials ne sont pas sauvegardés dans le même périmètre que la base. Leur récupération est documentée et testée.

Une sauvegarde non restaurée en environnement isolé n'est pas considérée comme vérifiée.
