# ADR-0003 — PostgreSQL comme source de vérité des jobs

- **Statut :** Accepted
- **Date :** 2026-11-16
- **Portée :** état durable et distribution

## Contexte
Les jobs doivent survivre aux redémarrages du Control Plane et aux pannes d'agents. L'état ne doit pas dépendre exclusivement d'une queue en mémoire ou d'un broker non durable.

## Décision
- PostgreSQL conserve l'état des jobs, tentatives, leases, nœuds, événements et outbox.
- Le scheduler utilise des transactions, réservations atomiques et leases expirables.
- Les agents réclament les travaux par connexion sortante.
- Le système utilise une sémantique at-least-once.
- L'idempotence est obligatoire sur les transitions et exports.
- Redis n'est pas une dépendance obligatoire de la v1.
- Une transactional outbox publie les événements après commit.

## Justification
Cette décision minimise les dépendances critiques et garde l'état durable au même endroit. Elle évite de synchroniser plusieurs sources de vérité dès le début.

## Conséquences
PostgreSQL est un composant critique et un point de défaillance en v1. Sauvegardes hors site, alertes et tests de restauration sont obligatoires. Le scheduler doit éviter les transactions longues.

## Critères de réouverture
Mesures démontrant que PostgreSQL ne satisfait plus le volume, la latence ou la concurrence requise. Toute migration doit préserver l'état durable et la reprise.
