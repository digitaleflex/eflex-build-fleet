# Isolation des builds et ressources

## Principe
Le code client est non fiable. L'agent, le Control Plane, le réseau de gestion et le runtime ne doivent jamais être exposés directement au processus de build.

Un simple conteneur Docker n'est pas une frontière de sécurité suffisante pour des builds hostiles multi-tenant.

## Niveaux d'isolation

### Niveau A — Développement et dépôts internes
- Nœuds dédiés ou réservés aux dépôts de confiance.
- BuildKit exécuté avec les privilèges minimaux possibles.
- Workspace temporaire.
- Quotas et nettoyage systématiques.
- Aucun secret de production.

### Niveau B — Projets clients
- Nœuds et sandboxes séparés des composants de contrôle.
- Identité et secrets temporaires par build.
- Egress limité et blocage du réseau de gestion.
- Caches séparés par tenant.
- Limites CPU, mémoire, PID, disque, durée et bande passante.
- Nettoyage et audit après chaque tâche.

### Niveau C — Code hostile multi-tenant
Une isolation VM ou sandbox renforcée est requise. Firecracker et gVisor sont candidats à qualifier, pas des choix d'implémentation déjà prouvés.

Firecracker nécessite KVM disponible et une configuration correcte du Jailer, namespaces, cgroups et filtres de sécurité. gVisor possède ses propres limites de compatibilité et de performance. Les deux doivent être évalués sur les VPS réellement utilisés.

Si l'hébergeur ne permet pas le niveau de séparation requis, le nœud ne reçoit pas de builds de ce niveau de confiance.

## BuildKit
BuildKit est l'exécuteur de référence. L'agent contrôle le cycle de vie du worker BuildKit. Les endpoints BuildKit ne sont jamais exposés publiquement et ne sont accessibles qu'à l'identité et au job autorisés.

Le mode rootless réduit certains privilèges mais ne constitue pas à lui seul une sandbox complète. Les contraintes réseau et capacités du noyau doivent être examinées.

## Politique réseau de build
Le profil par défaut refuse l'accès aux réseaux de gestion, runtime, bases de données, métadonnées cloud et services internes. Les téléchargements nécessaires sont permis selon une politique contrôlée.

Les domaines autorisés, DNS, plages IP et redirections doivent être traités contre le contournement et le DNS rebinding. Les règles s'appliquent au niveau qui contient effectivement le processus de build.

## Secrets
- Aucun secret runtime disponible par défaut.
- Credentials source et registry de courte durée et à portée limitée.
- Secrets de build montés via un mécanisme prévu par le moteur, jamais ajoutés à l'image.
- Valeurs masquées dans les logs.
- Secrets exclus des caches et artefacts.
- Une tâche ne peut pas lire les secrets d'une autre tâche.

## Ressources
Chaque job possède un profil avec CPU maximum, RAM maximum, nombre maximal de processus, espace temporaire, durée, bande passante, egress, nombre de tentatives et concurrence tenant. Le profil est validé côté serveur.

## Nettoyage
Workspace, credentials temporaires, sockets et ressources de sandbox sont supprimés après succès, échec, expiration ou annulation. Le nettoyage est idempotent et repris après redémarrage.

Un nettoyage échoué place le nœud en degraded ou quarantined selon la gravité, au lieu de réutiliser aveuglément l'environnement.

## Gate de production
Avant de recevoir du code client non fiable, chaque profil d'isolation doit avoir :
- tests de séparation entre tenants ;
- tests d'accès réseau interdit ;
- tests de limitation de ressources ;
- tests d'arrêt et nettoyage ;
- tests d'évasion en environnement contrôlé ;
- preuves de configuration du noyau et hyperviseur ;
- revue indépendante des risques résiduels.
