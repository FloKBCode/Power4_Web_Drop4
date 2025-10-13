package game

func (b *Board) Reset() {
	b.Grid = [Ligne][Colonnes]int{}
	b.Player = 1
	b.Winner = 0
	b.GameOver = false
}

func (b *Board) IsFull() bool {
	for col := 0; col < Colonnes; col++ {
		if b.Grid[0][col] == 0 {
			return false
		}
	}
	return true
}

func (b *Board) GetCell(row, col int) int {
	if row < 0 || row >= Ligne || col < 0 || col >= Colonnes {
		return -1 // valeur invalide
	}
	return b.Grid[row][col]
}

func (b *Board) CheckWin() bool {
    // Vérifier victoire
    if b.checkHorizontal() || b.checkVertical() || 
       b.checkDiagonalUp() || b.checkDiagonalDown() {
        b.Winner = 3 - b.Player // Celui qui vient de jouer
        b.GameOver = true
        return true
    }
    
    // Vérifier match nul
    if b.IsFull() {
        b.GameOver = true
        b.Winner = 0
        return true
    }
    
    return false
}

func (b *Board) checkHorizontal() bool {
    for row := 0; row < Ligne; row++ {
        for col := 0; col <= Colonnes-4; col++ {
            if b.Grid[row][col] != 0 &&
               b.Grid[row][col] == b.Grid[row][col+1] &&
               b.Grid[row][col] == b.Grid[row][col+2] &&
               b.Grid[row][col] == b.Grid[row][col+3] {
                return true
            }
        }
    }
    return false
}

func (b *Board) checkVertical() bool {
    for col := 0; col < Colonnes; col++ {
        for row := 0; row <= Ligne-4; row++ {
            if b.Grid[row][col] != 0 &&
               b.Grid[row][col] == b.Grid[row+1][col] &&
               b.Grid[row][col] == b.Grid[row+2][col] &&
               b.Grid[row][col] == b.Grid[row+3][col] {
                return true
            }
        }
    }
    return false
}

func (b *Board) checkDiagonalUp() bool {
    for row := 3; row < Ligne; row++ {
        for col := 0; col <= Colonnes-4; col++ {
            if b.Grid[row][col] != 0 &&
               b.Grid[row][col] == b.Grid[row-1][col+1] &&
               b.Grid[row][col] == b.Grid[row-2][col+2] &&
               b.Grid[row][col] == b.Grid[row-3][col+3] {
                return true
            }
        }
    }
    return false
}

func (b *Board) checkDiagonalDown() bool {
    for row := 0; row <= Ligne-4; row++ {
        for col := 0; col <= Colonnes-4; col++ {
            if b.Grid[row][col] != 0 &&
               b.Grid[row][col] == b.Grid[row+1][col+1] &&
               b.Grid[row][col] == b.Grid[row+2][col+2] &&
               b.Grid[row][col] == b.Grid[row+3][col+3] {
                return true
            }
        }
    }
    return false
}