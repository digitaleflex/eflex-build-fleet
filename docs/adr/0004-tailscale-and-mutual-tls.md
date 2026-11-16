# ADR-0004 — Tailscale pour le réseau privé, mTLS pour l'identité

- **Statut :** Accepted with validation gate
- **Date :** 2026-11-16
- **Portée :** communications internes

## Contexte
Les Build Nodes pourront être situés chez différents fournisseurs. L'administration publique de chaque nœud augmenterait la surface d'exposition.

## Décision
- Tailscale est le réseau privé de gestion initial.
- Les agents initient les connexions de gestion vers le Control Plane.
- Les appels applicatifs utilisent gRPC avec mTLS.
- Tailscale ne remplace ni l'identité applicative ni l'autorisation par job.
- La politique réseau est restrictive par défaut.
- Les sandboxes ne reçoivent pas l'accès au tailnet de l'hôte.
- Le protocole applicatif reste utilisable sur une autre couche réseau compatible.

## Justification
Tailscale simplifie la connectivité multi-hébergeur, tandis que mTLS maintient une identité et une autorisation indépendantes du réseau privé.

## Validation requise
ACL/grants minimaux ; cycle de vie des clés et certificats ; révocation d'un agent ; comportement lors d'une panne de coordination ; filtrage depuis la sandbox ; procédure SSH d'urgence.

## Conséquences
Tailscale reste une dépendance de coordination en v1. Les sessions et politiques déjà établies peuvent continuer pendant une panne de coordination, mais les nouveaux enrôlements et renouvellements peuvent être bloqués.

## Alternative future
Headscale ou WireGuard auto-administré si l'indépendance du plan de contrôle réseau devient une exigence mesurée.
