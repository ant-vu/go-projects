package main

import (
	"fmt"
	"math/rand"
	"strings"
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
		Discard:     make([]*Card, 0),
	}
	for i := range game.Players {
		game.Players[i] = &Player{
			Name: fmt.Sprintf("Player %d", i+1),
			Hand: make([]*Card, 0),
		}
		game.dealCards(game.Players[i], 7)
	}
	// Start the discard pile
	game.Discard = append(game.Discard, game.Deck[0])
	game.Deck = game.Deck[1:]
	return game
}

func (g *Game) PlayCard(index int) bool {
	if index < 0 || index >= len(g.Players[g.CurrentPlayer].Hand) {
		fmt.Println("Invalid card index")
		return false
	}
	card := g.Players[g.CurrentPlayer].Hand[index]
	if !g.isValidPlay(card) {
		fmt.Println("Invalid play")
		return false
	}
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

func (g *Game) isValidPlay(card *Card) bool {
	topCard := g.Discard[len(g.Discard)-1]
	if card.Suit == topCard.Suit || card.Rank == topCard.Rank {
		return true
	}
	if g.GameType == Uno && card.Rank == "Wild" {
		return true
	}
	if g.GameType == CrazyEights && card.Rank == "8" {
		return true
	}
	return false
}

func (g *Game) DrawCard() {
	if len(g.Deck) == 0 {
		fmt.Println("No cards left in deck")
		return
	}
	g.Players[g.CurrentPlayer].Hand = append(g.Players[g.CurrentPlayer].Hand, g.Deck[0])
	g.Deck = g.Deck[1:]
}

func (g *Game) HasWon() bool {
	return len(g.Players[g.CurrentPlayer].Hand) == 0
}

func generateDeck() []*Card {
	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	ranks := []string{"Ace", "2", "3", "4", "5", "6", "7", "8", "9", "10", "Jack", "Queen", "King"}
	if rand.Intn(2) == 0 {
		suits = []string{"Red", "Green", "Blue", "Yellow"}
		ranks = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "Reverse", "Skip", "Wild"}
	}
	deck := make([]*Card, 0)
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, &Card{Suit: suit, Rank: rank})
			if rank == "Wild" {
				deck = append(deck, &Card{Suit: suit, Rank: rank})
			}
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
	gameType = strings.TrimSpace(strings.ToLower(gameType))
	var gt GameType
	if gameType == "crazy eights" {
		gt = CrazyEights
	} else if gameType == "uno" {
		gt = Uno
	} else {
		fmt.Println("Invalid game type. Defaulting to Crazy Eights")
		gt = CrazyEights
	}
	game := NewGame(numPlayers, gt)
	for {
		fmt.Printf("Player %s's turn\n", game.Players[game.CurrentPlayer].Name)
		fmt.Println("Hand:")
		for i, card := range game.Players[game.CurrentPlayer].Hand {
			fmt.Printf("%d: %s of %s\n", i, card.Rank, card.Suit)
		}
		fmt.Printf("Top card: %s of %s\n", game.Discard[len(game.Discard)-1].Rank, game.Discard[len(game.Discard)-1].Suit)
		var action string
		fmt.Print("Enter action (play or draw): ")
		fmt.Scanln(&action)
		action = strings.TrimSpace(strings.ToLower(action))
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