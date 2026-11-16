# ADR-0005 — Distribution OCI et qualification de Zot

- **Statut :** Accepted with validation gate
- **Date :** 2026-11-16
- **Portée :** stockage et distribution d'artefacts

## Contexte
Les images doivent être séparées des nœuds de build et des runtimes. Un registre propriétaire unique créerait un verrou opérationnel.

## Décision
- OCI est le contrat d'image et de distribution.
- L'image de référence est identifiée par digest.
- Les consommateurs passent par une interface RegistryProvider.
- Zot est le premier candidat pour un registre auto-hébergé.
- Un registre existant peut être utilisé pendant le développement initial.
- Le modèle doit supporter un registre alternatif ou un miroir.
- Les caches ne sont pas assimilés aux artefacts publiés.

## Validation requise
Tester compatibilité BuildKit/runtime, authentification, ACL, stockage, backup/restore, réplication, rétention, métriques et charge. Le choix opérationnel de Zot est confirmé après ces tests.

## Conséquences
Nous possédons l'intégration et la politique de distribution, pas le protocole de registre. Le produit registre peut évoluer sans modifier les contrats métier.
