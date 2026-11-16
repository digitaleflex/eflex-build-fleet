# Résilience, dépendances et modes dégradés

## Objectif
Continuer les opérations qui restent possibles lorsqu'un composant, un serveur ou un fournisseur devient indisponible. La résilience est définie par des comportements testables, pas par le nombre de machines.

## Classes de défaillance
- Processus : agent ou Control Plane arrêté.
- Nœud : Build Node inaccessible ou saturé.
- Réseau : partition entre agent, Control Plane et Registry.
- Dépendance : Git provider, coordination Tailscale, Registry ou stockage objet.
- Données : PostgreSQL indisponible ou corrompu.
- Fournisseur : panne d'un hébergeur ou d'une région.
- Chaîne logicielle : moteur ou dépendance compromise.

## Comportements attendus
| Défaillance | Comportement |
|---|---|
| Un Build Node tombe | Détecter après heartbeat/lease ; réconcilier ; réessayer sur un nœud admissible |
| Control Plane indisponible | Aucun nouveau job ; les agents ne prennent pas de nouveau travail ; job actif continue uniquement dans les limites du lease et de la politique locale |
| PostgreSQL indisponible | Suspendre claim et transitions durables ; ne pas accepter de nouveaux jobs comme si l'état était garanti |
| Registry indisponible | Conserver localement le résultat temporaire si possible ; réessayer l'export ; ne pas annoncer completed avant vérification |
| Git provider indisponible | Utiliser un snapshot déjà récupéré et vérifié ; sinon attendre |
| Coordination Tailscale indisponible | Connexions et politiques en cache peuvent continuer ; enrôlement, renouvellement et changement de politique peuvent être bloqués |
| Un fournisseur VPS tombe | Les autres domaines continuent seulement si Control Plane, données et artefacts restent accessibles |
| Un moteur échoue | Rapporter et diagnostiquer ; aucun fallback silencieux |
| Un agent est compromis | Révoquer son identité, le mettre en quarantaine, invalider leases et vérifier les artefacts associés |

## Control Plane et état durable
PostgreSQL est le registre de vérité des jobs. Redis ou une file volatile ne peut pas être le seul endroit où un job existe.

Le démarrage vérifie la cohérence des leases et événements. Le transactional outbox évite de commettre un job sans rendre son événement disponible.

En v1, PostgreSQL reste un domaine de panne critique. Sauvegardes hors site, alertes et restauration testée sont obligatoires. Le système ne revendique pas HA du Control Plane avant la mise en place d'un mécanisme de bascule validé.

## Agent pendant une partition
L'agent ne crée pas de travail et ne renouvelle pas une autorisation expirée. Un job déjà accepté peut continuer uniquement dans les limites de son lease et de sa politique locale. L'agent conserve un spool borné pour les logs et résultats non acquittés.

À la reconnexion :
- l'agent présente l'état du job et son lease ;
- le Control Plane réconcilie avec l'état durable ;
- les artefacts sont vérifiés par digest ;
- les résultats obsolètes sont refusés ou marqués non promouvables ;
- les travaux expirés sont réessayés de façon idempotente.

## Multi-provider réel
Le multi-provider exige :
- Control Plane et base non exclusivement liés au même domaine de panne que tous les agents ;
- Build Nodes sur au moins deux fournisseurs indépendants ;
- chemins registry et stockage testés ;
- sauvegarde indépendante ;
- procédure de remplacement et enrôlement ;
- exercice de perte d'un fournisseur.

## RPO et RTO
Aucune valeur RPO/RTO n'est fixée dans cette baseline. Elles seront définies selon le coût métier et validées par tests. Les valeurs annoncées au client ne peuvent pas dépasser les résultats démontrés.

## Tests de panne requis
- kill d'un agent pendant un build ;
- arrêt d'un Build Node pendant l'export ;
- perte temporaire de connectivité Tailscale ;
- interruption PostgreSQL ;
- interruption Registry ;
- saturation disque et RAM ;
- expiration d'un credential ;
- perte d'un fournisseur de build ;
- redémarrage du Control Plane pendant une transition de job.
