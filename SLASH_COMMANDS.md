# Commandes Slash Tracker

## Vue d'ensemble

L'application Tracker Slack supporte plusieurs commandes slash pour interagir avec le système de tracking des déploiements.

## Commandes disponibles

### 📋 Commandes de création

#### `/deployment`
Ouvre un modal pour créer un nouveau déploiement.

**Utilisation :**
```
/deployment
```

**Fonctionnalités :**
- Dropdown des projets (récupéré depuis l'API Catalog)
- Sélection de l'environnement (PROD, PREP, UAT, DEV)
- Gestion de l'impact
- Notifications équipes (Release/Support)
- Dates de début/fin
- Stakeholders
- Liens (PR, Ticket)
- Description

#### `/incident`
Ouvre un modal pour créer un nouvel incident.

**Utilisation :**
```
/incident
```

**Fonctionnalités :**
- Dropdown des projets
- Sélection de l'environnement
- Niveau de priorité (P1, P2, P3, P4)
- Stakeholders
- Lien ticket
- Description obligatoire

#### `/drift`
Ouvre un modal pour créer un drift de configuration.

**Utilisation :**
```
/drift
```

**Fonctionnalités :**
- Dropdown des projets
- Sélection de l'environnement
- Stakeholders
- Liens (PR, Ticket)
- Description

#### `/rpa_usage`
Ouvre un modal pour créer un usage RPA.

**Utilisation :**
```
/rpa_usage
```

**Fonctionnalités :**
- Sélection de l'environnement
- Date de début
- Description obligatoire

### 📊 Commandes d'information

#### `/today`
Affiche tous les déploiements et événements prévus pour aujourd'hui.

**Utilisation :**
```
/today
```

**Fonctionnalités :**
- Même format que le message cron quotidien
- Groupement par environnement et projet
- Affichage public dans le canal
- Gestion d'erreur avec message privé

**Exemple de sortie :**
```
📅 Today Tracker Events :

🔴 PROD
├── 🚀 project-api
│   └── 09:00 Europe/Paris - Déploiement v2.1.0
├── 🚀 project-web
│   └── 14:30 Europe/Paris - Hotfix sécurité

🟡 PREP
├── 🚀 project-mobile
│   └── 10:15 Europe/Paris - Test nouvelle feature
```

## Configuration Slack

### Permissions requises

L'application nécessite les permissions suivantes :
- `commands` : Pour recevoir les commandes slash
- `chat:write` : Pour envoyer des messages
- `users:read` : Pour les mentions d'utilisateurs

### Variables d'environnement

```bash
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_SIGNING_SECRET=your-signing-secret
TRACKER_HOST=https://your-tracker-api.com
TRACKER_DEPLOYMENT_CHANNEL=C1234567890
TRACKER_DRIFT_CHANNEL=C1234567891
TRACKER_INCIDENT_CHANNEL=C1234567892
```

### Configuration des commandes

Dans l'interface Slack App :

1. **Slash Commands** → **Create New Command**

2. **Commandes à créer :**

| Commande | URL | Description | Usage Hint |
|----------|-----|-------------|------------|
| `/deployment` | `https://your-app.com/slack/command` | Créer un nouveau déploiement | |
| `/incident` | `https://your-app.com/slack/command` | Créer un nouvel incident | |
| `/drift` | `https://your-app.com/slack/command` | Créer un drift de config | |
| `/rpa_usage` | `https://your-app.com/slack/command` | Créer un usage RPA | |
| `/today` | `https://your-app.com/slack/command` | Voir les événements du jour | |

3. **Paramètres recommandés :**
   - **Escape channels, users, and links sent to your app** : ✅ Activé
   - **Short Description** : Description courte de la commande
   - **Usage Hint** : Laisser vide (pas de paramètres)

## Gestion d'erreur

### Erreurs API
- **Message privé** (ephemeral) à l'utilisateur
- **Log détaillé** côté serveur
- **Fallback gracieux** vers champ texte pour les projets

### Erreurs de validation
- **Messages d'erreur** dans les modals
- **Champs obligatoires** marqués clairement
- **Validation côté client** et serveur

## Monitoring

### Logs utiles
```bash
# Commandes reçues
"Handling /today command for user john.doe in channel general"

# Succès
"/today command processed successfully for user john.doe"

# Erreurs
"Error fetching today's events: connection timeout"
```

### Métriques
- Nombre de commandes par type
- Temps de réponse des modals
- Taux d'erreur API
- Utilisation par utilisateur/canal

## Développement

### Ajouter une nouvelle commande

1. **Ajouter le case dans `handleCommand`** :
```go
case "/ma_commande":
    handleMaCommande(w, s)
```

2. **Créer le handler** :
```go
func handleMaCommande(w http.ResponseWriter, s slack.SlashCommand) {
    // Logique de la commande
    sendSlackResponse(w, "Réponse", "in_channel")
}
```

3. **Configurer dans Slack App**

### Fonctions utilitaires

- `sendSlackResponse(w, text, type)` : Envoyer une réponse
- `fetchEvents()` : Récupérer les événements du jour
- `formatSlackMessageByEnvironment()` : Formater les messages
