# Cycle de vie d'un build

## Entrée d'un job
E-FLEX App soumet une demande contenant au minimum :
- tenant et projet autorisés ;
- dépôt et référence source ;
- commit SHA exact ou digest de snapshot ;
- répertoire racine de l'application ;
- stratégie demandée ou mode automatique ;
- cible CPU/OS ;
- profil de ressources ;
- politique réseau ;
- politique de cache ;
- références éventuelles à des secrets temporaires ;
- clé d'idempotence.

Les valeurs sensibles ne sont pas copiées en clair dans les paramètres durables. Les secrets sont représentés par des références à durée limitée.

## Machine d'états
États de référence :
- queued
- assigned
- leased
- preparing
- building
- exporting
- completed
- failed
- canceled
- expired

Chaque transition est validée par le service Job. Les transitions terminales ne sont pas réversibles. Une nouvelle exécution crée une nouvelle tentative liée au job logique initial.

## Flux nominal
1. Authentifier l'appelant et autoriser le tenant/projet.
2. Valider le contrat et la clé d'idempotence.
3. Enregistrer le job dans une transaction PostgreSQL.
4. Publier l'événement via transactional outbox.
5. Le scheduler sélectionne un nœud admissible.
6. Le job est associé à un lease avec expiration.
7. L'agent réclame le travail via sa connexion sortante.
8. L'agent vérifie l'autorisation et prépare un workspace temporaire.
9. Le commit exact est récupéré avec un credential read-only à durée limitée.
10. Le moteur sélectionné et sa version épinglée produisent le plan et l'image.
11. L'image est publiée dans un registre OCI.
12. L'agent rapporte digest, provenance, métriques, logs et statut.
13. Le Control Plane valide le résultat, ferme le lease et émet l'événement final.
14. E-FLEX App reçoit le résultat et décide du déploiement.

Build Fleet ne réalise pas le déploiement runtime.

## Lease et heartbeat
Un lease donne à un agent le droit temporaire d'exécuter une tentative précise. L'agent renouvelle le lease pendant l'exécution. Le Control Plane refuse les rapports d'une tentative qui ne possède plus le lease valide, sauf procédure explicite de réconciliation.

Un heartbeat prouve la vivacité, pas la capacité. CPU/RAM/disque disponibles et jobs actifs sont rapportés séparément.

À l'expiration du lease, le job peut être remis en file après réconciliation. Un build peut avoir terminé sans que son résultat ait été transmis. L'export doit donc être idempotent autant que possible et les artefacts identifiés par digest.

## Sémantique d'exécution
Le système garantit une livraison de travail **at-least-once**, pas une exécution exactement une fois. Les tentatives peuvent se répéter après une panne. Les opérations doivent être idempotentes et les artefacts adressés par contenu.

Aucun agent ne peut promouvoir une image en production.

## Annulation
Une annulation est une demande, pas une garantie instantanée. Le Control Plane marque la tentative cancel-requested ; l'agent arrête le processus de manière contrôlée, nettoie le workspace et rapporte le résultat. Un job déjà exporté reste un artefact non promu, sauf instruction distincte.

## Résultat minimal
Le BuildResult contient :
- job_id et attempt_id ;
- statut final et code de sortie ;
- commit source et digest du snapshot ;
- moteur, version et configuration digest ;
- image reference et image digest si succès ;
- node_id ;
- horodatages et durée ;
- référence de logs ;
- métriques de ressources ;
- SBOM/provenance disponibles ;
- avertissements et classification d'erreur.

## Classification des erreurs
Source inaccessible, détection, configuration, dépendance, build, export registry, quota, capacité, sécurité, annulation et infrastructure. La classification guide le diagnostic, mais ne déclenche pas de changement automatique de moteur.
