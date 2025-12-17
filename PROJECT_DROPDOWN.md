# Dropdown des Projets - Intégration API Catalog

## Vue d'ensemble

Cette fonctionnalité remplace le champ texte "Project" dans les modals Slack par un dropdown dynamique qui récupère la liste des projets depuis l'API Tracker Catalog.

## Fonctionnalités

### 🚀 Dropdown dynamique
- Liste des projets récupérée depuis `/api/v1alpha1/catalogs/list`
- Filtrage automatique des éléments de type "project"
- Tri alphabétique des projets
- Fallback automatique vers champ texte en cas d'erreur

### ⚡ Cache intelligent
- **TTL** : 1 heure (configurable)
- **Thread-safe** : Utilisation de mutex pour la concurrence
- **Double-check locking** : Évite les appels API redondants
- **Fallback** : Utilise les données en cache même expirées si l'API est indisponible

### 🔄 Rafraîchissement automatique
- **Cron job** : Toutes les heures (`0 * * * *`)
- **Initialisation** : Au démarrage de l'application (asynchrone)
- **Manuel** : Via l'endpoint `/cache/status`

## Architecture

### Fichiers créés
- `catalog.go` : Gestion du cache et appels API
- `modal_helpers.go` : Fonctions utilitaires pour les dropdowns
- `PROJECT_DROPDOWN.md` : Cette documentation

### Fichiers modifiés
- `main.go` : Initialisation du cache et tâche cron
- `slack_modal.go` : Remplacement des champs texte par des dropdowns
- `slack.go` : Extraction des valeurs du dropdown dans les modals

## API Endpoints

### `/health`
Endpoint de santé de l'application
```json
{"status":"ok"}
```

### `/cache/status`
Statistiques du cache des projets
```json
{
  "project_count": 15,
  "last_updated": "2024-12-17T06:45:00Z",
  "ttl_seconds": 3600,
  "is_expired": false,
  "age_seconds": 300
}
```

## Configuration

### Variables d'environnement
```bash
TRACKER_HOST=https://your-tracker-api.com
```

### Cache TTL
Modifiable dans `catalog.go` :
```go
var projectCache = &ProjectCache{
    ttl: 1 * time.Hour, // Modifier ici
}
```

## Comportement

### 1. Démarrage de l'application
```
[INFO] Initializing project cache...
[INFO] Fetching projects from: https://tracker-api/api/v1alpha1/catalogs/list
[INFO] Found 15 projects in catalog
[INFO] Cache updated: 15 projects loaded
[INFO] Project cache initialized successfully
```

### 2. Ouverture d'un modal
```
[INFO] Created project dropdown with 15 options
```

### 3. Cache hit
```
[INFO] Cache hit: returning 15 projects from cache
```

### 4. Cache miss / refresh
```
[INFO] Cache miss: fetching projects from API...
[INFO] Cache updated: 15 projects loaded
```

### 5. Erreur API avec fallback
```
[WARN] Error fetching projects from API: connection timeout
[INFO] Using stale cache data: 15 projects
[INFO] Created project text input (fallback)
```

## Modals affectés

Tous les modals incluent maintenant le dropdown des projets :
- **Deployment** (`/deployment`)
- **Drift** (`/drift`) 
- **Incident** (`/incident`)

Le modal **RPA Usage** (`/rpa_usage`) n'a pas de champ projet.

## Fallback et résilience

### Stratégies de fallback
1. **Cache expiré + API OK** → Rafraîchir le cache
2. **Cache expiré + API KO** → Utiliser cache périmé si disponible
3. **Pas de cache + API KO** → Champ texte
4. **Dropdown vide** → Champ texte

### Gestion d'erreurs
- Timeout API : 10 secondes
- Logs détaillés pour le debugging
- Pas de blocage de l'application

## Tests

### Test manuel
1. Démarrer l'application : `skaffold dev`
2. Exécuter `/deployment` dans Slack
3. Vérifier que le dropdown apparaît avec la liste des projets
4. Tester le fallback en coupant l'API Tracker

### Vérification du cache
```bash
curl http://localhost:8080/cache/status
```

## Performance

### Métriques attendues
- **Génération de modal** : < 50ms (cache hit)
- **Premier chargement** : < 2s (appel API)
- **Mémoire** : ~1KB par projet en cache
- **Réseau** : 1 appel API par heure maximum

### Optimisations
- Cache en mémoire (pas de base de données)
- Appels API asynchrones
- Tri des projets fait une seule fois
- Réutilisation des données en cas d'erreur

## Monitoring

### Logs à surveiller
- Erreurs d'appel API
- Échecs de parsing JSON
- Timeouts de connexion
- Statistiques de cache (hit/miss)

### Métriques utiles
- Nombre de projets en cache
- Âge du cache
- Fréquence des rafraîchissements
- Taux d'erreur API
