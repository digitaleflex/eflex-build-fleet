# Politique de décision et de gouvernance

## Objectif
Empêcher les réorientations techniques répétées et préserver une mémoire de décision durable. Ce document définit comment une décision d'architecture est prise, enregistrée, modifiée et retirée.

## Source de vérité
1. Les invariants et ADR acceptés définissent l'architecture de référence.
2. Les documents d'architecture expliquent comment appliquer ces décisions.
3. La roadmap définit l'ordre de réalisation et les preuves attendues.
4. Le code prouve ce qui est implémenté ; il ne modifie pas silencieusement l'architecture.
5. Une conversation, une issue ou une préférence ne remplace pas un ADR accepté.

En cas de contradiction, ouvrir une issue de documentation et corriger le référentiel avant de propager un nouveau choix dans le code.

## Catégories de décision
- **Accepted** : choix retenu pour la baseline ; à appliquer.
- **Accepted with validation gate** : direction arrêtée, qualification requise avant production.
- **Deferred** : décision volontairement reportée ; ne pas l'introduire avant son déclencheur.
- **Rejected** : option non retenue pour le périmètre indiqué.
- **Superseded** : remplacée par un ADR plus récent.

Une décision acceptée n'est pas une preuve d'implémentation.

## Quand rouvrir une décision
Une décision acceptée ne peut être réouverte que sur la base d'un fait vérifiable :
1. Vulnérabilité critique, faiblesse de sécurité ou exigence réglementaire incompatible.
2. Projet abandonné, non maintenu ou licence devenue incompatible avec l'usage prévu.
3. Échec reproductible d'un critère d'acceptation.
4. Limite technique mesurée qui empêche une capacité requise.
5. Coût total d'exploitation mesuré dépassant un seuil économique documenté.
6. Indisponibilité récurrente ou risque fournisseur non maîtrisable.
7. Besoin produit/client confirmé, impossible à traiter par extension de l'architecture existante.

Une nouvelle technologie populaire, une préférence de langage ou une possibilité théorique ne suffit pas.

## Processus obligatoire
Toute proposition doit fournir :
- problème concret et preuves ;
- décision actuelle et conséquences de son maintien ;
- alternatives, y compris « ne rien changer » ;
- résultats de benchmarks ou tests ;
- impact sécurité, licences, coûts et exploitation ;
- migration, compatibilité et retour arrière ;
- fichiers/ADR à remplacer ;
- critères d'acceptation.

Un changement accepté crée un nouvel ADR. L'ancien devient Superseded et conserve son historique.

## Revue des dépendances
Toute dépendance structurante possède une fiche de qualification : dépôt officiel, version épinglée, licence, politique de sécurité, cadence de publication, procédure de mise à jour, compatibilité, alternative et coût d'exploitation.

Les versions de moteurs, bases images, buildpacks et dépendances système sont épinglées. Une mise à jour est un changement contrôlé.

## Responsabilités
- Architecture : maintenir les ADR et les frontières.
- Sécurité : valider les frontières de confiance et scénarios de menace.
- Exploitation : valider le déploiement, les sauvegardes et les restaurations.
- Produit : confirmer besoin client et valeur économique.
- Mainteneur : ne pas fusionner une modification contradictoire sans ADR de remplacement.

## Règle de stabilité
La baseline v1.0 est le point de départ commun. Chercher d'abord à réaliser le besoin dans les frontières existantes. Toute nouvelle abstraction, service, base, protocole ou langage doit justifier sa nécessité par une preuve et un ADR.
