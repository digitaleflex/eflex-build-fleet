# ADR-0002 — BuildKit comme exécuteur et Railpack comme stratégie automatique

- **Statut :** Accepted with validation gate
- **Date :** 2026-11-16
- **Portée :** build applicatif

## Contexte
E-FLEX doit accepter les applications standards, les Dockerfiles explicites et les dépôts dont la stack doit être détectée. Réécrire un moteur général de compilation créerait une charge importante sans différenciation suffisante.

## Décision
- BuildKit est le backend d'exécution.
- Railpack est la stratégie automatique initiale et produit un Build Plan.
- Dockerfile est un chemin explicite de première classe.
- Nixpacks est conservé pour compatibilité historique, sans fork principal.
- Cloud Native Buildpacks est différé.
- ko et Jib restent des extensions spécialisées éventuelles.
- Chaque moteur et sa configuration sont épinglés et enregistrés par build.
- Aucun fallback silencieux entre moteurs.

## Justification
BuildKit apporte l'exécution, le cache et l'export OCI. Railpack apporte l'analyse et la génération du plan. Les rôles sont complémentaires. Nixpacks est en maintenance selon sa documentation et recommande Railpack comme successeur.

## Gate de validation
Avant production : qualifier une version Railpack exacte et la version BuildKit correspondante ; tester Node/Next.js, Laravel, Python et Dockerfile ; tester cache, secrets, réseau et échecs ; mesurer temps, CPU, RAM et coût ; vérifier licences et dépendances transitives.

## Conséquences
Nous ne possédons pas le moteur bas niveau de construction d'images. Nous possédons la sélection, les adaptateurs, les politiques, la distribution, l'isolation et la gouvernance.

## Critères de réouverture
Uniquement si Railpack devient incompatible, non maintenu, vulnérable, juridiquement incompatible ou incapable de satisfaire une stack client démontrée. Un besoin de provider peut d'abord être traité par configuration ou adaptateur.
