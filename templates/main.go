package main

import (
    "drop4/game"
    "encoding/json"
    "fmt"
    "net/http"
)

var board *game.Board

func main() {
    board = game.NewBoard() // Initialiser le plateau

    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/play", playHandler)

    // IMPORTANT : Activer CORS pour que le navigateur accepte les requêtes
    fmt.Println("✅ Serveur lancé : http://localhost:8088")
    http.ListenAndServe(":8088", corsMiddleware(http.DefaultServeMux))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Bienvenue sur Drop4 Web !")
}

func playHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method != "POST" {
        http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
        return
    }

    // Lire la requête JSON
    var req struct {
        Column int `json:"column"`
    }

    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Requête invalide", http.StatusBadRequest)
        return
    }

    // Jouer le coup
    success := board.Move(req.Column)

    // Renvoyer le plateau
    response := map[string]interface{}{
        "board":         board.Grid,
        "currentPlayer": board.Player,
        "success":       success,
    }

    json.NewEncoder(w).Encode(response)
}

// Middleware CORS pour autoriser les requêtes du navigateur
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}