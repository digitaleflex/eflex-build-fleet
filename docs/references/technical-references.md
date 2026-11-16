# Références techniques officielles

**Date de consultation de la baseline : 16 novembre 2026.** Ces liens sont les références primaires pour mettre à jour les choix. Une nouvelle version publiée ne change pas automatiquement une décision : appliquer la politique de décision.

## Architecture et moteurs de build
- BuildKit — dépôt : https://github.com/moby/buildkit
- BuildKit — documentation : https://docs.docker.com/build/buildkit/
- BuildKit — builders et drivers : https://docs.docker.com/build/builders/drivers/
- BuildKit — mode rootless : https://docs.docker.com/build/buildkit/rootless/
- Railpack — dépôt : https://github.com/railwayapp/railpack
- Railpack — releases : https://github.com/railwayapp/railpack/releases
- Railpack — site : https://railpack.com/
- Nixpacks — documentation et statut maintenance : https://nixpacks.com/docs
- Nixpacks — dépôt : https://github.com/railwayapp/nixpacks
- Cloud Native Buildpacks — documentation : https://buildpacks.io/docs/
- ko — dépôt : https://github.com/ko-build/ko
- Jib — dépôt : https://github.com/GoogleContainerTools/jib

## Réseau et identité
- Tailscale — control plane et data plane : https://tailscale.com/docs/concepts/control-data-planes
- Tailscale — comportement lors d'une panne du coordination server : https://tailscale.com/docs/reference/coordination-server-down
- Headscale — serveur de coordination auto-hébergé : https://github.com/juanfont/headscale
- WireGuard — documentation : https://www.wireguard.com/

## Artefacts et registre
- OCI Distribution Specification : https://github.com/opencontainers/distribution-spec
- OCI Image Specification : https://github.com/opencontainers/image-spec
- Zot — registre OCI : https://zotregistry.dev/
- Docker Buildx : https://docs.docker.com/build/buildx/

## Isolation et sécurité
- Firecracker : https://github.com/firecracker-microvm/firecracker
- Firecracker Jailer : https://github.com/firecracker-microvm/firecracker/blob/main/docs/jailer.md
- gVisor : https://gvisor.dev/docs/
- Syft : https://github.com/anchore/syft
- Trivy : https://github.com/aquasecurity/trivy
- Cosign : https://docs.sigstore.dev/cosign/
- SLSA build track : https://slsa.dev/spec/v1.2/levels
- SLSA provenance : https://slsa.dev/spec/v1.2/provenance

## Observabilité
- Prometheus : https://prometheus.io/docs/introduction/overview/
- OpenTelemetry : https://opentelemetry.io/docs/

## Comparaison d'architecture
- Depot — architecture des builds conteneurisés : https://depot.dev/docs/container-builds/overview
- Coolify — build servers : https://coolify.io/docs/knowledge-base/server/build-server

## Licences
Les licences doivent être vérifiées sur le fichier LICENSE et les notices du tag réellement épinglé, pas uniquement sur la page principale du dépôt. Les composants CNB et buildpacks sont contrôlés individuellement.

Le dépôt E-FLEX Build Fleet ne déclare pas encore de licence de réutilisation du code propre à E-FLEX. Ne pas déduire une licence permissive de sa visibilité publique.
