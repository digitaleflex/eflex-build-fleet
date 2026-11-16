# ADR-0008 — Pas d'orchestrateur généraliste obligatoire en v1

- **Statut :** Accepted
- **Date :** 2026-11-16
- **Portée :** orchestration et dépendances

## Contexte
Les orchestrateurs généralistes et moteurs de workflow apportent des capacités importantes, mais ajoutent des services, des opérations et des dépendances à administrer.

## Décision
- Un scheduler spécialisé E-FLEX est développé en Go.
- Il gère un nombre limité de nœuds et jobs de compilation.
- PostgreSQL fournit l'état durable, les leases et l'outbox.
- Kubernetes, Nomad et Temporal ne sont pas des dépendances de la première version.
- Le scheduler n'essaie pas de reproduire un orchestrateur généraliste.

## Alternatives examinées
- Kubernetes : puissant mais charge opérationnelle importante pour la flotte initiale.
- Nomad : pertinent pour le batch, mais la licence actuelle doit être examinée avant tout usage commercial intégré ou concurrent.
- Temporal : workflows durables robustes, mais service supplémentaire et complexité avant preuve du besoin.

## Critères de réouverture
La charge réelle dépasse les capacités mesurées du scheduler ; les reprises deviennent trop complexes ; le coût d'exploitation d'une solution spécialisée dépasse une alternative ; ou le provisionnement dynamique devient un besoin démontré.

Tout changement exige benchmark et ADR.
