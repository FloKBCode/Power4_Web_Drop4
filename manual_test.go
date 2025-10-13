package main

import (
	"fmt"
	"power4/game"
)

func testBoard() {
	fmt.Println("=== Tests board.go ===")

	// Test 1: NewBoard
	b := game.NewBoard()
	fmt.Printf("✓ NewBoard - Player: %d (attendu: 1)\n", b.Player)

	// Test 2: Reset
	b.Winner = 1
	b.GameOver = true
	b.Reset()
	fmt.Printf("✓ Reset - Winner: %d, GameOver: %v (attendu: 0, false)\n", b.Winner, b.GameOver)

	// Test 3: IsFull sur plateau vide
	fmt.Printf("✓ IsFull (vide): %v (attendu: false)\n", b.IsFull())

	// Test 4: IsFull sur plateau plein
	for col := 0; col < 7; col++ {
		for row := 0; row < 6; row++ {
			b.Grid[row][col] = 1
		}
	}
	fmt.Printf("✓ IsFull (plein): %v (attendu: true)\n", b.IsFull())

	// Test 5: GetCell
	b.Reset()
	b.Grid[3][2] = 1
	fmt.Printf("✓ GetCell(3,2): %d (attendu: 1)\n", b.GetCell(3, 2))
	fmt.Printf("✓ GetCell(-1,0): %d (attendu: -1)\n", b.GetCell(-1, 0))

	// Test 6: Move
	b.Reset()
	success := b.Move(3)
	fmt.Printf("✓ Move(3): %v, Grid[5][3]: %d, Player: %d (attendu: true, 1, 2)\n",
		success, b.Grid[5][3], b.Player)
}
