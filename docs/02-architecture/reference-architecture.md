# Architecture de référence

## Modèle général
E-FLEX Build Fleet est composé de deux applications Go et de services d'infrastructure indépendants :

1. **Control Plane** : API, état durable des jobs, scheduler, inventaire de flotte, politiques, quotas et audit.
2. **Build Agent** : agent installé sur chaque Build Node ; il réclame les tâches autorisées, prépare le workspace, exécute les adaptateurs et publie les résultats.

PostgreSQL est la source de vérité initiale. BuildKit réalise l'exécution des builds. Railpack fournit la stratégie automatique initiale. Un registre OCI stocke les artefacts. Tailscale fournit un réseau privé de gestion, complété par mTLS applicatif.

## Diagramme logique
    E-FLEX App
        | REST/OpenAPI versionnée
        v
    E-FLEX Build Fleet Control Plane (Go)
        |-- API / Auth
        |-- Job Service / State Machine
        |-- Fleet Inventory
        |-- Scheduler / Quotas / Policies
        |-- Event Outbox / Audit
        +---- PostgreSQL (état durable)
        |
        | gRPC / mTLS, connexions initiées par les agents
        v
    Build Agent (Go) sur chaque Build Node
        |-- Heartbeat / Claim / Lease
        |-- Workspace / Source Snapshot
        |-- Resource Enforcement / Cleanup
        |-- Builder Adapters
        |       |-- Railpack -> Build Plan
        |       |-- Dockerfile -> BuildKit
        |       '-- BuildKit Executor
        +---- OCI Registry (push par digest)
                         |
                         v
                  E-FLEX App / Runtime
                  Coolify déploie l'image

## Topologie physique
Les rôles sont séparés logiquement et peuvent partager temporairement une machine lorsque le risque est maîtrisé. Les Build Nodes doivent être séparés des serveurs runtime. Le Control Plane et PostgreSQL sont protégés contre les tâches de compilation.

Topologie cible :
- un ou plusieurs nœuds Control Plane selon la stratégie HA ultérieure ;
- PostgreSQL avec sauvegardes hors site et restauration qualifiée ;
- plusieurs Build Nodes répartis sur des domaines de panne distincts ;
- un registre OCI principal et un mécanisme de réplication/reprise ;
- des serveurs Runtime opérés séparément par E-FLEX App / Coolify.

Un déploiement n'est qualifié de multi-provider que si les nœuds sont réellement répartis entre fournisseurs et si les chemins réseau, artefacts et sauvegardes ont été testés.

## Forme logicielle
Le Control Plane est un **monolithe modulaire Go** au départ. API, scheduler, jobs, inventaire, quotas et événements sont des modules du même processus et du même dépôt. Ils ne sont pas séparés en microservices sans preuve de besoin indépendant.

Le Build Agent est un binaire Go séparé, déployé sur les Build Nodes. Il possède son identité et son cycle de mise à jour.

L'interface d'administration dédiée est différée. E-FLEX App reste l'interface client.

## Contrats
L'API produit est REST documentée par OpenAPI. La communication agents-Control Plane utilise gRPC et Protobuf avec mTLS. Les modèles internes ne sont pas exposés directement. Les ruptures de compatibilité exigent une version majeure ou une période de compatibilité documentée.

## Responsabilités
| Composant | Responsabilité | Ne fait pas |
|---|---|---|
| E-FLEX App | Expérience client, configuration, autorisations produit | Exécuter un build sur le runtime |
| Control Plane | État jobs, placement, politiques, quotas, identité agents | Compiler le code client |
| Scheduler | Choisir un nœud admissible et accorder un lease | Déployer l'application |
| Build Agent | Exécuter, mesurer, nettoyer, publier | Définir seul les politiques globales |
| BuildKit | Exécution des builds | Gérer les tenants E-FLEX |
| Railpack | Détection et Build Plan | Gérer la flotte |
| OCI Registry | Stocker et distribuer les artefacts | Être la source de vérité des jobs |
| Coolify / Runtime | Déployer et exploiter l'image | Être le scheduler Build Fleet |

## Déploiement et compatibilité
Chaque composant possède son artefact, sa configuration, sa santé, ses logs, ses mises à jour et son rollback. La mise à jour du Control Plane ne doit pas imposer une mise à jour simultanée de tous les agents. Les versions de protocole permettent une période de compatibilité contrôlée.

## Hors cible
Kubernetes, Nomad, Temporal, une console indépendante et le multi-région actif-actif ne font pas partie de la baseline initiale. Leur ajout exige un besoin mesuré et un ADR.
