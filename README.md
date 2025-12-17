<p align="center">
  <img src="https://img.shields.io/badge/Slack-Integration-4A154B?style=for-the-badge&logo=slack&logoColor=white" alt="Slack Integration">
</p>

<h1 align="center">Tracker Slack</h1>

<p align="center">
  <strong>Slack Integration for Tracker</strong>
  <br />
  <em>Create and manage Tracker events directly from Slack</em>
</p>

<p align="center">
  <a href="https://github.com/BananaOps/tracker-slack/actions/workflows/release.yml">
    <img src="https://github.com/BananaOps/tracker-slack/actions/workflows/release.yml/badge.svg" alt="CI Status">
  </a>
  <a href="https://github.com/BananaOps/tracker-slack/releases">
    <img src="https://img.shields.io/github/v/release/BananaOps/tracker-slack?include_prereleases" alt="Latest Release">
  </a>
  <a href="https://github.com/BananaOps/tracker-slack/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/BananaOps/tracker-slack" alt="License">
  </a>
  <a href="https://goreportcard.com/report/github.com/BananaOps/tracker-slack">
    <img src="https://goreportcard.com/badge/github.com/BananaOps/tracker-slack" alt="Go Report Card">
  </a>
  <a href="https://snyk.io/test/github/BananaOps/tracker-slack">
    <img src="https://snyk.io/test/github/BananaOps/tracker-slack/badge.svg" alt="Known Vulnerabilities">
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Slack-API-4A154B?style=flat&logo=slack&logoColor=white" alt="Slack API">
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/Kubernetes-Ready-326CE5?style=flat&logo=kubernetes&logoColor=white" alt="Kubernetes">
  <img src="https://img.shields.io/badge/Helm-Chart-0F1689?style=flat&logo=helm&logoColor=white" alt="Helm">
</p>

---

## 🎯 Qu'est-ce que Tracker Slack ?

