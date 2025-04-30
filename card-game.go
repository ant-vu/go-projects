package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Suit represents a card suit
type Suit string

// Card suits
const (
	Hearts   Suit = "Hearts"
	Diamonds Suit = "Diamonds"
	Clubs    Suit = "Clubs"
	Spades   Suit = "Spades"
)

// Rank represents a card rank
type Rank string

// Card ranks
const (
	Ace   Rank = "Ace"
	Two   Rank = "Two"
	Three Rank = "Three"
	Four  Rank = "Four"
	Five  Rank = "Five"
	Six   Rank = "Six"
	Seven Rank = "Seven"
	Eight Rank = "Eight"
	Nine  Rank = "Nine"
	Ten   Rank = "Ten"
	Jack  Rank = "Jack"
	Queen Rank = "Queen"
	King  Rank = "King"
)

// Card represents a playing card
type Card struct {
	Suit Suit
	Rank Rank
}

// String returns a string representation of the card
func (c *Card) String() string {
	return fmt.Sprintf("%s of %s", c.Rank, c.Suit)
}

// Deck represents a deck of cards
type Deck struct {
	Cards []*Card
}

// NewDeck returns a new deck of cards
func NewDeck() *Deck {
	deck := &Deck{}
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	ranks := []Rank{Ace, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King}

	for _, suit := range suits {
		for _, rank := range ranks {
			deck.Cards = append(deck.Cards, &Card{Suit: suit, Rank: rank})
		}
	}

	return deck
}

// Shuffle shuffles the deck
func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// DealCard deals a card from the deck
func (d *Deck) DealCard() *Card {
	if len(d.Cards) == 0 {
		return nil
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card
}

// Player represents a player in the game
type Player struct {
	Hand []*Card
}

// NewPlayer returns a new player
func NewPlayer() *Player {
	return &Player{Hand: make([]*Card, 0)}
}

// DrawCard draws a card from the deck and adds it to the player's hand
func (p *Player) DrawCard(deck *Deck) {
	card := deck.DealCard()
	if card != nil {
		p.Hand = append(p.Hand, card)
	}
}

// PlayCard plays a card from the player's hand
func (p *Player) PlayCard(index int) *Card {
	if index < 0 || index >= len(p.Hand) {
		return nil
	}
	card := p.Hand[index]
	p.Hand = append(p.Hand[:index], p.Hand[index+1:]...)
	return card
}

// Game represents the card game
type Game struct {
	Deck       *Deck
	Discard    []*Card
	Players    []*Player
	CurrentPlayer int
}

// NewGame returns a new game
func NewGame(numPlayers int) *Game {
	game := &Game{
		Deck:       NewDeck(),
		Discard:    make([]*Card, 0),
		Players:    make([]*Player, numPlayers),
		CurrentPlayer: 0,
	}
	game.Deck.Shuffle()
	for i := range game.Players {
		game.Players[i] = NewPlayer()
		for j := 0; j < 5; j++ {
			game.Players[i].DrawCard(game.Deck)
		}
	}
	game.Discard = append(game.Discard, game.Deck.DealCard())
	return game
}

// PlayCard plays a card from the current player's hand
func (g *Game) PlayCard(index int) bool {
	topCard := g.Discard[len(g.Discard)-1]
	card := g.Players[g.CurrentPlayer].PlayCard(index)
	if card == nil {
		return false
	}
	if card.Suit == topCard.Suit || card.Rank == topCard.Rank {
		g.Discard = append(g.Discard, card)
		g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.Players)
		return true
	}
	g.Players[g.CurrentPlayer].Hand = append(g.Players[g.CurrentPlayer].Hand, card)
	return false
}

// DrawCard draws a card from the deck and adds it to the current player's hand
func (g *Game) DrawCard() {
	g.Players[g.CurrentPlayer].DrawCard(g.Deck)
}

// HasWon checks if the current player has won
func (g *Game) HasWon() bool {
	return len(g.Players[g.CurrentPlayer].Hand) == 0
}

func main() {
	game := NewGame(2)
	for {
		fmt.Printf("Player %d's turn\n", game.CurrentPlayer+1)
		fmt.Println("Hand:")
		for i, card := range game.Players[game.CurrentPlayer].Hand {
			fmt.Printf("%d: %s\n", i, card)
		}
		fmt.Println("Discard pile:")
		for _, card := range game.Discard {
			fmt.Println(card)
		}
		fmt.Printf("Cards left in deck: %d\n", len(game.Deck.Cards))
		var action string
		fmt.Print("Enter 'play <index>' to play a card, 'draw' to draw a card, or 'quit' to quit: ")
		fmt.Scanln(&action)
		if action == "quit" {
			break
		} else if action == "draw" {
			game.DrawCard()
			game.CurrentPlayer = (game.CurrentPlayer + 1) % len(game.Players)
		} else if len(action) > 4 && action[:4] == "play" {
			var index int
			fmt.Sscan(action[5:], &index)
			if !game.PlayCard(index) {
				fmt.Println("Invalid move. Try again.")
			}
			if game.HasWon() {
				fmt.Printf("Player %d wins!\n", game.CurrentPlayer+1)
				break
			}
		} else {
			fmt.Println("Invalid action. Try again.")
		}
	}
}