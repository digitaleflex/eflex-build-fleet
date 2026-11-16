# Invariants du système

Ces règles s'appliquent à tous les composants, environnements et contributions.

## Produit et frontières
1. Build Fleet compile et publie des artefacts ; il ne gère pas le runtime des applications clientes.
2. E-FLEX App reste l'interface produit et le point d'entrée client.
3. Build Fleet possède son propre état et son propre cycle de livraison.
4. Les services ne partagent pas directement leurs bases de données.
5. Les communications inter-systèmes passent par des contrats versionnés.

## Jobs
6. Chaque job cible un commit source exact, jamais uniquement une branche mobile.
7. Chaque tentative est identifiée et auditée séparément.
8. Un lease expiré peut entraîner une nouvelle tentative ; les effets doivent être idempotents.
9. Un job n'est « completed » qu'après publication vérifiée de l'artefact attendu.
10. Un build échoué ne remplace jamais automatiquement une image saine déjà déployée.
11. Aucun moteur différent n'est sélectionné silencieusement après un échec.

## Agents et réseau
12. Un Build Agent ne possède pas d'accès libre aux serveurs runtime.
13. Les agents initient les connexions de gestion vers le Control Plane.
14. Chaque agent et chaque commande sensible sont authentifiés et autorisés au niveau applicatif.
15. La disparition du réseau de gestion n'autorise pas un agent à accepter de nouvelles tâches.
16. Les clés, certificats et jetons ont une portée et une durée limitées.

## Builds non fiables
17. Tout code source, script, Dockerfile, dépendance et paramètre client est non fiable.
18. Aucun secret runtime n'est injecté dans un build par défaut.
19. Les caches mutables sont séparés par tenant et niveau de confiance.
20. Les ressources, le réseau, la durée et le disque sont bornés.
21. Un conteneur seul n'est pas une frontière de sécurité suffisante pour du code multi-tenant hostile.

## Artefacts
22. L'artefact de référence est identifié par son digest OCI.
23. Les tags sont des pointeurs mutables, pas des identifiants immuables.
24. Le déploiement consomme le digest produit et vérifié.
25. Les métadonnées de provenance sont générées par la plateforme de confiance.

## Exploitation
26. Toute fonction critique possède des métriques, des logs corrélables et un runbook.
27. Toute promesse de disponibilité ou récupération doit être prouvée par des tests.
28. Les sauvegardes ne sont valides qu'après un test de restauration.
29. L'absence de configuration ou de preuve signifie que la capacité n'est pas garantie.
