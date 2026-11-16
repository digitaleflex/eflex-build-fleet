# Artefacts, registre OCI et cache

## Principe
Un build réussi produit un ou plusieurs artefacts immuables. L'image est identifiée par son digest OCI. Les tags sont des alias mutables et ne sont pas la référence de déploiement.

**Build once, deploy the immutable digest.**

## Abstraction de registre
Le Control Plane et les agents utilisent une interface interne RegistryProvider fondée sur les opérations OCI standard. Le cœur métier ne dépend pas des API privées d'un fournisseur.

Capacités attendues : push/pull, lecture de manifest et digest, authentification limitée, vérification de présence, suppression contrôlée, réplication, rétention et stockage des métadonnées.

## Zot
Zot est le premier candidat de registre auto-hébergé à qualifier. Il est retenu comme cible de prototype, pas comme dépendance irréversible.

Avant production, vérifier compatibilité BuildKit/runtime, authentification, ACL, stockage, sauvegarde/restauration, réplication, rétention, nettoyage, métriques et capacité. Si Zot ne satisfait pas ces critères, un autre registre OCI peut être choisi sans modifier le contrat RegistryProvider.

## Réplication et résilience
La réplication ne signifie pas qu'un registre secondaire est automatiquement prêt. Les politiques définissent les artefacts à répliquer, la source de vérité, le délai acceptable, la vérification des digests, les suppressions et la restauration.

Un déploiement ne peut utiliser un artefact secondaire qu'après vérification du digest.

## Cache
Le cache accélère les builds mais n'est pas un artefact de référence.

Règles :
- cache mutable isolé par tenant et niveau de confiance ;
- aucun partage de cache d'écriture entre tenants ;
- caches de base images éventuellement partagés en lecture seule après vérification ;
- cache externe considéré comme non fiable ;
- quota, expiration et nettoyage ;
- mesure du hit rate et du coût ;
- purge d'un cache sans affecter les images publiées.

Le cache ne contourne jamais les contrôles d'accès aux secrets ou dépendances.

## Provenance
Le résultat relie le digest de sortie au dépôt, commit, snapshot, stratégie et version du moteur, paramètres internes, dépendances résolues connues, identité du builder, durée, résultat, SBOM et attestations.

La provenance est produite par la plateforme de confiance à partir de données validées. Un script client ne peut pas déclarer lui-même une provenance fiable.

## Promotion runtime
E-FLEX App reçoit le digest et les métadonnées. Le déploiement runtime est une action distincte, autorisée et auditée. Un build terminé n'implique pas une promotion.
