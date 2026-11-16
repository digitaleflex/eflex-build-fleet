# Flotte et scheduler

## Objectif
Distribuer les jobs aux agents capables de les exécuter, en tenant compte de la compatibilité, capacité, quotas, équité et domaines de panne. Le scheduler est un module interne du Control Plane Go, pas un orchestrateur généraliste.

## Enregistrement d'un nœud
Chaque Build Node possède un identifiant stable et une identité cryptographique. L'agent s'enregistre et publie ses capacités :
- architecture CPU et système ;
- version agent et protocole ;
- moteurs et versions disponibles ;
- capacité CPU, mémoire et disque ;
- concurrence maximale configurée ;
- jobs actifs et profils de ressources ;
- stockage temporaire disponible ;
- labels de région, fournisseur et domaine de panne ;
- état de santé et dernier heartbeat.

Les capacités déclarées sont validées autant que possible, mais ne sont pas une autorité de sécurité.

## États d'un nœud
- provisioning
- enrolling
- ready
- draining
- degraded
- offline
- quarantined
- retired

Un nœud draining ne reçoit plus de nouveau job, mais peut terminer les jobs actifs selon leur deadline. Un nœud quarantined est exclu du placement jusqu'à enquête.

## Placement en deux étapes
1. **Filtrage d'admissibilité** : architecture, moteur/version, espace disque, mémoire minimale, CPU disponible, politique de confiance, région autorisée, quota tenant et santé.
2. **Classement** : coût estimé, charge, attente, cache local, diversité de fournisseurs et répartition.

Un nœud non admissible n'est jamais choisi, quel que soit son score.

Le scheduler doit être déterministe à paramètres identiques autant que possible. Les facteurs de score sont observables pour faciliter le diagnostic.

## Équité et quotas
La capacité disponible est réservée de manière atomique. Le scheduler doit éviter qu'un tenant monopolise la flotte. La politique initiale utilise des quotas par tenant et une limite de concurrence par nœud.

Les priorités ont des bornes et ne contournent pas les quotas de sécurité. Les priorités internes sont auditables.

## Capacité et réservation
Un job ne peut être attribué que si les ressources demandées sont disponibles. La réservation est atomique afin d'éviter que deux jobs consomment la même capacité déclarée. Le disque temporaire et la taille du workspace sont pris en compte.

## Perte d'un nœud
Un heartbeat manquant déclenche un état suspect, puis offline après une fenêtre configurable. Les leases sont réconciliés avant réattribution. Le système distingue nœud mort, agent arrêté, réseau indisponible, job toujours actif mais non joignable et export terminé sans acquittement.

Un nœud suspect n'est pas immédiatement déclaré perdu : cela évite les doubles tentatives lors d'une courte partition réseau.

## Multi-provider
Les nœuds portent les identifiants de fournisseur et de domaine de panne. Le scheduler peut imposer une distribution minimale entre domaines lorsque la flotte le permet.

Le multi-provider n'est déclaré opérationnel que si au moins deux fournisseurs indépendants sont utilisés, que les données/artefacts/sauvegardes restent accessibles et qu'un test de perte d'un fournisseur réussit.

## Autoscaling
L'autoscaling de machines est différé. La v1 gère l'inventaire et le placement des nœuds provisionnés. Le provisionnement automatique nécessitera des adaptateurs fournisseurs, des limites de dépenses et une procédure de destruction sécurisée.

## Pas de Kubernetes au départ
La flotte est spécialisée pour les tâches de compilation. Un scheduler métier limite la surface opérationnelle. Kubernetes peut être étudié si le placement ou l'élasticité dépassent les capacités mesurées du scheduler.
