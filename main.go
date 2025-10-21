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
	aiDifficulty string // "facile", "moyen", "difficile"
)

type GameData struct {
	*game.Board
	ScoreP1      int
	ScoreP2      int
	GamesPlayed  int
	ErrorMessage string
	AIMode       bool
	AIDifficulty string
}

const saveFile = "power4_save.json"

func main() {
	rand.Seed(time.Now().UnixNano())
	board = game.NewBoard()

	// Routes
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/start", startGameHandler)
	http.HandleFunc("/continue", continueHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/reset", resetHandler)
	http.HandleFunc("/reset-scores", resetScoresHandler)

	// Servir fichiers statiques
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("✅ Serveur lancé : http://localhost:8088")
	http.ListenAndServe(":8088", nil)
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		HasSave bool
	}{
		HasSave: hasSave(),
	}
	tmpl.ExecuteTemplate(w, "home.html", data)
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

	// Valeurs par défaut
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

	// Limiter longueur
	if len(player1) > 15 {
		player1 = player1[:15]
	}
	if len(player2) > 15 {
		player2 = player2[:15]
	}

	// Configuration IA
	aiMode = (aiModeStr == "on")
	if aiMode {
		aiDifficulty = difficulty
		if aiDifficulty == "" {
			aiDifficulty = "moyen"
		}
	}

	// Nouvelle partie
	board = game.NewBoardWithNames(player1, player2)
	scoreP1 = 0
	scoreP2 = 0
	
	// Supprimer ancienne sauvegarde
	deleteSave()

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
	data := GameData{
		Board:        board,
		ScoreP1:      scoreP1,
		ScoreP2:      scoreP2,
		AIMode:       aiMode,
		AIDifficulty: aiDifficulty,
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

	if board.IsColumnFull(col) {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	// Coup du joueur 1
	board.Move(col)
	board.TotalMoves++
	board.CheckWin()

	// Sauvegarder après coup joueur
	saveGame()

	// Si mode IA, joueur 2 est actif et partie pas finie
	if aiMode && board.Player == 2 && !board.GameOver {
		time.Sleep(500 * time.Millisecond)
		aiCol := getAIMove(board, aiDifficulty)
		if aiCol != -1 {
			board.Move(aiCol)
			board.TotalMoves++
			board.CheckWin()
			
			// Sauvegarder après coup IA
			saveGame()
		}
	}

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	// Incrémenter score si partie terminée
	if board.GameOver {
		switch board.Winner {
		case 1:
			scoreP1++
		case 2:
			scoreP2++
		}
		gamesPlayed++
	}

	// Conserver pseudos
	p1 := board.Player1Name
	p2 := board.Player2Name

	// Créer nouveau plateau
	board = game.NewBoardWithNames(p1, p2)
	
	// Sauvegarder le nouveau plateau
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

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

// ========== SYSTÈME IA ==========

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

// IA Facile : Joue au hasard
func aiEasy(b *game.Board) int {
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

// IA Moyen : Gagne si possible, bloque sinon joue au hasard
func aiMedium(b *game.Board) int {
	// 1. Essayer de gagner
	if col := findWinningMove(b, 2); col != -1 {
		return col
	}

	// 2. Bloquer le joueur
	if col := findWinningMove(b, 1); col != -1 {
		return col
	}

	// 3. Jouer au hasard
	return aiEasy(b)
}

// IA Difficile : Stratégie avancée
func aiHard(b *game.Board) int {
	// 1. Gagner
	if col := findWinningMove(b, 2); col != -1 {
		return col
	}

	// 2. Bloquer
	if col := findWinningMove(b, 1); col != -1 {
		return col
	}

	// 3. Privilégier centre
	if !b.IsColumnFull(3) {
		return 3
	}

	// 4. Colonnes proches du centre
	priority := []int{2, 4, 1, 5, 0, 6}
	for _, col := range priority {
		if !b.IsColumnFull(col) {
			return col
		}
	}

	return -1
}

// Trouve un coup gagnant pour le joueur spécifié
func findWinningMove(b *game.Board, player int) int {
	for col := 0; col < 7; col++ {
		if b.IsColumnFull(col) {
			continue
		}

		// Simuler le coup
		row := simulateMove(b, col, player)
		if row == -1 {
			continue
		}

		// Vérifier victoire
		if checkWinAt(b, row, col, player) {
			// Annuler simulation
			b.Grid[row][col] = 0
			return col
		}

		// Annuler simulation
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
	// Horizontal
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

	// Vertical
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

	// Diagonale ↗
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

	// Diagonale ↘
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
		AIMode       bool
		AIDifficulty string
	}

	data := SaveData{
		Board:        board,
		ScoreP1:      scoreP1,
		ScoreP2:      scoreP2,
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

func init() {
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
			return slice[i]
		},
	}
	tmpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
}