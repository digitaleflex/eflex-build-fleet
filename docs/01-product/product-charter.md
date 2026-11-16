# Charte produit — E-FLEX Build Fleet

## Vision
Construire l'infrastructure de compilation distribuée propre à E-FLEX, capable de transformer des sources applicatives en artefacts vérifiables sur une flotte de serveurs indépendants, sans coupler les compilations aux machines qui hébergent les applications clientes.

## Mission
Fournir à E-FLEX App une capacité de build fiable, sécurisée, mesurable et indépendante des fournisseurs d'hébergement. Donner à E-FLEX le contrôle de la planification, de l'exécution, des ressources, des caches, des artefacts et de la récupération après incident.

## Problème
Lorsque le build et le runtime partagent une même machine ou un même fournisseur, une compilation lourde peut affecter les applications déployées. Une dépendance à un service ou une machine unique crée un point de défaillance.

Build Fleet isole le Build Plane du Runtime Plane et répartit les tâches entre des nœuds possédant des capacités et des domaines de panne connus.

## Promesse produit
- Une demande de build est acceptée, planifiée et suivie.
- Un commit précis est compilé avec un moteur et des versions identifiés.
- Un agent exécute le build avec des ressources et permissions limitées.
- Un artefact immuable est publié dans un registre OCI.
- E-FLEX App reçoit un résultat vérifiable et peut demander le déploiement.
- Une panne locale n'entraîne pas automatiquement la perte de l'ensemble de la flotte.

Cette promesse est une cible architecturale. Elle ne constitue pas une garantie SLA avant mesure de la disponibilité, du RPO, du RTO et de la charge.

## Utilisateurs et consommateurs
- E-FLEX App : client principal de l'API.
- Opérateurs E-FLEX : gestion de flotte, quotas, incidents et politiques.
- Build Agents : exécutants techniques.
- Développeurs clients : utilisateurs indirects via E-FLEX App.

Aucune console client séparée n'est nécessaire dans le MVP.

## Périmètre inclus
- Enregistrement, identité et état des Build Nodes.
- Scheduler, jobs, leases et tentatives.
- Agents de build indépendants.
- Contrats de jobs et résultats.
- Adaptateurs de moteurs.
- Profils de ressources et quotas multi-tenant.
- Gestion des workspaces et nettoyage.
- Publication OCI par digest.
- Logs, événements, métriques et audit.
- Politiques de sécurité et récupération.

## Hors périmètre
- Hébergement et supervision runtime des applications.
- Domaines, DNS, TLS ou reverse proxy runtime.
- Orchestration générale de conteneurs.
- Gestion des bases de données clients.
- Réécriture d'un moteur universel remplaçant BuildKit.
- Migration universelle depuis Vercel, Railway, Render ou Netlify.
- Facturation finale ; Build Fleet fournit des mesures à E-FLEX App.
- Multi-région active-active dans la première version.

## Principes économiques
- Ajouter un serveur lorsque la capacité ou la résilience attendue le justifie.
- Mesurer le coût par build réussi, pas seulement le coût horaire des VPS.
- Distinguer attente, compilation, cache hit, CPU, RAM, disque et transfert.
- Ne pas sacrifier l'isolation pour augmenter artificiellement le débit.
- Optimiser à partir de mesures reproductibles.

## Définition de réussite
Les builds sont indépendants des serveurs runtime ; la flotte supporte la perte d'un nœud ; les artefacts sont identifiés et vérifiables ; les tenants sont isolés ; le coût d'exploitation par build est connu.

## Dépendances assumées
Le système peut dépendre d'un fournisseur Git, d'un hébergeur, d'un service de coordination réseau, d'un registre et d'un stockage objet. Chaque dépendance doit avoir un périmètre, un remplacement possible et un comportement documenté en cas d'indisponibilité. L'objectif est de rendre ces dépendances remplaçables, pas de prétendre qu'elles n'existent plus.
