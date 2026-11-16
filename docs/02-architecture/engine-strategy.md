# Stratégie des moteurs de build

## Distinction des rôles
- **Build planner / auto-detection** : analyse le dépôt et détermine les étapes.
- **Build executor** : exécute les instructions et construit les couches.
- **Buildpacks lifecycle** : détecte, construit et exporte selon le modèle Cloud Native Buildpacks.
- **Build definition** : Dockerfile ou commandes explicites.
- **Runtime topology** : Docker Compose décrit plusieurs services ; ce n'est pas un moteur de compilation.

## BuildKit — exécuteur de référence
BuildKit est le backend d'exécution principal. Il fournit le graphe de build, l'exécution parallèle, le cache, les exports et la prise en charge de Dockerfile/frontends.

E-FLEX ne réécrit pas BuildKit. L'intégration est encapsulée derrière une interface interne versionnée. Les versions et la configuration des workers sont épinglées et qualifiées.

## Railpack — stratégie automatique principale
Railpack est le premier moteur de détection et de génération automatique de Build Plan. Il produit un plan exploitable par un backend BuildKit.

La version étudiée à la date de baseline est **v0.43.0 (19 octobre 2026)**. Le projet étant encore en version 0.x, aucune montée de version automatique n'est autorisée. Une version exacte est épinglée dans chaque release E-FLEX et enregistrée dans les métadonnées du build.

Railpack est utilisé via un adaptateur : E-FLEX contrôle politiques, sélection, paramètres autorisés, version et exécution.

## Dockerfile — chemin explicite fondamental
Un Dockerfile fourni par le projet est une définition explicite de build, exécutée par BuildKit dans le même cadre de sécurité et de traçabilité. Son contenu est non fiable.

## Nixpacks — compatibilité, pas stratégie
Nixpacks est en maintenance et son projet recommande Railpack. Il reste disponible pour les projets historiques qui l'exigent. Aucun fork principal E-FLEX n'est prévu.

Une demande Nixpacks sélectionne une version épinglée et une voie qualifiée. Aucun fallback silencieux vers Railpack n'est autorisé.

## Cloud Native Buildpacks — extension différée
CNB est un écosystème standardisé, mais builders, buildpacks, lifecycle, ordre et compatibilités ajoutent une charge. L'intégration est différée jusqu'à un besoin client démontré et un benchmark comparatif.

Les licences sont vérifiées composant par composant : lifecycle, builder et chaque buildpack.

## Outils spécialisés
ko peut être étudié pour Go ; Jib pour Java. Ils ne sont ajoutés que si un cas d'usage réel démontre un gain.

## Politique de sélection
Le client choisit initialement : automatique, Dockerfile, ou stratégie avancée explicitement configurée. En mode automatique, la plateforme détermine le moteur exact selon des règles versionnées.

Chaque BuildRequest conserve strategy_id, engine_id, engine_version, configuration_digest, frontend version et digests des images de base disponibles.

## Aucun fallback silencieux
Un moteur qui échoue ne déclenche pas automatiquement un autre moteur. Un changement peut modifier les dépendances, commandes, couches et permissions. Un second moteur ne peut être utilisé que sur une nouvelle tentative explicitement demandée ou une politique préalablement approuvée et traçable.

## Docker Compose
Compose définit une topologie multi-service ; il ne remplace pas le pipeline de build.
1. Analyser les services et contextes.
2. Construire chaque image déclarée comme buildable.
3. Publier chaque image par digest.
4. Produire ou valider un Compose de déploiement qui référence les digests.
5. Transmettre artefacts et topologie à E-FLEX App / Coolify.

La réécriture d'un Compose client est limitée, explicite et testée. Les services ne sont pas silencieusement modifiés.

## Extensions et forks
Ordre de préférence : configuration officielle, extension/provider officiel, adaptateur E-FLEX, patch amont, fork en dernier recours.

Chaque fork exige un propriétaire, une politique de synchronisation amont, des tests, une publication, un suivi de sécurité et un coût de maintenance justifié.

## Qualification avant production
Vérifier versions et licences, compatibilité CPU/OS, réseau et secrets, reproductibilité du plan, tests sur applications représentatives, cache, performances, diagnostics et rollback de version.
