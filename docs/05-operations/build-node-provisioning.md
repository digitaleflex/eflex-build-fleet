# Provisionnement et cycle de vie d'un Build Node

## Objectif
Fournir une procédure reproductible d'ajout, qualification, drainage, mise à jour et retrait d'un nœud.

## Pré-requis d'un nœud
- Linux supporté et noyau qualifié.
- Architecture CPU identifiée.
- Ressources CPU, RAM et disque mesurées.
- Disque de workspace séparé ou quota contrôlé.
- Accès réseau nécessaire au Control Plane, Git provider et Registry.
- Pas de dépendance vers les réseaux de production.
- Accès KVM si le profil d'isolation Firecracker l'exige.
- Mises à jour sécurité et horloge synchronisée.
- Supervision de l'état disque, mémoire et processus.

La liste exacte des distributions, kernels et capacités KVM supportés sera publiée après les premiers tests sur les fournisseurs.

## Installation
1. Provisionner une machine sans données client runtime.
2. Appliquer le durcissement OS et les mises à jour.
3. Configurer firewall local et Tailscale selon le rôle du nœud.
4. Installer l'agent par un artefact signé et une version épinglée.
5. Créer l'identité du nœud et effectuer l'enrôlement mTLS.
6. Vérifier l'inventaire réel des capacités.
7. Installer/activer les versions qualifiées de BuildKit et des adaptateurs.
8. Configurer quotas, workspace, logs et nettoyage.
9. Exécuter les tests préflight et un build de qualification.
10. Passer le nœud à ready après approbation.

## Principe de moindre privilège
L'agent tourne avec un compte système dédié et les privilèges minimaux. Aucun accès direct au Docker socket de production, aux bases runtime ou aux secrets runtime. Si une capacité exige un privilège élevé, elle doit être isolée dans un composant et une politique explicitement documentés.

## Mise à jour
- Publier une version agent signée.
- Vérifier la compatibilité protocole avec Control Plane.
- Drainer le nœud.
- Attendre fin ou annuler proprement les jobs.
- Mettre à jour l'agent et les moteurs.
- Vérifier la configuration et les signatures.
- Exécuter un build test.
- Réinscrire le nœud en ready.
- Restaurer l'ancienne version si le test échoue.

Les moteurs de build ne sont pas mis à jour automatiquement avec l'agent.

## Drainage et retrait
Un nœud draining n'accepte plus de nouveaux jobs. Après fin des jobs et nettoyage, révoquer l'identité, supprimer les credentials locaux, vérifier les workspaces et retirer le nœud de l'inventaire.

## Quarantaine
Un nœud peut être quarantined en cas d'agent compromis, nettoyage incomplet, incohérence de digest, dérive de configuration, métriques anormales ou échec de tests de sécurité. Un nœud en quarantaine ne reçoit aucun nouveau job.

## Domaines de panne
Chaque nœud porte des labels fournisseur, région et zone/domaine de panne. Les labels sont vérifiés lors du provisionnement et utilisés par le scheduler pour la diversité.
