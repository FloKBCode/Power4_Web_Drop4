package game

const (
	Ligne    = 6
	Colonnes = 7
)

type Move struct {
 Player int
 Column int
 Row int
}

type Board struct {
	Grid   [Ligne][Colonnes]int
	Player int
    Winner   int  
    GameOver bool
    Error  string
    History []Move
    TotalMoves  int
    Player1Name string
    Player2Name string
}

// Créer un plateau vide
func NewBoard() *Board {
	return &Board{
        Player:      1,
        Player1Name: "Joueur 1",
        Player2Name: "Joueur 2",
    }
}

func NewBoardWithNames(p1, p2 string) *Board {
    return &Board{
        Player:      1,
        Player1Name: p1,
        Player2Name: p2,
    }
}

func (b *Board) GetCurrentPlayerName() string {
    if b.Player == 1 {
        return b.Player1Name
    }
    return b.Player2Name
}

func (b *Board) Move(col int) bool {
    if col < 0 || col >= Colonnes {
        b.Error = "Colonne invalide"
        return false
    }

    for ligne := Ligne - 1; ligne >= 0; ligne-- { // part du bas
        if b.Grid[ligne][col] == 0 {
            b.Grid[ligne][col] = b.Player

            b.History = append(b.History, Move{
            Player: b.Player,
            Column: col,
            Row: ligne,
            })

            // changer joueur
            if b.Player == 1 {
                b.Player = 2
            } else {
                b.Player = 1
            }
            return true
        }
    }
    return false 
}

func (b *Board) IsColumnFull(col int) bool {
    if col < 0 || col >= Colonnes {
        return true
    }
    return b.Grid[0][col] != 0
}