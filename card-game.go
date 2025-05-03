package main

import (
	"fmt"
	"math/rand"
	"time"
)

type GameType string

const (
	CrazyEights GameType = "Crazy Eights"
	Uno        GameType = "Uno"
)

type Card struct {
	Suit string
	Rank string
}

type Player struct {
	Name string
	Hand []*Card
}

type Game struct {
	GameType    GameType
	Players     []*Player
	CurrentPlayer int
	Deck        []*Card
	Discard     []*Card
	Scores      map[string]int
}

func NewGame(numPlayers int, gameType GameType) *Game {
	game := &Game{
		GameType: gameType,
		Players:     make([]*Player, numPlayers),
		Scores:      make(map[string]int),
		Deck:        generateDeck(),
	}
	for i := range game.Players {
		game.Players[i] = &Player{
			Name: fmt.Sprintf("Player %d", i+1),
			Hand: make([]*Card, 0),
		}
		game.dealCards(game.Players[i], 7)
	}
	return game
}

func (g *Game) PlayCard(index int) bool {
	card := g.Players[g.CurrentPlayer].Hand[index]
	g.Players[g.CurrentPlayer].Hand = append(g.Players[g.CurrentPlayer].Hand[:index], g.Players[g.CurrentPlayer].Hand[index+1:]...)
	g.Discard = append(g.Discard, card)
	if g.GameType == CrazyEights && card.Rank == "8" {
		var suit string
		fmt.Print("Enter suit: ")
		fmt.Scanln(&suit)
		g.Discard = append(g.Discard, &Card{Suit: suit, Rank: card.Rank})
	}
	return true
}

func (g *Game) DrawCard() {
	g.Players[g.CurrentPlayer].Hand = append(g.Players[g.CurrentPlayer].Hand, g.Deck[0])
	g.Deck = g.Deck[1:]
}

func (g *Game) HasWon() bool {
	return len(g.Players[g.CurrentPlayer].Hand) == 0
}

func generateDeck() []*Card {
	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	ranks := []string{"Ace", "2", "3", "4", "5", "6", "7", "8", "9", "10", "Jack", "Queen", "King"}
	deck := make([]*Card, 0)
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, &Card{Suit: suit, Rank: rank})
		}
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func (g *Game) dealCards(player *Player, numCards int) {
	for i := 0; i < numCards; i++ {
		player.Hand = append(player.Hand, g.Deck[0])
		g.Deck = g.Deck[1:]
	}
}

func main() {
	numPlayers := 2
	var gameType string
	fmt.Print("Enter game type (Crazy Eights or Uno): ")
	fmt.Scanln(&gameType)
	game := NewGame(numPlayers, GameType(gameType))
	for {
		fmt.Printf("Player %s's turn\n", game.Players[game.CurrentPlayer].Name)
		fmt.Println("Hand:")
		for i, card := range game.Players[game.CurrentPlayer].Hand {
			fmt.Printf("%d: %s of %s\n", i, card.Rank, card.Suit)
		}
		var action string
		fmt.Print("Enter action (play or draw): ")
		fmt.Scanln(&action)
		if action == "play" {
			var index int
			fmt.Print("Enter card index: ")
			fmt.Scanln(&index)
			game.PlayCard(index)
		} else if action == "draw" {
			game.DrawCard()
		}
		if game.HasWon() {
			fmt.Printf("Player %s wins!\n", game.Players[game.CurrentPlayer].Name)
			break
		}
		game.CurrentPlayer = (game.CurrentPlayer + 1) % len(game.Players)
	}
}