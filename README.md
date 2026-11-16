# E-FLEX Build Fleet

> **Infrastructure de compilation distribuée, contrôlée et indépendante du runtime.**

E-FLEX Build Fleet est le système interne d'E-FLEX chargé de planifier, exécuter, sécuriser et distribuer les compilations applicatives sur une flotte de serveurs indépendants.

Il ne remplace pas BuildKit, Railpack ou les standards OCI. Il apporte la couche que nous possédons : Control Plane, scheduler, agents, contrats, politiques multi-tenant, contrôle des ressources, isolation, résilience, traçabilité et intégration avec E-FLEX App.

## Statut de référence

**Architecture baseline v1.0 — approuvée le 16 novembre 2026.**

Cette baseline fixe les décisions de conception à appliquer. Elle ne signifie pas que toutes les fonctions sont déjà implémentées ou qualifiées en production. Les ADR indiquent les choix acceptés ; la roadmap définit les preuves requises avant chaque mise en service.

Toute évolution d'une décision acceptée doit suivre la [politique de décision](docs/00-governance/decision-policy.md) et produire un ADR motivé. Les décisions ne sont pas réouvertes sur la seule base d'une préférence ou d'une nouvelle tendance.

## Parcours de lecture

### 1. Produit et gouvernance
- [Charte produit et périmètre](docs/01-product/product-charter.md)
- [Politique de décision et gouvernance](docs/00-governance/decision-policy.md)
- [Invariants non négociables](docs/00-governance/system-invariants.md)
- [Contribution au projet](CONTRIBUTING.md)

### 2. Architecture
- [Architecture de référence](docs/02-architecture/reference-architecture.md)
- [Cycle de vie d'un build](docs/02-architecture/build-lifecycle.md)
- [Contrats API et messages](docs/02-architecture/api-contracts.md)
- [Modèle de persistance logique](docs/02-architecture/persistence-model.md)
- [Flotte et scheduler](docs/02-architecture/fleet-scheduler.md)
- [Stratégie des moteurs](docs/02-architecture/engine-strategy.md)
- [Artefacts, registre OCI et cache](docs/02-architecture/artifacts-and-registry.md)
- [Réseau privé et identité](docs/02-architecture/network-and-identity.md)

### 3. Sécurité et fiabilité
- [Modèle de menace](docs/03-security/threat-model.md)
- [Isolation des builds](docs/03-security/build-isolation.md)
- [Secrets et credentials](docs/03-security/secrets-and-credentials.md)
- [Résilience et modes dégradés](docs/04-reliability/failure-model.md)
- [Sauvegarde et restauration](docs/04-reliability/backup-and-restore.md)

### 4. Exploitation et évolution
- [Provisionnement des Build Nodes](docs/05-operations/build-node-provisioning.md)
- [Observabilité et exploitation](docs/05-operations/observability-and-operations.md)
- [Roadmap et critères de sortie](docs/06-roadmap/roadmap.md)
- [Index des ADR](docs/adr/README.md)
- [Références techniques officielles](docs/references/technical-references.md)

## Décisions acceptées

| Domaine | Décision de référence |
|---|---|
| Langage du cœur | Go |
| Forme initiale | Monolithe modulaire Control Plane + binaire Build Agent indépendant |
| Orchestration | Scheduler E-FLEX spécialisé |
| Exécution d'images | BuildKit |
| Build automatique | Railpack, version épinglée et qualifiée |
| Compatibilité historique | Nixpacks, sans fork principal |
| Queue et état | PostgreSQL, leases et transactional outbox |
| Réseau privé initial | Tailscale, avec mTLS et autorisation applicative obligatoires |
| Distribution | OCI, images identifiées par digest |
| Registre auto-hébergé | Zot comme premier candidat à qualifier derrière une abstraction OCI |
| Interface client | E-FLEX App ; pas de deuxième console dans le MVP |
| Runtime | Hors périmètre ; déploiement délégué à E-FLEX App / Coolify |

Les décisions détaillées et leurs alternatives sont enregistrées dans [docs/adr](docs/adr/).

## Invariants non négociables

- Le Build Plane est séparé du Runtime Plane.
- Un build est associé à un commit source immuable et à une version exacte de son moteur.
- Un agent ne reçoit jamais les privilèges d'un serveur de production.
- Le Control Plane conserve l'état durable des jobs.
- Les agents initient les connexions vers le Control Plane.
- Tailscale ne remplace pas mTLS ni l'autorisation applicative.
- Les images sont publiées et déployées par digest OCI, pas par tag mutable.
- Les builds clients sont considérés comme du code non fiable.
- Les caches et secrets sont isolés par tenant et par niveau de confiance.
- Aucune promesse de haute disponibilité n'est faite avant des tests de panne et de restauration.

## Ce que le projet ne cherche pas à faire

- Réécrire BuildKit ou créer un compilateur universel.
- Devenir un orchestrateur généraliste de conteneurs.
- Remplacer Kubernetes, Nomad ou Coolify dans tous leurs usages.
- Héberger les applications clientes ou gérer leur cycle de vie runtime.
- Ajouter une console indépendante qui duplique E-FLEX App.
- Déployer immédiatement une architecture multi-région complexe.

## Relation avec E-FLEX App

E-FLEX App est le produit et l'interface client. Build Fleet est un service d'infrastructure autonome. Ils communiquent via une API versionnée et des événements documentés. Ils ne partagent pas leur base de données et ne doivent pas être déployés comme une seule unité.

Voir [ADR-0007 — frontière produit / Build Fleet](docs/adr/0007-product-and-runtime-boundaries.md).

## Licence et réutilisation

Le dépôt ne contient pas encore de fichier de licence. Sa visibilité publique ne constitue pas une autorisation de réutilisation ou de redistribution. La licence du code E-FLEX doit faire l'objet d'une décision explicite. Les licences des dépendances tierces sont vérifiées séparément avant intégration ou redistribution.
