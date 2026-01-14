package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Structures pour l'API Catalog
type CatalogResponse struct {
	Catalogs []CatalogItem `json:"catalogs"`
}

type CatalogItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Cache des projets avec TTL
type ProjectCache struct {
	projects    []string
	lastUpdated time.Time
	ttl         time.Duration
	mutex       sync.RWMutex
}

// Instance globale du cache
var projectCache = &ProjectCache{
	ttl: 5 * time.Minute, // TTL de 5 minutes
}

// GetProjects retourne la liste des projets depuis le cache ou l'API
func GetProjects() ([]string, error) {
	projectCache.mutex.RLock()

	// Vérifier si le cache est encore valide
	if time.Since(projectCache.lastUpdated) < projectCache.ttl && len(projectCache.projects) > 0 {
		// Cache valide, retourner une copie
		projects := make([]string, len(projectCache.projects))
		copy(projects, projectCache.projects)
		projectCache.mutex.RUnlock()
		logger.Debug("Cache hit",
			slog.Int("project_count", len(projects)),
			slog.Duration("cache_age", time.Since(projectCache.lastUpdated)),
		)
		return projects, nil
	}

	projectCache.mutex.RUnlock()

	// Cache expiré ou vide, récupérer depuis l'API
	return refreshProjectCache()
}

// refreshProjectCache met à jour le cache depuis l'API
func refreshProjectCache() ([]string, error) {
	projectCache.mutex.Lock()
	defer projectCache.mutex.Unlock()

	// Double-check après avoir acquis le lock
	if time.Since(projectCache.lastUpdated) < projectCache.ttl && len(projectCache.projects) > 0 {
		projects := make([]string, len(projectCache.projects))
		copy(projects, projectCache.projects)
		logger.Debug("Cache hit after lock", slog.Int("project_count", len(projects)))
		return projects, nil
	}

	logger.Info("Cache miss, fetching projects from API")

	projects, err := fetchProjectsFromAPI()
	if err != nil {
		logger.Error("Error fetching projects from API", slog.Any("error", err))

		// Si on a des données en cache (même expirées), les retourner
		if len(projectCache.projects) > 0 {
			logger.Warn("Using stale cache data", slog.Int("project_count", len(projectCache.projects)))
			staleProjects := make([]string, len(projectCache.projects))
			copy(staleProjects, projectCache.projects)
			return staleProjects, nil
		}

		return nil, err
	}

	// Mettre à jour le cache
	projectCache.projects = projects
	projectCache.lastUpdated = time.Now()

	logger.Info("Cache updated", slog.Int("project_count", len(projects)))

	// Retourner une copie
	result := make([]string, len(projects))
	copy(result, projects)
	return result, nil
}

// fetchProjectsFromAPI récupère la liste des projets depuis l'API Tracker
func fetchProjectsFromAPI() ([]string, error) {
	trackerHost := os.Getenv("TRACKER_HOST")
	if trackerHost == "" {
		return nil, fmt.Errorf("TRACKER_HOST environment variable not set")
	}

	apiURL := trackerHost + "/api/v1alpha1/catalogs/list"

	// Client HTTP avec timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	logger.Debug("Fetching projects from API", slog.String("url", apiURL))

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch catalogs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var catalogResponse CatalogResponse
	if err := json.Unmarshal(body, &catalogResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Filtrer les projets
	var projects []string
	for _, item := range catalogResponse.Catalogs {
		if item.Type == "project" {
			projects = append(projects, item.Name)
		}
	}

	// Trier par ordre alphabétique
	sort.Strings(projects)

	logger.Debug("Projects fetched from catalog", slog.Int("count", len(projects)))
	return projects, nil
}

// InitProjectCache initialise le cache au démarrage de l'application
func InitProjectCache() {
	logger.Info("Initializing project cache")

	go func() {
		_, err := GetProjects()
		if err != nil {
			logger.Warn("Failed to initialize project cache", slog.Any("error", err))
		} else {
			logger.Info("Project cache initialized successfully")
		}
	}()
}

// RefreshProjectCache force le rafraîchissement du cache
func RefreshProjectCache() error {
	logger.Debug("Manually refreshing project cache")
	_, err := refreshProjectCache()
	return err
}

// GetCacheStats retourne des statistiques sur le cache
func GetCacheStats() map[string]interface{} {
	projectCache.mutex.RLock()
	defer projectCache.mutex.RUnlock()

	return map[string]interface{}{
		"project_count": len(projectCache.projects),
		"last_updated":  projectCache.lastUpdated,
		"ttl_seconds":   projectCache.ttl.Seconds(),
		"is_expired":    time.Since(projectCache.lastUpdated) > projectCache.ttl,
		"age_seconds":   time.Since(projectCache.lastUpdated).Seconds(),
	}
}

// SearchProjectsInCache recherche des projets dans le cache sans appeler l'API
func SearchProjectsInCache(query string) []string {
	projectCache.mutex.RLock()
	defer projectCache.mutex.RUnlock()

	// Si le cache est vide, retourner une liste vide
	if len(projectCache.projects) == 0 {
		return []string{}
	}

	// Si pas de query, retourner tous les projets (limité à 100)
	if query == "" {
		limit := len(projectCache.projects)
		if limit > 100 {
			limit = 100
		}
		result := make([]string, limit)
		copy(result, projectCache.projects[:limit])
		return result
	}

	// Recherche case-insensitive
	queryLower := strings.ToLower(query)
	var results []string

	for _, project := range projectCache.projects {
		if strings.Contains(strings.ToLower(project), queryLower) {
			results = append(results, project)
			// Limiter à 100 résultats
			if len(results) >= 100 {
				break
			}
		}
	}

	return results
}
