package main

import (
	"fmt"
	"html/template"
	"net/http"
	"power4/game"
	"strconv"
)

var (
	board       *game.Board
	tmpl        *template.Template
	scoreP1     int
	scoreP2     int
	gamesPlayed int
)

type GameData struct {
	*game.Board
	ScoreP1      int
	ScoreP2      int
	GamesPlayed  int
	ErrorMessage string
}

func main() {
	board = game.NewBoard()

	// Routes
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/start", startGameHandler)
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
	tmpl.ExecuteTemplate(w, "home.html", nil)
}

func startGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	player1 := r.FormValue("player1")
	player2 := r.FormValue("player2")

	// Valeurs par défaut si vides
	if player1 == "" {
		player1 = "Rouge"
	}
	if player2 == "" {
		player2 = "Jaune"
	}

	// Limiter longueur
	if len(player1) > 15 {
		player1 = player1[:15]
	}
	if len(player2) > 15 {
		player2 = player2[:15]
	}

	// Nouvelle partie avec pseudos
	board = game.NewBoardWithNames(player1, player2)
	scoreP1 = 0
	scoreP2 = 0

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	data := GameData{
		Board:   board,
		ScoreP1: scoreP1,
		ScoreP2: scoreP2,
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

	board.Move(col)
	board.TotalMoves++
	board.CheckWin()

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if board.GameOver {
		if board.Winner == 1 {
			scoreP1++
		} else if board.Winner == 2 {
			scoreP2++
		}
		gamesPlayed++
	}

	p1 := board.Player1Name
	p2 := board.Player2Name

	board = game.NewBoardWithNames(p1, p2)
	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func resetScoresHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	scoreP1 = 0
	scoreP2 = 0

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

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