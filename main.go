package main

import (
    "power4/game"
    "html/template"
    "net/http"
    "strconv"
    "fmt"
)

var board *game.Board
var tmpl *template.Template

func main() {
    board = game.NewBoard()
    tmpl = template.Must(template.ParseGlob("templates/*.html"))
    
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
    tmpl.ExecuteTemplate(w, "game.html", board)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        colStr := r.FormValue("column")
        col, _ := strconv.Atoi(colStr)
        board.Move(col)
        board.CheckWin()
    }
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
    board.Reset()
    http.Redirect(w, r, "/", http.StatusSeeOther)
}