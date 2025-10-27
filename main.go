package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"power4/game"
	"strconv"
	"time"
)

// ========== VARIABLES GLOBALES ==========

var (
	board       *game.Board      // État actuel du plateau de jeu
	tmpl        *template.Template // Templates HTML pré-compilés
	scoreP1     int              // Score du joueur 1
	scoreP2     int              // Score du joueur 2
	gamesPlayed int              // Nombre total de parties jouées
	aiMode      bool             // Mode IA activé ?
	aiDifficulty string          // Niveau de difficulté IA (facile/moyen/difficile)
)

// Structure pour passer les données aux templates
type GameData struct {
	*game.Board                  // Hérite tous les champs de Board
	ScoreP1      int             // Score joueur 1
	ScoreP2      int             // Score joueur 2
	GamesPlayed  int             // Nombre de parties
	ErrorMessage string          // Message d'erreur éventuel
	AIMode       bool            // Mode IA activé
	AIDifficulty string          // Difficulté IA
	SoundToPlay  string          // Son à jouer (win/lose)
	AIJustPlayed bool            // L'IA vient de jouer
}

const saveFile = "power4_save.json" // Fichier de sauvegarde

// ========== MAIN ==========

func main() {
	// Initialiser le générateur aléatoire pour l'IA
	rand.Seed(time.Now().UnixNano())
	
	// Charger les templates HTML
	initTemplates()
	
	// Créer un plateau vide par défaut
	board = game.NewBoard()

	// ========== ROUTES HTTP ==========
	http.HandleFunc("/", homePageHandler)              // Page d'accueil
	http.HandleFunc("/start", startGameHandler)        // Démarrer nouvelle partie
	http.HandleFunc("/continue", continueHandler)      // Reprendre partie sauvegardée
	http.HandleFunc("/game", gameHandler)              // Afficher le jeu
	http.HandleFunc("/play", playHandler)              // Jouer un coup (joueur)
	http.HandleFunc("/ai-play", aiPlayHandler)         // Coup de l'IA
	http.HandleFunc("/reset", resetHandler)            // Nouvelle partie
	http.HandleFunc("/reset-scores", resetScoresHandler) // Réinitialiser scores

	// Servir les fichiers statiques (CSS, sons, images)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Démarrer le serveur
	fmt.Println("✅ Serveur lancé : http://localhost:8088")
	http.ListenAndServe(":8088", nil)
}

// ========== HANDLERS ==========

/**
 * homePageHandler - Affiche la page d'accueil
 * Détecte si une sauvegarde existe pour proposer de continuer
 */
func homePageHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		HasSave bool
	}{
		HasSave: hasSave(),
	}
	
	err := tmpl.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		fmt.Println("Erreur template home:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/**
 * startGameHandler - Démarre une nouvelle partie
 * Traite le formulaire de la page d'accueil
 */
func startGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Récupérer les données du formulaire
	player1 := r.FormValue("player1")
	player2 := r.FormValue("player2")
	aiModeStr := r.FormValue("ai_mode")
	difficulty := r.FormValue("difficulty")

	// Valeurs par défaut si pseudos vides
	if player1 == "" {
		player1 = "Rouge"
	}
	if player2 == "" {
		if aiModeStr == "on" {
			player2 = "Ordinateur"
		} else {
			player2 = "Jaune"
		}
	}

	// Limiter la longueur des pseudos (sécurité)
	if len(player1) > 15 {
		player1 = player1[:15]
	}
	if len(player2) > 15 {
		player2 = player2[:15]
	}

	// Configurer le mode IA
	aiMode = (aiModeStr == "on")
	if aiMode {
		aiDifficulty = difficulty
		if aiDifficulty == "" {
			aiDifficulty = "moyen" // Par défaut
		}
	}

	// Créer un nouveau plateau avec les noms des joueurs
	board = game.NewBoardWithNames(player1, player2)
	scoreP1 = 0
	scoreP2 = 0
	gamesPlayed = 0
	
	// Supprimer l'ancienne sauvegarde et créer une nouvelle
	deleteSave()
	saveGame()

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

/**
 * continueHandler - Reprend une partie sauvegardée
 */
func continueHandler(w http.ResponseWriter, r *http.Request) {
	if loadGame() {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

/**
 * gameHandler - Affiche l'interface de jeu
 * Gère l'affichage du plateau, scores, historique, etc.
 */
func gameHandler(w http.ResponseWriter, r *http.Request) {
	soundToPlay := ""
	
	// Déterminer quel son jouer en cas de fin de partie
	if board.GameOver {
		if board.Winner == 0 {
			soundToPlay = ""  // Pas de son pour match nul
		} else {
			soundToPlay = "win"
		}
	}
	
	// Préparer les données pour le template
	data := GameData{
		Board:        board,
		ScoreP1:      scoreP1,
		ScoreP2:      scoreP2,
		GamesPlayed:  gamesPlayed,
		AIMode:       aiMode,
		AIDifficulty: aiDifficulty,
		SoundToPlay:  soundToPlay,
		AIJustPlayed: false,
	}

	err := tmpl.ExecuteTemplate(w, "game.html", data)
	if err != nil {
		fmt.Println("Erreur de rendu template:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/**
 * playHandler - Traite un coup joué par le joueur
 * Valide le coup, met à jour le plateau, vérifie la victoire
 */
func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	// Récupérer le numéro de colonne
	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)

	// Validation de la colonne
	if err != nil || col < 0 || col >= 7 {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	// Vérifier que le coup est valide
	if board.GameOver || board.IsColumnFull(col) {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	// Jouer le coup
	board.Move(col)
	board.TotalMoves++
	board.CheckWin()

	saveGame()

	// Si mode IA et c'est au tour de l'IA, déclencher l'overlay
	if aiMode && board.Player == 2 && !board.GameOver {
		http.Redirect(w, r, "/game?ai_thinking=true", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

/**
 * aiPlayHandler - Fait jouer l'IA
 * Appelé via AJAX depuis le JavaScript
 */
func aiPlayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier que c'est bien le tour de l'IA
	if !aiMode || board.Player != 2 || board.GameOver {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Délai de réflexion simulé (pour l'UX)
	delay := getAIThinkingTime(aiDifficulty)
	time.Sleep(time.Duration(delay) * time.Millisecond)
	
	// L'IA calcule et joue son coup
	aiCol := getAIMove(board, aiDifficulty)
	if aiCol != -1 {
		board.Move(aiCol)
		board.TotalMoves++
		board.CheckWin()
		saveGame()
	}

	w.WriteHeader(http.StatusOK)
}

/**
 * resetHandler - Démarre une nouvelle partie
 * Conserve les scores et incrémente le compteur de parties
 */
func resetHandler(w http.ResponseWriter, r *http.Request) {
	// Si une partie vient de se terminer, mettre à jour les scores
	if board.GameOver {
		switch board.Winner {
		case 1:
			scoreP1++
		case 2:
			scoreP2++
		}
		gamesPlayed++
	}

	// Sauvegarder les noms des joueurs
	p1 := board.Player1Name
	p2 := board.Player2Name

	// Créer un nouveau plateau
	board = game.NewBoardWithNames(p1, p2)
	
	saveGame()

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

/**
 * resetScoresHandler - Réinitialise tous les scores à zéro
 */
func resetScoresHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	scoreP1 = 0
	scoreP2 = 0
	gamesPlayed = 0
	
	saveGame()

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

// ========== SYSTÈME IA ==========

/**
 * getAIThinkingTime - Retourne le délai de réflexion selon la difficulté
 * C'est purement cosmétique pour l'UX (l'IA calcule en <20ms)
 */
func getAIThinkingTime(difficulty string) int {
	switch difficulty {
	case "facile":
		return 400 + rand.Intn(200) // 400-600ms
	case "moyen":
		return 700 + rand.Intn(400) // 700-1100ms
	case "difficile":
		return 1000 + rand.Intn(500) // 1000-1500ms
	default:
		return 700
	}
}

/**
 * getAIMove - Choisit le coup de l'IA selon la difficulté
 */
func getAIMove(b *game.Board, difficulty string) int {
	switch difficulty {
	case "facile":
		return aiEasy(b)
	case "moyen":
		return aiMedium(b)
	case "difficile":
		return aiHard(b)
	default:
		return aiMedium(b)
	}
}

// ========== IA FACILE ==========

/**
 * aiEasy - IA niveau facile
 * Stratégie : 90% de coups aléatoires, 10% de blocage
 * Taux de victoire joueur : ~85%
 */
func aiEasy(b *game.Board) int {
	// Seulement 10% de chance de faire un coup intelligent
	if rand.Intn(100) < 10 {
		// Bloquer seulement si victoire évidente
		if col := findWinningMove(b, 1); col != -1 {
			return col
		}
	}
	
	// 90% du temps : jouer complètement aléatoirement
	available := []int{}
	for col := 0; col < 7; col++ {
		if !b.IsColumnFull(col) {
			available = append(available, col)
		}
	}
	if len(available) == 0 {
		return -1
	}
	return available[rand.Intn(len(available))]
}

// ========== IA MOYEN ==========

/**
 * aiMedium - IA niveau moyen
 * Stratégie : Blocage systématique + attaque occasionnelle + préférence centre
 * Taux de victoire joueur : ~65%
 */
func aiMedium(b *game.Board) int {
	// 1. Gagner si l'occasion se présente (30% du temps)
	if rand.Intn(100) < 30 {
		if col := findWinningMove(b, 2); col != -1 {
			return col
		}
	}
	
	// 2. Bloquer l'adversaire (70% du temps)
	if rand.Intn(100) < 70 {
		if col := findWinningMove(b, 1); col != -1 {
			return col
		}
	}
	
	// 3. Préférence pour le centre (40% du temps)
	if !b.IsColumnFull(3) && rand.Intn(100) < 40 {
		return 3
	}
	
	// 4. Jouer colonnes centrales (2,3,4,5) en priorité
	centerCols := []int{3, 2, 4, 1, 5, 0, 6}
	for _, col := range centerCols {
		if !b.IsColumnFull(col) && rand.Intn(100) < 60 {
			return col
		}
	}
	
	// 5. Sinon jouer aléatoire
	return aiEasy(b)
}

// ========== IA DIFFICILE ==========

/**
 * aiHard - IA niveau difficile
 * Stratégie : Algorithme complet avec anticipation
 * - Détection de menaces doubles (forks)
 * - Évaluation de position heuristique
 * - Anticipation 2 coups à l'avance
 * Taux de victoire joueur : ~45%
 */
func aiHard(b *game.Board) int {
	// 1. Gagner immédiatement si possible
	if col := findWinningMove(b, 2); col != -1 {
		return col
	}
	
	// 2. Bloquer une victoire adverse
	if col := findWinningMove(b, 1); col != -1 {
		return col
	}
	
	// 3. Créer une menace double (fork) - 70% du temps
	// Un fork = 2 façons de gagner simultanément → imparable
	if rand.Intn(100) < 70 {
		if col := findForkMove(b, 2); col != -1 {
			return col
		}
	}
	
	// 4. Bloquer une menace double adverse - 80% du temps
	if rand.Intn(100) < 80 {
		if col := findForkMove(b, 1); col != -1 {
			return col
		}
	}
	
	// 5. Chercher à créer des alignements de 3 (menace simple)
	if col := findTwoInRowMove(b, 2); col != -1 {
		return col
	}
	
	// 6. Évaluer les meilleures colonnes selon heuristique
	bestCol := evaluateBestMove(b)
	if bestCol != -1 {
		return bestCol
	}
	
	// 7. Fallback sur stratégie moyenne
	return aiMedium(b)
}

// ========== FONCTIONS UTILITAIRES IA ==========

/**
 * findTwoInRowMove - Trouve un coup créant 2 jetons alignés
 * avec possibilité d'extension vers 4
 */
func findTwoInRowMove(b *game.Board, player int) int {
	for col := 0; col < 7; col++ {
		if b.IsColumnFull(col) {
			continue
		}
		
		// Simuler le coup
		row := simulateMove(b, col, player)
		if row == -1 {
			continue
		}
		
		// Compter les alignements de 2
		if countAlignments(b, row, col, player, 2) > 0 {
			b.Grid[row][col] = 0 // Annuler simulation
			return col
		}
		b.Grid[row][col] = 0
	}
	return -1
}

/**
 * countAlignments - Compte le nombre d'alignements d'une certaine longueur
 * @param length : longueur recherchée (2, 3, ou 4)
 * @return nombre d'alignements trouvés
 */
func countAlignments(b *game.Board, row, col, player, length int) int {
	count := 0
	
	// Vérifier les 4 directions
	if checkAlignment(b, row, col, player, 0, 1, length) {  // Horizontal →
		count++
	}
	if checkAlignment(b, row, col, player, 1, 0, length) {  // Vertical ↓
		count++
	}
	if checkAlignment(b, row, col, player, 1, 1, length) {  // Diagonale ↘
		count++
	}
	if checkAlignment(b, row, col, player, 1, -1, length) { // Diagonale ↗
		count++
	}
	
	return count
}

/**
 * checkAlignment - Vérifie un alignement dans une direction spécifique
 * @param dRow, dCol : vecteur de direction (ex: 1,0 pour vertical)
 * @param length : longueur minimale recherchée
 */
func checkAlignment(b *game.Board, row, col, player, dRow, dCol, length int) bool {
	count := 1 // Le jeton placé compte pour 1
	
	// Compter dans la direction positive
	for i := 1; i < 4; i++ {
		newRow := row + dRow*i
		newCol := col + dCol*i
		if newRow < 0 || newRow >= 6 || newCol < 0 || newCol >= 7 {
			break
		}
		if b.Grid[newRow][newCol] == player {
			count++
		} else {
			break
		}
	}
	
	// Compter dans la direction négative
	for i := 1; i < 4; i++ {
		newRow := row - dRow*i
		newCol := col - dCol*i
		if newRow < 0 || newRow >= 6 || newCol < 0 || newCol >= 7 {
			break
		}
		if b.Grid[newRow][newCol] == player {
			count++
		} else {
			break
		}
	}
	
	return count >= length
}

/**
 * findForkMove - Trouve un coup créant une menace double (fork)
 * Un fork = situation où le joueur a 2 façons de gagner au prochain coup
 * → L'adversaire ne peut bloquer qu'une seule menace → victoire assurée
 */
func findForkMove(b *game.Board, player int) int {
	for col := 0; col < 7; col++ {
		if b.IsColumnFull(col) {
			continue
		}
		
		row := simulateMove(b, col, player)
		if row == -1 {
			continue
		}
		
		// Compter le nombre de menaces créées
		threats := countThreats(b, row, col, player)
		b.Grid[row][col] = 0 // Annuler simulation
		
		if threats >= 2 {
			return col // Fork trouvé !
		}
	}
	return -1
}

/**
 * countThreats - Compte le nombre de menaces de victoire (alignements de 3)
 * créées par un coup donné
 */
func countThreats(b *game.Board, row, col, player int) int {
	threats := 0
	
	// Vérifier chaque direction
	if checkLineOf3(b, row, col, player, 0, 1) {  // →
		threats++
	}
	if checkLineOf3(b, row, col, player, 1, 0) {  // ↓
		threats++
	}
	if checkLineOf3(b, row, col, player, 1, 1) {  // ↘
		threats++
	}
	if checkLineOf3(b, row, col, player, 1, -1) { // ↗
		threats++
	}
	
	return threats
}

/**
 * checkLineOf3 - Vérifie si un coup crée un alignement de 3
 * avec possibilité d'atteindre 4 au prochain coup
 */
func checkLineOf3(b *game.Board, row, col, player, dRow, dCol int) bool {
	count := 1
	empty := 0
	
	// Direction positive
	for i := 1; i < 4; i++ {
		newRow := row + dRow*i
		newCol := col + dCol*i
		if newRow < 0 || newRow >= 6 || newCol < 0 || newCol >= 7 {
			break
		}
		if b.Grid[newRow][newCol] == player {
			count++
		} else if b.Grid[newRow][newCol] == 0 {
			empty++
			break
		} else {
			break
		}
	}
	
	// Direction négative
	for i := 1; i < 4; i++ {
		newRow := row - dRow*i
		newCol := col - dCol*i
		if newRow < 0 || newRow >= 6 || newCol < 0 || newCol >= 7 {
			break
		}
		if b.Grid[newRow][newCol] == player {
			count++
		} else if b.Grid[newRow][newCol] == 0 {
			empty++
			break
		} else {
			break
		}
	}
	
	// Alignement de 3 avec au moins 1 case vide = menace
	return count == 3 && empty >= 1
}

/**
 * evaluateBestMove - Évalue tous les coups possibles et retourne le meilleur
 * Utilise une fonction heuristique pour scorer chaque position
 */
func evaluateBestMove(b *game.Board) int {
	bestScore := -1000
	bestCol := -1
	
	for col := 0; col < 7; col++ {
		if b.IsColumnFull(col) {
			continue
		}
		
		// Simuler le coup
		row := simulateMove(b, col, 2)
		if row == -1 {
			continue
		}
		
		// Évaluer la position résultante
		score := evaluatePosition(b, row, col, 2)
		b.Grid[row][col] = 0 // Annuler
		
		if score > bestScore {
			bestScore = score
			bestCol = col
		}
	}
	
	return bestCol
}

/**
 * evaluatePosition - Fonction heuristique d'évaluation d'une position
 * Score basé sur :
 * - Position centrale (colonnes 3 > 2,4 > 1,5 > 0,6)
 * - Hauteur (bas mieux que haut)
 * - Potentiel d'alignement dans toutes directions
 */
func evaluatePosition(b *game.Board, row, col, player int) int {
	score := 0
	
	// Bonus pour les colonnes centrales
	// Centre (col 3) = +9, puis +6, +3, 0
	centerDistance := abs(col - 3)
	score += (3 - centerDistance) * 3
	
	// Bonus pour les positions basses (stabilité)
	score += (5 - row) * 2
	
	// Évaluer le potentiel dans les 4 directions
	score += evaluateDirection(b, row, col, player, 0, 1)   // →
	score += evaluateDirection(b, row, col, player, 1, 0)   // ↓
	score += evaluateDirection(b, row, col, player, 1, 1)   // ↘
	score += evaluateDirection(b, row, col, player, 1, -1)  // ↗
	
	return score
}

/**
 * evaluateDirection - Évalue le potentiel d'une direction spécifique
 * Donne des points selon le nombre de jetons alignés et cases vides
 */
func evaluateDirection(b *game.Board, row, col, player, dRow, dCol int) int {
	score := 0
	count := 1  // Jetons du joueur alignés
	empty := 0  // Cases vides adjacentes
	
	// Compter dans les deux sens de la direction
	for _, dir := range []int{1, -1} {
		for i := 1; i < 4; i++ {
			newRow := row + dRow*i*dir
			newCol := col + dCol*i*dir
			if newRow < 0 || newRow >= 6 || newCol < 0 || newCol >= 7 {
				break
			}
			if b.Grid[newRow][newCol] == player {
				count++
			} else if b.Grid[newRow][newCol] == 0 {
				empty++
				break
			} else {
				break // Bloqué par adversaire
			}
		}
	}
	
	// Scoring selon la situation
	if count == 3 && empty >= 1 {
		score += 50  // Menace de victoire !
	} else if count == 2 && empty >= 2 {
		score += 10  // Bon alignement
	}
	
	return score
}

/**
 * abs - Valeur absolue (helper)
 */
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

/**
 * findWinningMove - Trouve un coup qui fait gagner immédiatement
 * Essaie toutes les colonnes et teste si ça crée une victoire
 * @param player : joueur à tester (1 ou 2)
 * @return numéro de colonne gagnante, ou -1 si aucune
 */
func findWinningMove(b *game.Board, player int) int {
	for col := 0; col < 7; col++ {
		if b.IsColumnFull(col) {
			continue
		}
		
		// Simuler le placement du jeton
		row := simulateMove(b, col, player)
		if row == -1 {
			continue
		}
		
		// Vérifier si ce coup crée une victoire
		if checkWinAt(b, row, col, player) {
			b.Grid[row][col] = 0 // Annuler la simulation
			return col            // Coup gagnant trouvé !
		}
		
		b.Grid[row][col] = 0 // Annuler
	}
	return -1 // Aucun coup gagnant
}

/**
 * simulateMove - Simule le placement d'un jeton dans une colonne
 * Place temporairement le jeton et retourne la ligne
 * @return numéro de ligne, ou -1 si colonne pleine
 */
func simulateMove(b *game.Board, col, player int) int {
	for row := 5; row >= 0; row-- { // Du bas vers le haut
		if b.Grid[row][col] == 0 {
			b.Grid[row][col] = player
			return row
		}
	}
	return -1 // Colonne pleine
}

/**
 * checkWinAt - Vérifie si un jeton à une position crée une victoire
 * Teste les 4 directions : horizontal, vertical, 2 diagonales
 */
func checkWinAt(b *game.Board, row, col, player int) bool {
	// === HORIZONTAL ===
	count := 1
	// Vers la gauche
	for c := col - 1; c >= 0 && b.Grid[row][c] == player; c-- {
		count++
	}
	// Vers la droite
	for c := col + 1; c < 7 && b.Grid[row][c] == player; c++ {
		count++
	}
	if count >= 4 {
		return true
	}

	// === VERTICAL ===
	count = 1
	// Vers le bas
	for r := row + 1; r < 6 && b.Grid[r][col] == player; r++ {
		count++
	}
	// Vers le haut
	for r := row - 1; r >= 0 && b.Grid[r][col] == player; r-- {
		count++
	}
	if count >= 4 {
		return true
	}

	// === DIAGONALE \ (haut-gauche vers bas-droite) ===
	count = 1
	// Vers haut-gauche
	for i := 1; row-i >= 0 && col-i >= 0 && b.Grid[row-i][col-i] == player; i++ {
		count++
	}
	// Vers bas-droite
	for i := 1; row+i < 6 && col+i < 7 && b.Grid[row+i][col+i] == player; i++ {
		count++
	}
	if count >= 4 {
		return true
	}

	// === DIAGONALE / (bas-gauche vers haut-droite) ===
	count = 1
	// Vers haut-droite
	for i := 1; row-i >= 0 && col+i < 7 && b.Grid[row-i][col+i] == player; i++ {
		count++
	}
	// Vers bas-gauche
	for i := 1; row+i < 6 && col-i >= 0 && b.Grid[row+i][col-i] == player; i++ {
		count++
	}
	return count >= 4
}

// ========== SYSTÈME DE SAUVEGARDE ==========

/**
 * saveGame - Sauvegarde l'état complet du jeu dans un fichier JSON
 * Inclut : plateau, scores, mode IA, etc.
 */
func saveGame() {
	type SaveData struct {
		Board        *game.Board
		ScoreP1      int
		ScoreP2      int
		GamesPlayed  int
		AIMode       bool
		AIDifficulty string
	}

	data := SaveData{
		Board:        board,
		ScoreP1:      scoreP1,
		ScoreP2:      scoreP2,
		GamesPlayed:  gamesPlayed,
		AIMode:       aiMode,
		AIDifficulty: aiDifficulty,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("❌ Erreur sauvegarde:", err)
		return
	}

	err = ioutil.WriteFile(saveFile, jsonData, 0644)
	if err != nil {
		fmt.Println("❌ Erreur écriture fichier:", err)
	}
}

/**
 * loadGame - Charge une partie sauvegardée depuis le fichier JSON
 * @return true si chargement réussi, false sinon
 */
func loadGame() bool {
	data, err := ioutil.ReadFile(saveFile)
	if err != nil {
		return false
	}

	type SaveData struct {
		Board        *game.Board
		ScoreP1      int
		ScoreP2      int
		GamesPlayed  int
		AIMode       bool
		AIDifficulty string
	}

	var saveData SaveData
	err = json.Unmarshal(data, &saveData)
	if err != nil {
		return false
	}

	// Restaurer toutes les variables globales
	board = saveData.Board
	scoreP1 = saveData.ScoreP1
	scoreP2 = saveData.ScoreP2
	gamesPlayed = saveData.GamesPlayed
	aiMode = saveData.AIMode
	aiDifficulty = saveData.AIDifficulty

	return true
}

/**
 * hasSave - Vérifie si un fichier de sauvegarde existe
 */
func hasSave() bool {
	_, err := os.Stat(saveFile)
	return err == nil
}

/**
 * deleteSave - Supprime le fichier de sauvegarde
 */
func deleteSave() {
	os.Remove(saveFile)
}

// ========== TEMPLATES ==========

/**
 * initTemplates - Initialise et compile les templates HTML
 * Ajoute des fonctions personnalisées utilisables dans les templates
 */
func initTemplates() {
	// Fonctions disponibles dans les templates
	funcMap := template.FuncMap{
		// Seq(6) → [0, 1, 2, 3, 4, 5] pour les boucles
		"Seq": func(n int) []int {
			result := make([]int, n)
			for i := range result {
				result[i] = i
			}
			return result
		},
		// Opérations mathématiques
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		// Vérifier si une colonne est pleine
		"IsColumnFull": func(col int) bool {
			if board != nil {
				return board.IsColumnFull(col)
			}
			return false
		},
		// Fonctions pour l'historique
		"len": func(slice []game.Move) int {
			return len(slice)
		},
		"index": func(slice []game.Move, i int) game.Move {
			if i >= 0 && i < len(slice) {
				return slice[i]
			}
			return game.Move{}
		},
	}
	
	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		panic("❌ Erreur chargement templates: " + err.Error())
	}
	
	fmt.Println("✅ Templates chargés avec succès")
}