**Tracker Slack** est une intégration Slack qui connecte votre plateforme de communication d'équipe avec le système de suivi d'événements [Tracker](https://github.com/BananaOps/tracker). Elle permet aux équipes de créer, modifier et gérer les événements de déploiement, les incidents, les dérives de configuration et l'utilisation RPA directement depuis Slack en utilisant des commandes slash et des modales interactives.

### Pourquoi Tracker Slack ?

- **🚀 Intégration Slack Native** - Créez des événements sans quitter Slack
- **📝 Modales Interactives** - Formulaires riches pour la création d'événements détaillés
- **🔄 Mises à jour en Temps Réel** - Modifiez et mettez à jour les événements avec des boutons interactifs
- **📅 Résumés Quotidiens** - Résumés automatisés des événements quotidiens via cron
- **🎯 Support Multi-Événements** - Gérez les déploiements, incidents, dérives et utilisation RPA
- **🔗 Workflow Transparent** - Intégration directe avec l'API Tracker
- **⚡ Actions Rapides** - Mises à jour de statut avec réactions emoji et réponses en fil

### Cas d'Usage

- **Notifications de Déploiement** - Annoncez les déploiements avec des workflows d'approbation
- **Signalement d'Incidents** - Création rapide d'incidents et suivi de statut
- **Alertes de Dérive de Configuration** - Signalez et suivez les changements d'infrastructure
- **Journalisation d'Utilisation RPA** - Documentez les exécutions d'automatisation
- **Coordination d'Équipe** - Notifications des parties prenantes et mises à jour de statut
- **Support Daily Standup** - Résumés automatisés des événements planifiés

---

## ✨ Fonctionnalités

### 🎯 Commandes Slack
- **`/deployment`** - Créez des événements de déploiement avec des workflows d'approbation
- **`/incident`** - Signalez et suivez les incidents avec des niveaux de priorité
- **`/drift`** - Documentez la détection de dérive de configuration
- **`/rpa_usage`** - Enregistrez les exécutions d'automatisation de processus robotiques

### 📝 Modales Interactives
- **Formulaires Riches** : Création d'événements complète avec toutes les métadonnées
- **Sélection de Projet** : Menu déroulant dynamique depuis le catalogue Tracker
- **Support d'Environnement** : Environnements PROD, PREP, UAT, DEV
- **Gestion des Parties Prenantes** : Sélection multi-utilisateurs pour les notifications
- **Intégration de Liens** : Support pour les tickets et pull requests
- **Sélecteurs Date/Heure** : Planification précise avec support de fuseau horaire

### 🔄 Gestion d'Événements
- **Modifier les Événements** : Modifiez les événements existants via des boutons interactifs
- **Mises à jour de Statut** : Changements de statut rapides avec des menus déroulants
- **Workflow d'Approbation** : Approuvez/rejetez les déploiements avec des réactions
- **Réponses en Fil** : Mises à jour automatiques de statut dans les fils de messages
- **Réactions Emoji** : Indicateurs de statut visuels sur les messages

### 📅 Rapports Automatisés
- **Résumés Quotidiens** : Rapports programmés des événements du jour
- **Groupement par Environnement** : Événements organisés par environnement et projet
- **Liens Directs** : Accès rapide aux messages Slack originaux
- **Support de Fuseau Horaire** : Fuseau horaire configurable pour l'affichage de l'heure
- **Planification Cron** : Planification flexible avec des expressions cron

### 🔗 Intégration Tracker
- **API REST** : Communication directe avec le backend Tracker
- **Synchronisation d'Événements** : Synchronisation en temps réel entre Slack et Tracker
- **Mappage de Métadonnées** : Mappage complet des champs entre plateformes
- **Suivi de Statut** : Mises à jour de statut bidirectionnelles
- **Préservation des Liens** : Maintien des connexions vers les ressources externes

---

## 🚀 Démarrage Rapide

### Prérequis

1. **Instance Tracker** : Vous avez besoin d'une instance [Tracker](https://github.com/BananaOps/tracker) en cours d'exécution
2. **Application Slack** : Créez une application Slack avec les permissions requises
3. **Variables d'Environnement** : Configurez les variables d'environnement requises

### Configuration de l'Application Slack

1. **Créez une Application Slack** sur [api.slack.com](https://api.slack.com/apps)

2. **Configurez OAuth & Permissions** avec ces scopes :
   ```
   chat:write
   commands
   reactions:write
   users:read
   ```

3. **Ajoutez les Commandes Slash** :
   - `/deployment` - URL de requête : `https://votre-domaine.com/slack/command`
   - `/incident` - URL de requête : `https://votre-domaine.com/slack/command`
   - `/drift` - URL de requête : `https://votre-domaine.com/slack/command`
   - `/rpa_usage` - URL de requête : `https://votre-domaine.com/slack/command`

4. **Activez l'Interactivité** :
   - URL de requête : `https://votre-domaine.com/slack/interactive_api_endpoint`

5. **Obtenez vos tokens** :
   - Bot User OAuth Token (commence par `xoxb-`)
   - Signing Secret

### Utilisation avec Docker (Recommandé)

```bash
# Téléchargez l'image
docker pull bananaops/tracker-slack:latest

# Exécutez avec les variables d'environnement
docker run -d \
  -p 8080:8080 \
  -e SLACK_BOT_TOKEN="xoxb-votre-bot-token" \
  -e SLACK_SIGNING_SECRET="votre-signing-secret" \
  -e TRACKER_HOST="http://votre-instance-tracker:8080" \
  -e TRACKER_DEPLOYMENT_CHANNEL="deployments" \
  -e TRACKER_INCIDENT_CHANNEL="incidents" \
  -e TRACKER_DRIFT_CHANNEL="drift-alerts" \
  -e TRACKER_SLACK_WORKSPACE="votre-workspace" \
  -e TRACKER_TIMEZONE="Europe/Paris" \
  -e TRACKER_SLACK_CRON_MESSAGE="0 8 * * *" \
  bananaops/tracker-slack:latest
```

### Utilisation avec Kubernetes et Helm

```bash
# Ajoutez le dépôt Helm (si disponible)
helm repo add bananaops https://bananaops.github.io/helm-charts

# Installez avec des valeurs personnalisées
helm install tracker-slack bananaops/tracker-slack \
  --set secret.slack.bot_token="xoxb-votre-bot-token" \
  --set secret.slack.signing_secret="votre-signing-secret" \
  --set env.tracker.host="http://tracker.tracker.svc.cluster.local:8080" \
  --set ingress.hosts[0].host="tracker-slack.votredomaine.com"
```

### Depuis les Sources

```bash
# Clonez le dépôt
git clone https://github.com/BananaOps/tracker-slack.git
cd tracker-slack

# Définissez les variables d'environnement
export SLACK_BOT_TOKEN="xoxb-votre-bot-token"
export SLACK_SIGNING_SECRET="votre-signing-secret"
export TRACKER_HOST="http://localhost:8080"
export TRACKER_DEPLOYMENT_CHANNEL="deployments"
export TRACKER_INCIDENT_CHANNEL="incidents"
export TRACKER_DRIFT_CHANNEL="drift-alerts"
export TRACKER_SLACK_WORKSPACE="votre-workspace"
export TRACKER_TIMEZONE="Europe/Paris"
export TRACKER_SLACK_CRON_MESSAGE="0 8 * * *"

# Exécutez l'application
go run main.go
```

---

## 📖 Configuration

### Variables d'Environnement

| Variable | Description | Défaut | Requis |
|----------|-------------|---------|----------|
| `SLACK_BOT_TOKEN` | Token OAuth Bot User Slack | - | ✅ |
| `SLACK_SIGNING_SECRET` | Secret de Signature de l'App Slack | - | ✅ |
| `TRACKER_HOST` | URL de base de l'API Tracker | - | ✅ |
| `TRACKER_DEPLOYMENT_CHANNEL` | Canal Slack pour les déploiements | `deployments` | ✅ |
| `TRACKER_INCIDENT_CHANNEL` | Canal Slack pour les incidents | `incidents` | ✅ |
| `TRACKER_DRIFT_CHANNEL` | Canal Slack pour les alertes de dérive | `drift-alerts` | ✅ |
| `TRACKER_SLACK_WORKSPACE` | Nom du workspace Slack | - | ✅ |
| `TRACKER_TIMEZONE` | Fuseau horaire pour l'affichage des dates | `UTC` | ❌ |
| `TRACKER_SLACK_CRON_MESSAGE` | Planification cron pour les résumés quotidiens | `0 8 * * *` | ❌ |

### Configuration des Canaux Slack

Créez des canaux dédiés pour différents types d'événements :

```bash
# Structure de canaux recommandée
#deployments     - Annonces de déploiement et approbations
#incidents       - Rapports d'incidents et suivi
#drift-alerts    - Notifications de dérive de configuration
#general         - Résumés quotidiens (optionnel)
```

### Exemples de Planification Cron

```bash
# Quotidien à 8h
TRACKER_SLACK_CRON_MESSAGE="0 8 * * *"

# Jours de semaine à 9h
TRACKER_SLACK_CRON_MESSAGE="0 9 * * 1-5"

# Toutes les 2 heures pendant les heures de bureau
TRACKER_SLACK_CRON_MESSAGE="0 9-17/2 * * 1-5"
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Workspace Slack                       │
│                   /deployment /incident                    │
│                   /drift /rpa_usage                        │
└────────────────────────┬────────────────────────────────────┘
                         │ Commandes Slash & Interactions
                         │
                ┌────────▼────────┐
                │  Tracker Slack  │
                │   (Service Go)  │
                │                 │
                │ ┌─────────────┐ │
                │ │ Gestionnaire│ │
                │ │ API Slack   │ │
                │ └─────────────┘ │
                │ ┌─────────────┐ │
                │ │ Générateur  │ │
                │ │ de Modales  │ │
                │ └─────────────┘ │
                │ ┌─────────────┐ │
                │ │ Planificateur│ │
                │ │ Cron        │ │
                │ └─────────────┘ │
                └────────┬────────┘
                         │ Appels API REST
                         │
                ┌────────▼────────┐
                │   API Tracker   │
                │ (Service Principal) │
                │                 │
                │ ┌─────────────┐ │
                │ │ API Event   │ │
                │ └─────────────┘ │
                │ ┌─────────────┐ │
                │ │ API Catalog │ │
                │ └─────────────┘ │
                └────────┬────────┘
                         │
                ┌────────▼────────┐
                │   Base de       │
                │   Données       │
                │   (MongoDB)     │
                └─────────────────┘
```

### Flux des Composants

1. **Commandes Slack** → L'utilisateur tape `/deployment` dans Slack
2. **API Slack** → Envoie un webhook au service Tracker Slack
3. **Génération de Modale** → Crée un formulaire interactif basé sur le type d'événement
4. **Saisie Utilisateur** → L'utilisateur remplit le formulaire et soumet
5. **Intégration API** → Tracker Slack appelle l'API Tracker
6. **Création d'Événement** → Événement stocké dans la base de données Tracker
7. **Réponse Slack** → Message formaté posté dans le canal approprié
8. **Mises à jour Interactives** → Les utilisateurs peuvent modifier/mettre à jour via des boutons
9. **Résumés Quotidiens** → Le job cron récupère et poste les événements quotidiens

---

## 🎯 Exemples d'Utilisation

### 1. Créer un Événement de Déploiement

**Dans Slack :**
```
/deployment
```

Cela ouvre une modale interactive où vous pouvez remplir :
- **Résumé** : "Déployer user-service v2.1.0 en production"
- **Projet** : Sélectionner dans le menu déroulant (récupéré depuis le catalogue Tracker)
- **Environnement** : PROD, PREP, UAT, ou DEV
- **Impact** : Oui/Non
- **Date de Début** : Sélecteur de date/heure
- **Parties Prenantes** : Sélecteur multi-utilisateurs
- **Ticket** : Lien vers le ticket Jira/GitHub
- **Pull Request** : Lien vers la PR GitHub
- **Description** : Notes détaillées du déploiement

**Résultat** : Un message formaté est posté dans le canal de déploiement avec des boutons interactifs pour l'approbation, l'édition et les mises à jour de statut.

### 2. Signaler un Incident

**Dans Slack :**
```
/incident
```

**La modale inclut :**
- **Résumé** : "API Gateway retourne des erreurs 500"
- **Projet** : "api-gateway"
- **Environnement** : PROD
- **Priorité** : P1, P2, P3, ou P4
- **Parties Prenantes** : Membres de l'équipe d'astreinte
- **Ticket** : Lien vers le système de gestion d'incidents
- **Description** : Description détaillée de l'incident

**Résultat** : Posté dans le canal d'incidents avec bouton de fermeture et suivi automatique du statut.

### 3. Alerte de Dérive de Configuration

**Dans Slack :**
```
/drift
```

**Cas d'usage** : L'équipe infrastructure détecte des changements manuels en production
- **Résumé** : "Changements manuels de groupe de sécurité détectés"
- **Projet** : "infrastructure"
- **Environnement** : PROD
- **Description** : Détails sur la dérive

### 4. Journalisation d'Utilisation RPA

**Dans Slack :**
```
/rpa_usage
```

**Cas d'usage** : Documenter l'exécution d'automatisation
- **Résumé** : "Lot de traitement de factures terminé"
- **Environnement** : PROD
- **Date de Début** : Heure d'exécution
- **Description** : "150 factures traitées, 2 échecs nécessitant une révision manuelle"

### 5. Gestion Interactive d'Événements

Après avoir créé un événement, les membres de l'équipe peuvent interagir avec celui-ci en utilisant des boutons et des menus déroulants.

**Actions Disponibles :**

**Modifier l'Événement** : Cliquez sur le bouton "✏️ Modifier"
**Approuver le Déploiement** : Cliquez sur "✅ Approbation" pour approuver le déploiement
**Mettre à jour le Statut** : Utilisez le menu déroulant pour changer le statut :
- 🔄 EnCours
- ⏸️ Pause  
- ❌ Annulé
- ⏳ Reporté
- ✅ Terminé

**Mises à jour de Statut** : Réponses automatiques en fil et réactions emoji montrent le statut actuel.

### 6. Résumé d'Événements Quotidiens

**Automatisé à 8h** (configurable) :
```
📅 Événements Tracker d'Aujourd'hui :

🏭 PROD
  user-service:
    - 09:00 CET - Déployer user-service v2.1.0 en production [fil]
    - 14:30 CET - Déploiement hotfix pour bug critique [fil]

🧪 UAT  
  payment-service:
    - 10:00 CET - Déployer payment-service v1.5.0 pour test [fil]
```

### 7. Liaison d'Événements

Les événements se lient automatiquement vers Tracker :
- **ID de Message Slack** stocké dans Tracker
- **Liens directs** dans les résumés quotidiens
- **Synchronisation bidirectionnelle** entre plateformes

---

## 🛠️ Stack Technologique

### Backend
- **Langage** : Go 1.22.7+
- **Intégration Slack** : [slack-go/slack](https://github.com/slack-go/slack) v0.16.0
- **Planification** : [robfig/cron](https://github.com/robfig/cron) v3.0.1
- **Serveur HTTP** : Bibliothèque standard Go
- **Intégration API** : Client REST pour l'API Tracker

### Fonctionnalités Slack
- **Commandes Slash** : `/deployment`, `/incident`, `/drift`, `/rpa_usage`
- **Composants Interactifs** : Modales, boutons, menus déroulants, sélecteurs de date
- **Block Kit** : Formatage de messages riches avec le Block Kit de Slack
- **Webhooks** : Vérification de signature et analyse de payload
- **Mises à jour Temps Réel** : Édition de messages et réponses en fil

### DevOps & Déploiement
- **Conteneurisation** : Docker avec builds multi-étapes
- **Orchestration** : Déploiement Kubernetes
- **Chart Helm** : Chart Helm prêt pour la production inclus
- **Configuration** : Variables d'environnement et secrets Kubernetes
- **Monitoring** : Journalisation structurée pour l'observabilité

### Architecture d'Intégration
- **API Tracker** : Intégration RESTful avec le service Tracker principal
- **Mappage d'Événements** : Mappage complet des champs entre Slack et Tracker
- **Support de Fuseau Horaire** : Gestion configurable des fuseaux horaires
- **Planification Cron** : Planification flexible avec expressions cron

---

## 📊 Statut du Projet

- ✅ **Commandes Slack** : Prêt pour la production - Les 4 commandes slash implémentées
- ✅ **Modales Interactives** : Prêt pour la production - Formulaires riches pour la création d'événements
- ✅ **Gestion d'Événements** : Prêt pour la production - Créer, modifier, mettre à jour les événements
- ✅ **Résumés Quotidiens** : Prêt pour la production - Rapports automatisés basés sur cron
- ✅ **Docker** : Prêt pour la production - Builds Docker multi-étapes
- ✅ **Kubernetes** : Prêt pour la production - Chart Helm avec support ingress
- ✅ **Intégration Tracker** : Prêt pour la production - Intégration API complète
- 🚧 **Validation Webhook** : Fonctionnalités de sécurité avancées planifiées
- 🚧 **Workflows Avancés** : Processus d'approbation multi-étapes planifiés

---

## 🤝 Contribuer

Nous accueillons les contributions pour améliorer Tracker Slack ! Voici comment vous pouvez aider :

1. **🐛 Signaler des Bugs** : [Ouvrir une issue](https://github.com/BananaOps/tracker-slack/issues)
2. **💡 Suggérer des Fonctionnalités** : [Démarrer une discussion](https://github.com/BananaOps/tracker-slack/discussions)
3. **📝 Améliorer la Documentation** : Soumettre des améliorations de documentation
4. **🔧 Soumettre des PRs** : Corriger des bugs ou ajouter des fonctionnalités

### Configuration de Développement

```bash
# Cloner le dépôt
git clone https://github.com/BananaOps/tracker-slack.git
cd tracker-slack

# Installer les dépendances
go mod download

# Configurer les variables d'environnement (voir section Configuration)
cp .env.example .env
# Éditer .env avec vos valeurs

# Exécuter localement
go run main.go

# Exécuter les tests
go test ./...

# Formater le code
make fmt

# Linter le code
make lint
```

### Bonnes Premières Issues
- Ajouter le support pour plus de composants interactifs Slack
- Améliorer la gestion d'erreurs et les retours utilisateur
- Ajouter des tests unitaires pour la génération de modales
- Améliorer le formatage des résumés quotidiens
- Ajouter le support pour des réactions emoji personnalisées

---

## 🔗 Projets Liés

- **[Tracker](https://github.com/BananaOps/tracker)** - Plateforme principale de suivi d'événements
- **[Tracker GitHub Action](https://github.com/BananaOps/tracker-action)** - Intégration GitHub Actions
- **[Tracker CLI](https://github.com/BananaOps/tracker-cli)** - Interface en ligne de commande

## 💬 Communauté & Support

- **GitHub Issues** : [Signaler des bugs ou demander des fonctionnalités](https://github.com/BananaOps/tracker-slack/issues)
- **GitHub Discussions** : [Poser des questions et partager des idées](https://github.com/BananaOps/tracker-slack/discussions)
- **Projet Principal** : [Documentation Tracker](https://github.com/BananaOps/tracker)

## 🔒 Sécurité

Si vous découvrez une vulnérabilité de sécurité, veuillez envoyer un email à security@bananaops.org. Toutes les vulnérabilités de sécurité seront traitées rapidement.

---

<p align="center">
  Fait avec ❤️ par la communauté <a href="https://github.com/BananaOps">BananaOps</a>
</p>

<p align="center">
  <a href="https://github.com/BananaOps/tracker-slack/stargazers">⭐ Donnez-nous une étoile sur GitHub</a>
  •
  <a href="https://github.com/BananaOps/tracker-slack/issues">🐛 Signaler un Bug</a>
  •
  <a href="https://github.com/BananaOps/tracker-slack/discussions">💬 Rejoindre la Discussion</a>
</p>
