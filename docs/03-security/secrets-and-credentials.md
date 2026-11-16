# Secrets, credentials et accès temporaires

## Principe
Le Control Plane autorise l'accès aux secrets ; l'agent ne reçoit que la valeur nécessaire, pour une tâche précise et pendant une durée limitée. Un secret n'est jamais une donnée ordinaire du BuildRequest.

## Catégories
- Source credentials : accès read-only au dépôt et au commit autorisé.
- Registry credentials : push vers le repository/namespace autorisé, pull selon besoin.
- Build secrets : variables/jetons explicitement requis par une étape de build.
- Agent identity : certificats de service propres au nœud.
- Control Plane secrets : clés de signature, de chiffrement et d'intégration.

Les secrets runtime des applications clientes ne sont pas disponibles au Build Plane par défaut.

## Cycle de vie
1. E-FLEX App demande un build pour un projet autorisé.
2. Le Control Plane détermine les droits et références nécessaires.
3. Un credential temporaire est émis ou récupéré depuis un gestionnaire de secrets.
4. Le credential est transmis au canal agent authentifié, avec portée et expiration.
5. L'agent le monte via un mécanisme dédié, jamais dans un argument de commande visible ou un Dockerfile.
6. Le credential est révoqué/expiré et l'espace de travail est nettoyé.
7. Les logs, caches et artefacts sont inspectés pour éviter toute fuite.

## Interdictions
- Pas de secret dans BuildRequest en clair.
- Pas de secret dans Docker ARG ou ENV par défaut.
- Pas de secret inclus dans une couche, un cache exporté, un log ou une provenance.
- Pas de token GitHub de longue durée sur un Build Node.
- Pas de credential registry permettant de supprimer ou administrer tout le registre.
- Pas d'accès runtime par défaut.
- Pas de secret partagé entre tenants.

## Chiffrement et gestion de clés
Les secrets persistés doivent être chiffrés au repos via chiffrement enveloppe. Les clés maîtresses doivent être séparées de PostgreSQL et sauvegardées selon un processus distinct. Le fournisseur concret de KMS/gestion de clés reste à choisir lors de la phase d'implémentation ; ce choix ne doit pas être remplacé par une clé statique stockée dans le dépôt ou dans la même base.

## Rotation et révocation
Les identités d'agents, clés de signature, credentials Git et registry possèdent une procédure de renouvellement et révocation. Lorsqu'un agent est compromis, révoquer son identité avant de le réinscrire.

## Masquage et audit
Les logs masquent les valeurs secrètes, y compris encodées ou tronquées lorsque détectables. Les événements d'audit enregistrent l'identifiant, le scope et le résultat d'une opération, jamais la valeur du secret.

## Validation
Tests obligatoires : secret absent des logs, image, cache et provenance ; expiration ; révocation ; scope insuffisant ; accès croisé tenant refusé ; agent compromis révoqué.
