---
inclusion: always
---

# Go Quality Checks - Règles de Validation

## Règles de Validation du Code

Après chaque modification de code Go, tu DOIS exécuter les outils de validation suivants :

### 1. golangci-lint
```bash
golangci-lint run
```

**Objectif** : Vérifier la qualité du code, les erreurs de style, les bugs potentiels et les mauvaises pratiques.

**Résultat attendu** : `0 issues.`

### 2. gosec
```bash
gosec ./...
```

**Objectif** : Scanner le code pour détecter les vulnérabilités de sécurité.

**Résultat attendu** : `Issues : 0`

## Règles Importantes

### ❌ NE PAS faire de `go build`
- Ne pas exécuter `go build` ou `go build .` après les modifications
- Les builds sont gérés par le pipeline CI/CD et skaffold

### ✅ À FAIRE systématiquement
1. Modifier le code
2. Exécuter `golangci-lint run`
3. Exécuter `gosec ./...`
4. Corriger toutes les erreurs détectées
5. Répéter jusqu'à obtenir 0 issues

## Erreurs Courantes à Corriger

### golangci-lint

**ineffassign** - Assignation inefficace
```go
// ❌ Mauvais
changeType := "default"
switch action {
case "edit":
    changeType = "updated"
default:
    changeType = "default" // Redondant
}

// ✅ Bon
changeType := "default"
switch action {
case "edit":
    changeType = "updated"
}
```

**staticcheck ST1023** - Type inféré inutile
```go
// ❌ Mauvais
var priority string = "P3"

// ✅ Bon
priority := "P3"
```

**unused** - Fonction non utilisée
```go
// ❌ Supprimer les fonctions non utilisées
func unusedFunction() { }
```

### gosec

**G104** - Erreurs non gérées
```go
// ❌ Mauvais
resp.Body.Close()

// ✅ Bon
if err := resp.Body.Close(); err != nil {
    fmt.Printf("Error closing response body: %v\n", err)
}
```

**G304** - Inclusion de fichiers non validés
```go
// ❌ Mauvais
ioutil.ReadFile(userInput)

// ✅ Bon - Valider le chemin d'abord
if !isValidPath(userInput) {
    return errors.New("invalid path")
}
ioutil.ReadFile(userInput)
```

## Tests

### Créer des tests unitaires
- Créer des fichiers `*_test.go` pour les nouvelles fonctionnalités
- Utiliser `go test ./...` pour exécuter les tests
- Viser une couverture de code > 70%

### Structure des tests
```go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test"
    expected := "result"
    
    // Act
    result := FunctionName(input)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

## Workflow de Développement

1. **Créer une branche** pour chaque fonctionnalité
2. **Modifier le code** selon les besoins
3. **Valider avec golangci-lint** → corriger les erreurs
4. **Valider avec gosec** → corriger les vulnérabilités
5. **Écrire/mettre à jour les tests** si nécessaire
6. **Commit et push** une fois que tout est vert
7. **Créer une PR** pour review

## Commandes Utiles

```bash
# Vérifier la qualité du code
golangci-lint run

# Scanner les vulnérabilités
gosec ./...

# Exécuter les tests
go test ./...

# Vérifier la couverture
go test -cover ./...

# Formater le code
go fmt ./...

# Vérifier les imports
goimports -w .
```

## Notes Importantes

- **Toujours corriger les erreurs** avant de commit
- **Ne jamais ignorer les warnings** de sécurité
- **Documenter les fonctions publiques** avec des commentaires
- **Utiliser des noms de variables explicites**
- **Gérer toutes les erreurs** retournées par les fonctions
