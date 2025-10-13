package main

import (
    "power4/game"
    "html/template"
    "net/http"
    "strconv"
    "fmt"
)

var (
 board *game.Board
 tmpl *template.Template
 scoreP1 int
 scoreP2 int
)

type GameData struct {
 *game.Board
 ScoreP1 int
 ScoreP2 int
}


func main() {
    board = game.NewBoard()
		
    // Routes obligatoires
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/play", playHandler)
    http.HandleFunc("/reset", resetHandler)
    
    // Servir fichiers statiques
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
    fmt.Println("✅ Serveur lancé : http://localhost:8088")
    http.ListenAndServe(":8088", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := GameData{
	Board: board,
	ScoreP1: scoreP1,
	ScoreP2: scoreP2,
	}
    err := tmpl.ExecuteTemplate(w, "game.html", data)
    if err != nil {
        fmt.Println("Erreur de rendu template :", err)
    }
}

func playHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    
    colStr := r.FormValue("column")
    col, err := strconv.Atoi(colStr)
    
    if err != nil || col < 0 || col >= 7 {
        // Gérer erreur
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    
    success := board.Move(col)
    if !success {
        // Colonne pleine
    }
    
    board.CheckWin()
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if board.GameOver {
	if board.Winner == 1 {
	scoreP1++
	} else if board.Winner == 2 {
	scoreP2++
	}
	}
	board.Reset()
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
    }
    tmpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
}

func Seq(n int) []int {
    seq := make([]int, n)
    for i := range seq {
        seq[i] = i
    }
    return seq
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
    }
    tmpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
}
