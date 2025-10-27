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

var (
	board       *game.Board
	tmpl        *template.Template
	scoreP1     int
	scoreP2     int
	gamesPlayed int
	aiMode      bool
	aiDifficulty string
)

type GameData struct {
	*game.Board
	ScoreP1      int
	ScoreP2      int
	GamesPlayed  int
	ErrorMessage string
	AIMode       bool
	AIDifficulty string
	SoundToPlay  string
	AIJustPlayed bool
}

const saveFile = "power4_save.json"

func main() {
	rand.Seed(time.Now().UnixNano())
	
	initTemplates()
	
	board = game.NewBoard()

	// Routes
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/start", startGameHandler)
	http.HandleFunc("/continue", continueHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/ai-play", aiPlayHandler)
	http.HandleFunc("/reset", resetHandler)
	http.HandleFunc("/reset-scores", resetScoresHandler)

	// Servir fichiers statiques
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("✅ Serveur lancé : http://localhost:8088")
	http.ListenAndServe(":8088", nil)
}

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

func startGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	player1 := r.FormValue("player1")
	player2 := r.FormValue("player2")
	aiModeStr := r.FormValue("ai_mode")
	difficulty := r.FormValue("difficulty")

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

	if len(player1) > 15 {
		player1 = player1[:15]
	}
	if len(player2) > 15 {
		player2 = player2[:15]
	}

	aiMode = (aiModeStr == "on")
	if aiMode {
		aiDifficulty = difficulty
		if aiDifficulty == "" {
			aiDifficulty = "moyen"
		}
	}

	board = game.NewBoardWithNames(player1, player2)
	scoreP1 = 0
	scoreP2 = 0
	gamesPlayed = 0
	
	deleteSave()
	saveGame()

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func continueHandler(w http.ResponseWriter, r *http.Request) {
	if loadGame() {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	soundToPlay := ""
	
	// Déterminer quel son jouer
	if board.GameOver {
		if board.Winner == 0 {
			soundToPlay = ""  // Pas de son pour match nul
		} else {
			soundToPlay = "win"
		}
	}
	
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

func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)

	if err != nil || col < 0 || col >= 7 {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	if board.GameOver || board.IsColumnFull(col) {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	// Coup du joueur 1
	board.Move(col)
	board.TotalMoves++
	board.CheckWin()

	saveGame()

	// Si la partie n'est pas terminée et c'est le mode IA
	if aiMode && board.Player == 2 && !board.GameOver {
		http.Redirect(w, r, "/game?ai_thinking=true", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func aiPlayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier que c'est bien le mode IA et le tour de l'IA
	if !aiMode || board.Player != 2 || board.GameOver {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Délai pour simulation de réflexion
	delay := getAIThinkingTime(aiDifficulty)
	time.Sleep(time.Duration(delay) * time.Millisecond)
	
	// L'IA joue
	aiCol := getAIMove(board, aiDifficulty)
	if aiCol != -1 {
		board.Move(aiCol)
		board.TotalMoves++
		board.CheckWin()
		saveGame()
	}

	w.WriteHeader(http.StatusOK)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if board.GameOver {
		switch board.Winner {
		case 1:
			scoreP1++
		case 2:
			scoreP2++
		}
		gamesPlayed++
	}

	p1 := board.Player1Name
	p2 := board.Player2Name

	board = game.NewBoardWithNames(p1, p2)
	
	saveGame()

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

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

// ========== SYSTÈME IA OPTIMISÉ ==========

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

// FACILE : 80-90% de chance pour le joueur de gagner
// L'IA joue presque aléatoirement avec seulement 10% de chance de bloquer
func aiEasy(b *game.Board) int {
	// Seulement 10% de chance de faire un coup intelligent
	if rand.Intn(100) < 10 {
		// Bloquer seulement si évident
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

// MOYEN : 60-70% de chance pour le joueur de gagner
// L'IA bloque systématiquement mais ne cherche pas à gagner agressivement
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
	
	// 5. Sinon aléatoire
	return aiEasy(b)
}

// DIFFICILE : 40-50% de chance pour le joueur de gagner
// L'IA joue stratégiquement avec anticipation
func aiHard(b *game.Board) int {
	// 1. Gagner immédiatement
	if col := findWinningMove(b, 2); col != -1 {
		return col
	}
	
	// 2. Bloquer l'adversaire
	if col := findWinningMove(b, 1); col != -1 {
		return col
	}
	
	// 3. Créer une menace double (fork) - 70% du temps
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
	
	// 5. Chercher à créer des alignements de 3
	if col := findTwoInRowMove(b, 2); col != -1 {
		return col
	}
	
	// 6. Évaluer les meilleures colonnes stratégiquement
	bestCol := evaluateBestMove(b)
	if bestCol != -1 {
		return bestCol
	}
	
	// 7. Fallback sur stratégie moyenne
	return aiMedium(b)
}

// Trouve un coup qui crée 2 jetons alignés avec possibilité d'extension
func findTwoInRowMove(b *game.Board, player int) int {
	for col := 0; col < 7; col++ {
		if b.IsColumnFull(col) {
			continue
		}
		
		row := simulateMove(b, col, player)
		if row == -1 {
			continue
		}
		
		if countAlignments(b, row, col, player, 2) > 0 {
			b.Grid[row][col] = 0
			return col
		}
		b.Grid[row][col] = 0
	}
	return -1
}

// Compte le nombre d'alignements d'une certaine longueur
func countAlignments(b *game.Board, row, col, player, length int) int {
	count := 0
	
	// Horizontal
	if checkAlignment(b, row, col, player, 0, 1, length) {
		count++
	}
	// Vertical
	if checkAlignment(b, row, col, player, 1, 0, length) {
		count++
	}
	// Diagonale \
	if checkAlignment(b, row, col, player, 1, 1, length) {
		count++
	}
	// Diagonale /
	if checkAlignment(b, row, col, player, 1, -1, length) {
		count++
	}
	
	return count
}

// Vérifie un alignement dans une direction
func checkAlignment(b *game.Board, row, col, player, dRow, dCol, length int) bool {
	count := 1
	
	// Direction positive
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
	
	// Direction négative
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

// Trouve un coup qui crée une menace double (fork)
func findForkMove(b *game.Board, player int) int {
	for col := 0; col < 7; col++ {
		if b.IsColumnFull(col) {
			continue
		}
		
		row := simulateMove(b, col, player)
		if row == -1 {
			continue
		}
		
		threats := countThreats(b, row, col, player)
		b.Grid[row][col] = 0
		
		if threats >= 2 {
			return col
		}
	}
	return -1
}

// Compte le nombre de menaces créées par un coup
func countThreats(b *game.Board, row, col, player int) int {
	threats := 0
	
	if checkLineOf3(b, row, col, player, 0, 1) {
		threats++
	}
	if checkLineOf3(b, row, col, player, 1, 0) {
		threats++
	}
	if checkLineOf3(b, row, col, player, 1, 1) {
		threats++
	}
	if checkLineOf3(b, row, col, player, 1, -1) {
		threats++
	}
	
	return threats
}

// Vérifie si un coup crée un alignement de 3 avec possibilité de 4
func checkLineOf3(b *game.Board, row, col, player, dRow, dCol int) bool {
	count := 1
	empty := 0
	
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
	
	return count == 3 && empty >= 1
}

// Évalue et retourne le meilleur coup basé sur un score
func evaluateBestMove(b *game.Board) int {
	bestScore := -1000
	bestCol := -1
	
	for col := 0; col < 7; col++ {
		if b.IsColumnFull(col) {
			continue
		}
		
		row := simulateMove(b, col, 2)
		if row == -1 {
			continue
		}
		
		score := evaluatePosition(b, row, col, 2)
		b.Grid[row][col] = 0
		
		if score > bestScore {
			bestScore = score
			bestCol = col
		}
	}
	
	return bestCol
}

// Évalue la qualité d'une position
func evaluatePosition(b *game.Board, row, col, player int) int {
	score := 0
	
	// Préférence pour le centre
	centerDistance := abs(col - 3)
	score += (3 - centerDistance) * 3
	
	// Préférence pour les positions basses
	score += (5 - row) * 2
	
	// Évaluer toutes les directions
	score += evaluateDirection(b, row, col, player, 0, 1)
	score += evaluateDirection(b, row, col, player, 1, 0)
	score += evaluateDirection(b, row, col, player, 1, 1)
	score += evaluateDirection(b, row, col, player, 1, -1)
	
	return score
}

// Évalue une direction spécifique
func evaluateDirection(b *game.Board, row, col, player, dRow, dCol int) int {
	score := 0
	count := 1
	empty := 0
	
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
				break
			}
		}
	}
	
	if count == 3 && empty >= 1 {
		score += 50
	} else if count == 2 && empty >= 2 {
		score += 10
	}
	
	return score
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func findWinningMove(b *game.Board, player int) int {
	for col := 0; col < 7; col++ {
		if b.IsColumnFull(col) {
			continue
		}
		row := simulateMove(b, col, player)
		if row == -1 {
			continue
		}
		if checkWinAt(b, row, col, player) {
			b.Grid[row][col] = 0
			return col
		}
		b.Grid[row][col] = 0
	}
	return -1
}

func simulateMove(b *game.Board, col, player int) int {
	for row := 5; row >= 0; row-- {
		if b.Grid[row][col] == 0 {
			b.Grid[row][col] = player
			return row
		}
	}
	return -1
}

func checkWinAt(b *game.Board, row, col, player int) bool {
	count := 1
	for c := col - 1; c >= 0 && b.Grid[row][c] == player; c-- {
		count++
	}
	for c := col + 1; c < 7 && b.Grid[row][c] == player; c++ {
		count++
	}
	if count >= 4 {
		return true
	}

	count = 1
	for r := row + 1; r < 6 && b.Grid[r][col] == player; r++ {
		count++
	}
	for r := row - 1; r >= 0 && b.Grid[r][col] == player; r-- {
		count++
	}
	if count >= 4 {
		return true
	}

	count = 1
	for i := 1; row-i >= 0 && col-i >= 0 && b.Grid[row-i][col-i] == player; i++ {
		count++
	}
	for i := 1; row+i < 6 && col+i < 7 && b.Grid[row+i][col+i] == player; i++ {
		count++
	}
	if count >= 4 {
		return true
	}

	count = 1
	for i := 1; row-i >= 0 && col+i < 7 && b.Grid[row-i][col+i] == player; i++ {
		count++
	}
	for i := 1; row+i < 6 && col-i >= 0 && b.Grid[row+i][col-i] == player; i++ {
		count++
	}
	return count >= 4
}

// ========== SAUVEGARDE ==========

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
		fmt.Println("Erreur sauvegarde:", err)
		return
	}

	err = ioutil.WriteFile(saveFile, jsonData, 0644)
	if err != nil {
		fmt.Println("Erreur écriture fichier:", err)
	}
}

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

	board = saveData.Board
	scoreP1 = saveData.ScoreP1
	scoreP2 = saveData.ScoreP2
	gamesPlayed = saveData.GamesPlayed
	aiMode = saveData.AIMode
	aiDifficulty = saveData.AIDifficulty

	return true
}

func hasSave() bool {
	_, err := os.Stat(saveFile)
	return err == nil
}

func deleteSave() {
	os.Remove(saveFile)
}

// ========== TEMPLATES ==========

func initTemplates() {
	funcMap := template.FuncMap{
		"Seq": func(n int) []int {
			result := make([]int, n)
			for i := range result {
				result[i] = i
			}
			return result
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"IsColumnFull": func(col int) bool {
			if board != nil {
				return board.IsColumnFull(col)
			}
			return false
		},
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
		panic("Erreur chargement templates: " + err.Error())
	}
	
	fmt.Println("✅ Templates chargés avec succès")
}