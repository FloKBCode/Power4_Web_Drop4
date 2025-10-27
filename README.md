# 🎮 Power4 Web - Jeu de Puissance 4 en ligne


> Jeu de Puissance 4 moderne développé en Go avec une IA avancée à 3 niveaux de difficulté

---

## 📋 Table des matières

- [✨ Fonctionnalités](#-fonctionnalités)
- [🚀 Installation rapide](#-installation-rapide)
- [🎮 Comment jouer](#-comment-jouer)
- [🤖 Système d'IA](#-système-dia)
- [🏗️ Architecture](#️-architecture)
- [📚 Documentation technique](#-documentation-technique)

---

## ✨ Fonctionnalités

### 🎯 Modes de jeu
- **👥 2 Joueurs (PvP)** - Affrontez un ami en local
- **🤖 Contre IA** - 3 niveaux de difficulté (Facile / Moyen / Difficile)

### 🎨 Interface moderne
- ✅ Design néon moderne avec animations fluides
- ✅ Responsive : mobile, tablette, desktop
- ✅ Effets visuels : glow, animations de chute, highlight victoire
- ✅ Tableau des scores en temps réel
- ✅ Historique complet des coups

### 🧠 IA Avancée
- **Facile (😊)** : Coups aléatoires (victoire joueur ~85%)
- **Moyen (🤔)** : Blocage et attaque basique (victoire joueur ~65%)
- **Difficile (😈)** : Stratégie avancée avec anticipation (victoire joueur ~45%)

### 💾 Fonctionnalités avancées
- ✅ Sauvegarde automatique de la partie
- ✅ Détection de victoire dans toutes les directions
- ✅ Gestion du match nul
- ✅ Système audio immersif
- ✅ Tutoriel intégré

---

## 🚀 Installation rapide

### Prérequis
- Go 1.25 ou supérieur
- Un navigateur moderne (Chrome, Firefox, Safari, Edge)

### Installation

```bash
# 1. Cloner le projet
git clone https://github.com/votre-username/power4-web.git
cd power4-web

# 2. Vérifier Go
go version  # Doit afficher Go 1.25+

# 3. Créer le dossier des sons (si fichiers audio disponibles)
mkdir -p static/sounds

# 4. Lancer le serveur
go run main.go
```

Le serveur démarre sur **http://localhost:8088**

### Installation des dépendances

Aucune dépendance externe ! Le projet utilise uniquement la bibliothèque standard Go.

---

## 🎮 Comment jouer

### Démarrage rapide

1. **Ouvrir** : http://localhost:8088
2. **Choisir un mode** :
   - 👥 2 Joueurs (local)
   - 🤖 Contre l'IA
3. **Entrer les pseudos** (optionnel, max 15 caractères)
4. **Jouer** : Cliquer sur une colonne pour déposer un jeton

### Règles

- 🎯 **Objectif** : Aligner 4 jetons de votre couleur
- ➡️ Horizontalement, ⬇️ Verticalement, ou ↘️ En diagonale
- 🔴 Le joueur Rouge commence toujours
- 🔄 Jouez chacun votre tour
- 🚫 Une colonne pleine ne peut plus recevoir de jetons

### Stratégies gagnantes

```
💡 Conseil 1 : Contrôlez le CENTRE
   │ │ │🔴│ │ │ │
   Le centre offre le plus d'opportunités d'alignement

💡 Conseil 2 : Créez des MENACES DOUBLES (Fork)
   🔴─🔴─🔴─⚪  ← Menace horizontale
      │
      🔴        ← Menace verticale simultanée
      
💡 Conseil 3 : ANTICIPEZ 2-3 coups à l'avance
   Ne vous contentez pas de réagir, prévoyez !
```

---

## 🤖 Système d'IA

### Architecture algorithmique

Notre IA utilise une combinaison d'algorithmes :

#### 1️⃣ **Niveau Facile** (😊)
```go
// 90% coups aléatoires, 10% blocage évident
if rand.Intn(100) < 10 {
    bloque_menace_immediate()
} else {
    joue_aleatoire()
}
```
- **Temps de réflexion** : 600-1000ms
- **Chance de victoire joueur** : ~85%

#### 2️⃣ **Niveau Moyen** (🤔)
```go
// Stratégie défensive/offensive basique
1. Gagner si possible (30%)
2. Bloquer adversaire (70%)
3. Privilégier colonnes centrales
4. Éviter les pièges évidents
```
- **Temps de réflexion** : 1000-1500ms
- **Chance de victoire joueur** : ~65%

#### 3️⃣ **Niveau Difficile** (😈)
```go
// Algorithme stratégique avancé
1. findWinningMove()         // Gagner immédiatement
2. blockOpponentWin()        // Bloquer victoire adverse
3. createFork()              // Créer menace double
4. blockOpponentFork()       // Bloquer menace double adverse
5. buildThreeInRow()         // Créer alignements de 3
6. evaluateBestPosition()    // Évaluation heuristique
```
- **Temps de réflexion** : 1500-2300ms
- **Chance de victoire joueur** : ~45%
- **Fonctionnalités** :
  - Détection de forks (menaces doubles)
  - Évaluation de position (contrôle centre, hauteur)
  - Anticipation 2 coups à l'avance
  - Stratégie d'ouverture

### Exemple d'évaluation de position

```
Valeur des colonnes (IA Difficile) :
  0   1   2   3   4   5   6
 [+2][+4][+6][+9][+6][+4][+2]
           ↑
        Centre = maximum
```

---

## 🏗️ Architecture

### Structure du projet

```
power4-web/
├── main.go                 # Serveur HTTP + Routes + IA
├── game/
│   ├── board.go            # Structure + Logique plateau
│   └── win.go              # Détection victoire + Reset
├── templates/
│   ├── home.html           # Page d'accueil + Formulaire
│   └── game.html           # Interface de jeu
├── static/
│   ├── style.css           # Styles + Animations
│   
├── power4_save.json        # Sauvegarde auto (généré)
├── README.md               # Documentation
├── manual_test.md          # Plan de tests
└── go.mod                  # Dépendances Go
```


---

## 📚 Documentation technique

### API Endpoints

| Endpoint | Méthode | Description |
|----------|---------|-------------|
| `/` | GET | Page d'accueil |
| `/start` | POST | Démarrer nouvelle partie |
| `/continue` | GET | Reprendre partie sauvegardée |
| `/game` | GET | Afficher plateau de jeu |
| `/play` | POST | Jouer un coup (param: `column`) |
| `/ai-play` | POST | Coup de l'IA |
| `/reset` | POST | Nouvelle partie (conserve scores) |
| `/reset-scores` | POST | Réinitialiser scores |
| `/static/*` | GET | Fichiers statiques (CSS, sons) |

### Structure de données

#### Board
```go
type Board struct {
    Grid         [6][7]int      // Plateau 6 lignes × 7 colonnes
    Player       int            // Joueur actuel (1 ou 2)
    Winner       int            // Gagnant (0 = aucun, 1 ou 2)
    GameOver     bool           // Partie terminée ?
    History      []Move         // Historique des coups
    WinningCells [][2]int       // Cellules gagnantes
    Player1Name  string         // Pseudo joueur 1
    Player2Name  string         // Pseudo joueur 2
}
```

#### Move
```go
type Move struct {
    Player int  // Joueur qui a joué (1 ou 2)
    Column int  // Colonne jouée (0-6)
    Row    int  // Ligne finale (0-5)
}
```

#### SaveData (JSON)
```json
{
  "Board": { /* état complet */ },
  "ScoreP1": 2,
  "ScoreP2": 1,
  "GamesPlayed": 3,
  "AIMode": true,
  "AIDifficulty": "difficile"
}
```

### Algorithmes clés

#### 1. Détection de victoire
```go
func (b *Board) CheckWin() bool {
    // Vérifie 4 directions pour le dernier joueur
    if b.checkHorizontal() || 
       b.checkVertical() || 
       b.checkDiagonalUp() || 
       b.checkDiagonalDown() {
        b.Winner = 3 - b.Player
        b.GameOver = true
        return true
    }
    
    // Match nul si grille pleine
    if b.IsFull() {
        b.GameOver = true
        b.Winner = 0
        return true
    }
    
    return false
}
```

#### 2. IA - Recherche de coup gagnant
```go
func findWinningMove(b *Board, player int) int {
    for col := 0; col < 7; col++ {
        if b.IsColumnFull(col) { continue }
        
        // Simulation du coup
        row := simulateMove(b, col, player)
        
        // Test victoire dans toutes directions
        if checkWinAt(b, row, col, player) {
            b.Grid[row][col] = 0  // Annuler simulation
            return col             // Coup gagnant trouvé !
        }
        
        b.Grid[row][col] = 0
    }
    return -1  // Aucun coup gagnant
}
```

#### 3. IA - Détection de Fork (menace double)
```go
func findForkMove(b *Board, player int) int {
    for col := 0; col < 7; col++ {
        row := simulateMove(b, col, player)
        
        // Compter les menaces créées
        threats := countThreats(b, row, col, player)
        b.Grid[row][col] = 0
        
        if threats >= 2 {
            return col  // Fork trouvé !
        }
    }
    return -1
}

func countThreats(b *Board, row, col, player int) int {
    threats := 0
    
    // Pour chaque direction
    if checkLineOf3(b, row, col, player, 0, 1)  { threats++ }  // Horizontal
    if checkLineOf3(b, row, col, player, 1, 0)  { threats++ }  // Vertical
    if checkLineOf3(b, row, col, player, 1, 1)  { threats++ }  // Diag \
    if checkLineOf3(b, row, col, player, 1, -1) { threats++ }  // Diag /
    
    return threats
}
```

#### 4. Évaluation heuristique de position
```go
func evaluatePosition(b *Board, row, col, player int) int {
    score := 0
    
    // Bonus centre (colonnes 3 > 2,4 > 1,5 > 0,6)
    centerDistance := abs(col - 3)
    score += (3 - centerDistance) * 3
    
    // Bonus hauteur (bas > haut)
    score += (5 - row) * 2
    
    // Évaluer potentiel dans 4 directions
    score += evaluateDirection(b, row, col, player, 0, 1)   // →
    score += evaluateDirection(b, row, col, player, 1, 0)   // ↓
    score += evaluateDirection(b, row, col, player, 1, 1)   // ↘
    score += evaluateDirection(b, row, col, player, 1, -1)  // ↗
    
    return score
}
```

### Templates Go

Le projet utilise les templates Go avec fonctions personnalisées :

```go
funcMap := template.FuncMap{
    "Seq": func(n int) []int { /* 0 à n-1 */ },
    "add": func(a, b int) int { return a + b },
    "sub": func(a, b int) int { return a - b },
    "IsColumnFull": func(col int) bool { /* ... */ },
}
```

Exemple d'utilisation :
```html
{{range $i := Seq 6}}
  {{range $j := Seq 7}}
    <button class="cell">
      {{$cellValue := $.GetCell $i $j}}
      {{if eq $cellValue 1}}
        <div class="token red"></div>
      {{end}}
    </button>
  {{end}}
{{end}}
```

---

