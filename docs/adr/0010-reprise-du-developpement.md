# ADR-0010 — Reprise du développement de Build Fleet ; le chemin Coolify reste intérimaire

- **Statut :** Accepted
- **Date :** 2026-11-23
- **Remplace :** ADR-0009, point 1 (« le développement de Build Fleet est différé »). Les points 2 à 4 d'ADR-0009 restent en vigueur.
- **Ne remplace pas :** ADR-0001 à ADR-0008.

## Contexte
ADR-0009 (16 novembre) différait Build Fleet au profit d'un serveur de build Coolify. Le propriétaire a depuis confirmé que **le développement se poursuit** : le dépôt contient déjà un plan de contrôle, un agent, un exécuteur de build isolé, un planificateur avec réconciliation des baux, des mesures et des garde-fous (limitation de débit, isolation par client). La décision doit donc suivre la réalité plutôt que l'inverse.

## État réel constaté (relecture du dépôt le 2026-11-23, **sans exécution**)
- Code Go : environ 20 fichiers et 1 400 lignes, 6 fichiers de tests. Un tel volume correspond au début de la phase 1 de la feuille de route, pas à une flotte complète.
- Aucune preuve versée qu'un build de bout en bout (source → BuildKit → registre → digest) fonctionne sur un serveur réel.
- Aucun nœud de production, aucun mTLS validé, aucun registre qualifié : les portes de validation des ADR-0002, 0004, 0005 et 0006 restent **fermées**.

## Décision
1. **Le développement de Build Fleet reprend.** Il suit la feuille de route (`docs/06-roadmap/roadmap.md`) dans l'ordre : une phase n'est terminée que lorsque ses critères sont démontrés.
2. **Le chemin intérimaire reste le chemin de production** : serveur de build Coolify dédié + registre OCI, réseau privé, runner pour les applications Compose (ADR-0009, points 2 et 3). **Aucun client n'est basculé sur Build Fleet** avant la sortie de la phase 2 (premiers builds représentatifs reproductibles en préproduction) **et** la réussite des portes de validation des ADR-0002, 0005 et 0006.
3. **Le registre est une donnée à sauvegarder.** Les images qu'il contient sont la seule trace des artefacts construits ; sa perte obligerait à tout reconstruire. Le plan de sauvegarde (voir `docs/04-reliability/backup-and-restore.md`) prévoit au moins deux destinations indépendantes : un stockage objet auto-hébergé (AURA) et Backblaze B2, avec chiffrement avant envoi et restauration testée depuis chacune.
4. **Question ouverte, à trancher par un ADR dédié avant la phase 3 :** « un seul moteur Go à terme » (ADR-0009, point 4). Le dépôt `eflex.app` contient le moteur Axiom, en pause (≈ 61 000 lignes), et Build Fleet est un second code Go actif. Deux options : Build Fleet reprend ou remplace les agents d'Axiom, ou Axiom est archivé hors du dépôt principal.

## Conséquences
- Les efforts se répartissent entre le produit (E-FLEX App), la fiabilité du chemin intérimaire et Build Fleet. **Risque assumé** : un projet porté par une seule personne ; la règle de la phase terminée « par preuve » sert de garde-fou.
- Les documents d'architecture décrivent à nouveau un système **en construction**, et non plus seulement une cible.
- Tant que la phase 2 n'est pas démontrée, E-FLEX App ne dépend pas de Build Fleet.

## Critères de réouverture
Abandon de Build Fleet ; dépassement du temps disponible constaté sur le produit ; décision du propriétaire.
