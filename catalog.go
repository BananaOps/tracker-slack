package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
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
	ttl: 1 * time.Hour, // TTL d'1 heure
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
		fmt.Printf("Cache hit: returning %d projects from cache\n", len(projects))
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
		fmt.Printf("Cache hit after lock: returning %d projects\n", len(projects))
		return projects, nil
	}

	fmt.Println("Cache miss: fetching projects from API...")

	projects, err := fetchProjectsFromAPI()
	if err != nil {
		fmt.Printf("Error fetching projects from API: %v\n", err)

		// Si on a des données en cache (même expirées), les retourner
		if len(projectCache.projects) > 0 {
			fmt.Printf("Using stale cache data: %d projects\n", len(projectCache.projects))
			staleProjects := make([]string, len(projectCache.projects))
			copy(staleProjects, projectCache.projects)
			return staleProjects, nil
		}

		return nil, err
	}

	// Mettre à jour le cache
	projectCache.projects = projects
	projectCache.lastUpdated = time.Now()

	fmt.Printf("Cache updated: %d projects loaded\n", len(projects))

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

	fmt.Printf("Fetching projects from: %s\n", apiURL)

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

	fmt.Printf("Found %d projects in catalog\n", len(projects))
	return projects, nil
}

// InitProjectCache initialise le cache au démarrage de l'application
func InitProjectCache() {
	fmt.Println("Initializing project cache...")

	go func() {
		_, err := GetProjects()
		if err != nil {
			fmt.Printf("Warning: Failed to initialize project cache: %v\n", err)
		} else {
			fmt.Println("Project cache initialized successfully")
		}
	}()
}

// RefreshProjectCache force le rafraîchissement du cache
func RefreshProjectCache() error {
	fmt.Println("Manually refreshing project cache...")
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
