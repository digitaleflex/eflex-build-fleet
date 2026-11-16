# ADR-0006 — Isolation des builds et niveaux de confiance

- **Statut :** Accepted with validation gate
- **Date :** 2026-11-16
- **Portée :** sécurité des Build Nodes

## Contexte
Les scripts et Dockerfiles clients peuvent exécuter du code arbitraire. Un conteneur standard ne suffit pas à garantir l'isolation d'un environnement multi-tenant hostile.

## Décision
- Les builds internes et clients sont classés par niveau de confiance.
- Un build client n'a pas accès aux secrets runtime.
- Les workers sont séparés du Control Plane et du Runtime Plane.
- Le mode rootless est une réduction de privilèges, pas une sandbox complète.
- Firecracker et gVisor sont les candidats à qualifier pour les builds non fiables.
- Aucun Build Node ne reçoit de build multi-tenant hostile si le niveau d'isolation requis n'est pas démontré.
- Les quotas, l'egress, le nettoyage et la séparation des caches sont obligatoires.

## Justification
La sécurité est définie par les frontières effectives du processus et les capacités du fournisseur, pas par l'étiquette Docker ou rootless.

## Validation requise
Tester KVM, compatibilité BuildKit, performance, réseau, stockage, accès aux métadonnées, séparation multi-tenant et nettoyage sur les classes de VPS ciblées.

## Conséquences
Le MVP peut être limité aux builds internes et environnements contrôlés. La disponibilité multi-tenant commerciale dépend de la qualification de la sandbox.
