# ADR-0009 — Démarrage différé ; serveur de build Coolify en attendant

- **Statut :** Superseded en partie par [ADR-0010](0010-reprise-du-developpement.md) (point 1, développement différé). Les points 2 à 4 restent en vigueur.
- **Date :** 2026-11-16
- **Portée :** calendrier du projet et chemin de build intérimaire
- **Ne remplace pas :** ADR-0001 à ADR-0008, qui restent la cible

## Contexte
La baseline v1.0 décrit une flotte de build complète (Control Plane, agents, scheduler, isolation par
sandbox, registre, supply chain). Les phases 1 à 4 de la roadmap représentent plusieurs mois de
développement.

Situation constatée le 16 novembre 2026 :
- E-FLEX est maintenu par une seule personne et sert quelques clients sur un à trois serveurs ;
- le besoin immédiat est de **sortir les builds du serveur de production** : les builds y saturent la
  mémoire et font tomber les sites, et un build Docker n'isole pas du code client hostile ;
- Coolify, déjà utilisé comme runtime, sait construire sur un serveur de build dédié, pousser dans un
  registre et faire tirer l'image par la production ; il s'appuie sur BuildKit et prend en charge Railpack ;
- le dépôt `eflex.app` contient déjà un moteur Go (Axiom, gelé) avec agents, exécuteur et planificateur.
  Démarrer Build Fleet maintenant créerait un troisième moteur à maintenir.

## Décision
1. **Le développement de Build Fleet est différé.** Aucune phase de la roadmap n'est démarrée avant
   qu'un déclencheur ci-dessous soit constaté et documenté.
2. **Chemin intérimaire :** serveur de build Coolify dédié + registre OCI auto-hébergé, reliés à la
   production par réseau privé (Tailscale). Les applications Compose, que le serveur de build Coolify ne
   prend pas en charge, passent par un runner CI auto-hébergé sur ce même serveur.
3. **Les invariants de la baseline s'appliquent dès maintenant** au chemin intérimaire, dans la mesure
   où Coolify le permet :
   - build rattaché à un commit précis ;
   - déploiement par digest quand Coolify l'accepte, sinon tag immuable par commit ;
   - aucun secret de production ni accès de production sur le serveur de build ;
   - builds clients traités comme non fiables : serveur de build distinct du runtime, sans données clients ;
   - aucune promesse de disponibilité sans tests de panne.
4. **Un seul moteur Go à terme.** Si Build Fleet démarre, il reprend ou remplace les agents d'Axiom au
   lieu d'en ajouter un autre ; la décision fait l'objet d'un ADR.

## Déclencheurs de démarrage
Au moins un fait vérifiable, documenté par des mesures :
- plus de trois serveurs de build à piloter, ou un besoin de répartition que Coolify ne couvre pas ;
- limite du serveur de build Coolify constatée en production (concurrence, cache, files d'attente,
  Compose) et non contournable par configuration ;
- client payant exigeant une isolation de build plus forte que des serveurs dédiés ;
- coût ou temps d'exploitation du chemin intérimaire supérieur à un seuil fixé dans un ADR ;
- capacité humaine disponible pour mener au moins les phases 1 et 2 sans ralentir le produit.

## Conséquences
- Les documents d'architecture restent la cible et sont maintenus, mais ne décrivent pas un système
  existant.
- Les efforts vont au produit (E-FLEX App) et à la fiabilité du chemin intérimaire.
- Les écarts entre le chemin intérimaire et les invariants sont listés dans l'issue de suivi du
  serveur de build, pour être résorbés par Build Fleet le moment venu.

## Alternatives examinées
- **Démarrer la phase 1 maintenant** : écarté, coût élevé avant toute preuve de besoin, et aucun gain
  immédiat pour les clients actuels.
- **Abandonner Build Fleet** : écarté, l'architecture reste pertinente si l'activité grandit.
- **GitHub Actions hébergés seulement** : écarté comme solution unique, quota limité pour les dépôts
  privés ; utilisé seulement avec un runner auto-hébergé.

## Critères de réouverture
Constat d'un déclencheur ci-dessus ; ou abandon par Coolify du serveur de build ou du registre.
