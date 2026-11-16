# Modèle de menace

## Périmètre
Le modèle couvre Control Plane, Build Agents, Build Nodes, sandboxes, sources Git, moteurs de build, caches, registry, réseau de gestion et intégration E-FLEX App.

Un dépôt client, Dockerfile, script, dépendance ou paramètre fourni par un client peut être hostile.

## Actifs à protéger
- Secrets d'organisation, tokens Git et credentials registry.
- Clés et certificats des agents.
- Base PostgreSQL et historique des jobs.
- Images, digests, SBOM et provenance.
- Données et caches de chaque tenant.
- Réseau privé, Control Plane et Runtime Plane.
- Disponibilité et capacité de calcul.
- Journaux et traces d'audit.

## Acteurs et scénarios
| Acteur / menace | Objectif | Conséquence |
|---|---|---|
| Dépôt client malveillant | Exécuter du code hostile | Compromission worker ou exfiltration |
| Dépendance compromise | Injecter du code à l'installation | Image compromise |
| Tenant malveillant | Lire un autre cache ou dépasser les quotas | Fuite ou déni de service |
| Agent compromis | Falsifier un résultat ou accéder au réseau | Build frauduleux, mouvement latéral |
| Control Plane compromis | Distribuer des jobs ou credentials abusifs | Compromission de flotte |
| Registry compromis | Remplacer une image ou manifest | Déploiement altéré |
| Fournisseur indisponible | Interrompre réseau, stockage ou machines | Perte temporaire de capacité |
| Mainteneur compromis | Publier une version malveillante | Compromission supply chain |
| Erreur d'exploitation | Mauvaise ACL, secret ou suppression | Exposition ou perte de données |

## Frontières de confiance
1. Client / E-FLEX App.
2. E-FLEX App / Control Plane.
3. Control Plane / Build Agent.
4. Build Agent / sandbox.
5. Sandbox / réseau externe et services internes.
6. Build Agent / Registry.
7. Build Plane / Runtime Plane.
8. Dépendances amont / artefacts promus.

Chaque frontière possède authentification, autorisation, validation, journalisation et limites adaptées.

## Mesures obligatoires
- Authentification mutuelle et autorisation par job/tenant.
- Identités d'agents révocables et rotation.
- Sandbox isolée et workspace éphémère.
- Pas de Docker socket de production.
- Pas de secret runtime dans le build.
- Egress restrictif et blocage des réseaux internes / métadonnées.
- Quotas CPU, RAM, PID, disque, durée, bande passante et concurrence.
- Cache séparé par tenant et niveau de confiance.
- Images par digest et provenance signée.
- SBOM et scan de vulnérabilités.
- Logs avec masquage des secrets.
- Audit des actions admin, jobs, politiques et versions.
- Versions de moteurs et bases images épinglées et qualifiées.

## Risques résiduels
- Une vulnérabilité noyau ou hyperviseur peut traverser une frontière d'isolation.
- Une compromission du Control Plane affecte potentiellement la flotte.
- Une clé de signature compromise permet des attestations frauduleuses.
- Un fournisseur commun peut concentrer des domaines de panne.
- Le digest de sortie ne garantit pas à lui seul la reproductibilité du build.
- Le réseau Tailscale dépend d'un service de coordination tiers.
- Les dépendances téléchargées peuvent changer si elles ne sont pas verrouillées.

## Conditions de mise en production
Les builds de dépôts clients non fiables ne sont pas autorisés sur un worker partagé non qualifié. Le niveau d'isolation, l'egress, les limites et les tests d'évasion doivent être validés avant ouverture multi-tenant.

Un test sur une seule machine ne qualifie pas toute la flotte : les configurations des hébergeurs, kernels, runtimes et architectures peuvent différer.
