# ADR-0007 — Séparation E-FLEX App, Build Fleet et Runtime

- **Statut :** Accepted
- **Date :** 2026-11-16
- **Portée :** frontière produit et infrastructure

## Contexte
Le portail, le moteur de build et les serveurs runtime ont des besoins de disponibilité, sécurité et mise à jour différents.

## Décision
- E-FLEX App reste le produit et l'interface client.
- Build Fleet est un service d'infrastructure autonome dans son propre dépôt.
- Le Runtime Plane reste opéré par E-FLEX App / Coolify dans la stratégie actuelle.
- Communication par API et événements versionnés.
- Aucune base de données partagée.
- Aucun déploiement applicatif directement depuis le Build Agent.
- Aucune console Next.js indépendante pour le MVP.

## Justification
La séparation permet de changer les moteurs, serveurs et fournisseurs de build sans réécrire le portail ou le runtime. Elle réduit les privilèges et évite qu'un incident de build bloque les applications déjà exécutées.

## Conséquences
Les contrats API, l'identité inter-service et la gestion des versions deviennent des responsabilités explicites. Les mises à jour du portail et de Build Fleet restent indépendantes.

## Flux
E-FLEX App demande un build, suit son état et reçoit un résultat. Build Fleet publie une image OCI et retourne son digest. E-FLEX App décide si elle doit demander à Coolify de déployer ce digest.
