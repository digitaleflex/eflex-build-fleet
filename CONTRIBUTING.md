# Contribuer à E-FLEX Build Fleet

## Avant toute modification
Lire dans cet ordre :
1. README.md
2. docs/00-governance/decision-policy.md
3. docs/00-governance/system-invariants.md
4. Architecture et ADR concernés
5. Roadmap

## Règles obligatoires
- Ne pas introduire un langage, service, base, broker, framework ou moteur sans justification.
- Ne pas contredire un ADR Accepted sans ADR de remplacement.
- Ne jamais ajouter de secret, credential ou fichier de configuration réel.
- Épingler les versions des moteurs et dépendances structurantes.
- Vérifier les licences de chaque composant et dépendance redistribuée.
- Conserver les contrats indépendants des modèles SQL.
- Ne pas exposer un endpoint BuildKit ou agent publiquement.
- Ne pas ajouter d'accès runtime à un Build Agent.
- Ne jamais partager de cache d'écriture entre tenants.
- Toute fonction de sécurité ou reprise doit être accompagnée de tests.

## Changements d'architecture
Une modification de frontière, protocole, moteur par défaut, stockage durable, réseau, identité, isolation ou contrat inter-service exige un ADR.

Les changements documentaires sont réalisés dans le même changement que le code lorsqu'ils décrivent un comportement modifié.

## Tests attendus
Les tests sont ajoutés au niveau approprié : unitaires, intégration, contract tests, tests de panne, sécurité, compatibilité moteur et benchmarks. Les tests de panne ou de sécurité ne doivent pas exécuter de code hostile hors d'un environnement contrôlé.

## Pull requests
Une PR décrit :
- le problème ;
- la décision ou ADR concerné ;
- les changements ;
- les risques sécurité et exploitation ;
- les tests exécutés ;
- les limites connues ;
- le plan de retour arrière.

## Qualité
Les erreurs sont explicites, les logs structurés et dépourvus de secrets, les opérations importantes idempotentes, les états persistés et les permissions minimales.
