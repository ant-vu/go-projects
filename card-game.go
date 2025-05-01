package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ... (existing code remains the same)

// GameType represents the type of card game
type GameType string

const (
	CrazyEights GameType = "Crazy Eights"
	Uno        GameType = "Uno"
)

// Game represents the card game
type Game struct {
	// ... (existing fields)
	GameType GameType
	Scores   map[string]int
}

// NewGame returns a new game
func NewGame(numPlayers int, gameType GameType) *Game {
	// ... (existing code)
	game := &Game{
		// ... (existing fields)
		GameType: gameType,
		Scores:   make(map[string]int),
	}
	// ... (existing code)
	return game
}

// PlayCard plays a card from the current player's hand
func (g *Game) PlayCard(index int) bool {
	// ... (existing code)
	if g.GameType == CrazyEights && card.Rank == Eight {
		// Implement "Crazy Eights" rule: player specifies suit
		var suit string
		fmt.Print("Enter suit: ")
		fmt.Scanln(&suit)
		g.Discard = append(g.Discard, &Card{Suit: Suit(suit), Rank: card.Rank})
	}
	// ... (existing code)
}

// DrawCard draws a card from the deck and adds it to the current player's hand
func (g *Game) DrawCard() {
	// ... (existing code)
	if g.GameType == Uno && len(g.Deck.Cards) == 0 {
		// Implement "Uno" rule: reverse direction
		g.CurrentPlayer = (g.CurrentPlayer - 1 + len(g.Players)) % len(g.Players)
	}
	// ... (existing code)
}

// HasWon checks if the current player has won
func (g *Game) HasWon() bool {
	// ... (existing code)
	if g.GameType == CrazyEights {
		// Implement "Crazy Eights" scoring
		for _, player := range g.Players {
			for _, card := range player.Hand {
				g.Scores[player.Name] += getCardValue(card)
			}
		}
	}
	return len(g.Players[g.CurrentPlayer].Hand) == 0
}

func getCardValue(card *Card) int {
	switch card.Rank {
	case Ace:
		return 1
	case Jack, Queen, King:
		return 10
	default:
		return int(card.Rank[0] - '0')
	}
}

func main() {
	numPlayers := 2
	var gameType string
	fmt.Print("Enter game type (Crazy Eights or Uno): ")
	fmt.Scanln(&gameType)
	game := NewGame(numPlayers, GameType(gameType))
	for {
		// ... (existing game loop)
		if game.HasWon() {
			fmt.Printf("Player %s wins!\n", game.Players[game.CurrentPlayer].Name)
			fmt.Println("Final Scores:")
			for player, score := range game.Scores {
				fmt.Printf("%s: %d\n", player, score)
			}
			break
		}
	}
}