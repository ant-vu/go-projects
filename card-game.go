package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Suit string

const (
	Hearts   Suit = "Hearts"
	Diamonds Suit = "Diamonds"
	Clubs    Suit = "Clubs"
	Spades   Suit = "Spades"
)

type Rank string

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

type Card struct {
	Suit Suit
	Rank Rank
}

type Deck struct {
	Cards []*Card
}

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

func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) DealCard() *Card {
	if len(d.Cards) == 0 {
		return nil
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card
}

func main() {
	deck := NewDeck()
	deck.Shuffle()
	fmt.Println(deck.DealCard())
}