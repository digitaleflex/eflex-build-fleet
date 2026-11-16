# Sauvegarde et restauration

## Périmètre
Les données critiques du Build Fleet sont :
- base PostgreSQL et migrations ;
- configuration du Control Plane ;
- inventaire des nœuds et états des identités ;
- métadonnées de jobs, tentatives, leases et événements ;
- références et métadonnées d'artefacts ;
- configuration et clés nécessaires à la récupération.

Les workspaces temporaires et caches ne sont pas des données de sauvegarde prioritaires. Les images OCI suivent une politique séparée de réplication et rétention.

## Décisions
- Les sauvegardes de la base sont chiffrées et stockées hors du fournisseur primaire.
- La sauvegarde n'est pas stockée uniquement sur le même disque ou VPS que la base.
- Les clés nécessaires à la restauration sont séparées des données sauvegardées.
- Le registre et le stockage objet ont leur propre plan de sauvegarde/réplication.
- RPO et RTO ne sont pas déclarés avant mesure.

## Procédure de restauration
1. Déclarer l'incident et choisir le point de restauration.
2. Isoler l'environnement de restauration.
3. Récupérer la sauvegarde et vérifier son intégrité.
4. Récupérer les clés de déchiffrement par le processus distinct.
5. Restaurer PostgreSQL et appliquer/valider les migrations.
6. Vérifier tenants, projets, jobs et événements.
7. Réconcilier les leases avec les agents connus ; invalider les leases incertains.
8. Vérifier que les digests et références de registry sont disponibles.
9. Révoquer les identités potentiellement compromises.
10. Réactiver progressivement le Control Plane et les agents.
11. Documenter le temps réel de restauration et la perte de données.

## Contrôle d'intégrité
Une sauvegarde est considérée vérifiée uniquement après restauration dans un environnement isolé et exécution de contrôles de cohérence.

## Cadence et objectifs
La fréquence de sauvegarde, rétention, RPO et RTO doivent être fixés en fonction de l'usage et du coût. Avant la première production, une cadence explicite et un test de restauration sont obligatoires. Les objectifs annoncés ne peuvent être supérieurs aux résultats observés.

## Sauvegarde du registre
Les images et attestations doivent être répliquées ou sauvegardées séparément. Une image n'est déclarée récupérable que si le manifest, ses blobs, sa signature et sa provenance peuvent être vérifiés après restauration.

## Incident de clés
La perte des clés peut rendre une sauvegarde chiffrée inutilisable. Les procédures de récupération, séparation, accès et rotation des clés doivent être testées indépendamment.
