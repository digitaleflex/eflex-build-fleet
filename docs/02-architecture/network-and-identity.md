# Réseau privé, identité et communications

## Objectif
Contrôler une flotte répartie entre plusieurs hébergeurs sans exposer les ports d'administration des Build Nodes sur Internet.

## Décision réseau
Tailscale est le réseau privé de gestion initial. Il facilite la connectivité entre fournisseurs et réduit le besoin d'exposer SSH ou les API d'agents.

Tailscale est une couche de connectivité, pas le système d'identité applicative. L'architecture doit rester compatible avec un réseau WireGuard ou un réseau privé d'hébergeur.

## Authentification applicative
Les échanges Control Plane / Agent utilisent gRPC sur TLS avec authentification mutuelle. Chaque agent possède une identité unique, révocable et limitée. Les certificats doivent être émis, renouvelés et révoqués par un mécanisme contrôlé.

L'adresse Tailscale ou le nom d'une machine ne suffisent jamais à autoriser une commande.

## Modèle de connexion
- Les agents initient une connexion sortante vers le Control Plane.
- Le Control Plane n'ouvre pas un port d'administration public sur chaque agent.
- Le canal transporte heartbeat, claim, lease, progression, logs et résultat.
- Toute commande est liée à un job, tenant, lease et identité autorisée.
- Les commandes sensibles sont journalisées.

## Politique réseau Tailscale
La politique par défaut est deny. Les grants/ACL autorisent uniquement les flux nécessaires :
- Agent vers endpoint Control Plane.
- Agent vers Registry avec droit de push limité.
- Agent vers Git provider avec accès source temporaire.
- Admin vers nœud pour SSH explicitement accordé.
- Sandbox de build vers tailnet : interdit par défaut.

Les sandboxes n'héritent pas de l'accès réseau Tailscale de l'hôte. Le filtrage empêche un script client de joindre les adresses Tailscale, les services de métadonnées cloud, le Control Plane et les réseaux runtime.

## Disponibilité Tailscale
La coordination Tailscale est un service tiers. Lorsqu'elle devient indisponible, les connexions déjà établies et politiques en cache peuvent continuer, mais enrôlement, renouvellement des clés et changements de politique peuvent être bloqués.

Le système ne promet pas l'indépendance totale de Tailscale. Les certificats applicatifs suivent une durée limitée. Une panne prolongée ne permet pas à des identités expirées d'agir indéfiniment.

## Secrets
- Git : jetons read-only limités au dépôt et de courte durée.
- Registry : credentials limités au push d'un dépôt précis.
- Build : références temporaires, accès explicite, jamais dans les logs.
- Runtime : aucun accès par défaut aux agents.
- Contrôle : hors workspace et hors variables visibles par le build.

## SSH
SSH sert à l'administration d'urgence, pas au dispatch normal. L'accès est limité, audité, nominatif et révoqué lorsqu'il n'est plus nécessaire.

## Alternative future
Headscale ou WireGuard auto-administré pourra être étudié si l'indépendance du plan de contrôle réseau devient une exigence mesurée. Le remplacement doit conserver les mêmes contrats d'identité et de sécurité applicatives.
