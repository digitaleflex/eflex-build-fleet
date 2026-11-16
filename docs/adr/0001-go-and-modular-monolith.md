# ADR-0001 — Go et monolithe modulaire pour le Control Plane

- **Statut :** Accepted
- **Date :** 2026-11-16
- **Portée :** architecture logicielle v1

## Contexte
Build Fleet doit gérer des connexions réseau longues, de nombreux agents, des jobs asynchrones, le contrôle de ressources et l'intégration à BuildKit. La base doit être exploitable par une petite équipe sans multiplier les langages et services.

## Décision
- Go est le langage principal du Control Plane et du Build Agent.
- Le Control Plane est un monolithe modulaire Go.
- Le Build Agent est un binaire Go séparé.
- Les frontières internes sont des interfaces et contrats explicites.
- L'API produit est REST/OpenAPI.
- La communication agents-Control Plane est gRPC/Protobuf avec mTLS.
- Pas de console web indépendante dans la v1.

## Alternatives examinées
- Node.js/TypeScript pour tout : productif pour l'API, moins naturel pour les agents système et l'intégration Go de BuildKit.
- Rust pour tout : garanties mémoire intéressantes, coût de développement et maintenance plus élevé pour le périmètre initial.
- Microservices dès le départ : complexité de déploiement, observabilité et réseau disproportionnée.

## Justification
Go est proche de l'écosystème BuildKit, permet des binaires autonomes et une concurrence adaptée aux agents réseau. Le monolithe modulaire réduit le nombre de services à exploiter tout en gardant les composants séparables si nécessaire.

## Conséquences
Positives : un langage principal, intégration directe avec les bibliothèques Go, déploiement simple, moins de dépendances opérationnelles.

Négatives : discipline de modules et d'interfaces obligatoire ; le Control Plane peut devenir un point de charge dont la séparation future nécessitera des preuves.

## Critères de réouverture
Uniquement si des benchmarks démontrent une limite importante, si une bibliothèque essentielle devient indisponible ou si une exigence de sécurité ne peut être satisfaite.
