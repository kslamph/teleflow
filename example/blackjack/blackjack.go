package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	teleflow "github.com/kslamph/teleflow/core"
)

// --- Game Logic and State ---

const (
	CardRepoURL = "https://raw.githubusercontent.com/hanhaechi/playing-cards/master/"
)

type Card struct {
	Suit  string
	Rank  string
	Value int
}

func (c Card) String() string {
	return fmt.Sprintf("%s of %s", c.Rank, c.Suit)
}

func (c Card) ImageURL() string {
	rank := c.Rank
	if rank == "10" { // The repo uses '10' for the rank, not T
		// No change needed
	} else if len(rank) > 1 {
		// Ace, King, Queen, Jack -> A, K, Q, J
		rank = string(rank[0])
	}
	return fmt.Sprintf("%s%s_%s.png", CardRepoURL, strings.ToLower(c.Suit), rank)
}

// BlackjackData holds the state for a single user's game.
type BlackjackData struct {
	Deck       []Card
	PlayerHand []Card
	DealerHand []Card
	MessageIDs []int // To keep track of card messages to delete later
}

func newDeck() []Card {
	suits := []string{"Spades", "Hearts", "Diamonds", "Clubs"}
	ranks := map[string]int{"2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8, "9": 9, "10": 10, "Jack": 10, "Queen": 10, "King": 10, "Ace": 11}
	deck := make([]Card, 0, 52)

	for _, suit := range suits {
		for rank, value := range ranks {
			deck = append(deck, Card{Suit: suit, Rank: rank, Value: value})
		}
	}
	return deck
}

func shuffle(deck []Card) {

	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func deal(deck *[]Card) Card {
	card := (*deck)[0]
	*deck = (*deck)[1:]
	return card
}

func calculateHandValue(hand []Card) int {
	value := 0
	numAces := 0
	for _, card := range hand {
		value += card.Value
		if card.Rank == "Ace" {
			numAces++
		}
	}
	// Adjust for Aces
	for value > 21 && numAces > 0 {
		value -= 10
		numAces--
	}
	return value
}

// --- Bot and Flow Setup ---

// sendHand sends photos of the cards as a media group and returns their message IDs.
func sendHand(ctx *teleflow.Context, title string, hand []Card, showHoleCard bool) ([]int, error) {
	if len(hand) == 0 {
		return []int{}, nil
	}

	media := make([]interface{}, len(hand))
	for i, card := range hand {
		var photoURL string
		if i == 0 && !showHoleCard { // Dealer's first card is hidden
			photoURL = CardRepoURL + "back_dark.png"
		} else {
			photoURL = card.ImageURL()
		}

		p := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(photoURL))
		if i == 0 {
			p.Caption = title
		}
		media[i] = p
	}

	// For a single card (e.g., when hitting), send as a single photo to avoid media group overhead.
	if len(media) == 1 {
		photo := tgbotapi.NewPhoto(ctx.ChatID(), tgbotapi.FileURL(hand[0].ImageURL()))
		photo.Caption = title
		msg, err := ctx.API().Send(photo)
		if err != nil {
			return nil, err
		}
		return []int{msg.MessageID}, nil
	}

	msgs, err := ctx.API().SendMediaGroup(tgbotapi.NewMediaGroup(ctx.ChatID(), media))
	if err != nil {
		return nil, err
	}

	messageIDs := make([]int, len(msgs))
	for i, msg := range msgs {
		messageIDs[i] = msg.MessageID
	}
	return messageIDs, nil
}

// clearPreviousCards deletes the messages for the previous hand.
func clearPreviousCards(ctx *teleflow.Context, data *BlackjackData) {
	for _, id := range data.MessageIDs {
		ctx.DeleteMessage(id)
	}
	data.MessageIDs = []int{}
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	blackjackFlow := teleflow.NewFlow[BlackjackData]()

	// 1. Player Turn Step: Deals cards (if needed) and asks the player to Hit or Stand.
	playerTurnStep := teleflow.Step[BlackjackData]{
		Name: "PlayerTurn",
		Prompt: func(ctx *teleflow.Context, data *BlackjackData) error {
			// If hands are empty, it's a new game or a restart.
			if len(data.PlayerHand) == 0 {
				ctx.ReplyText("Shuffling and dealing...")
				clearPreviousCards(ctx, data)

				deck := newDeck()
				shuffle(deck)
				data.Deck = deck
				data.PlayerHand = []Card{deal(&data.Deck), deal(&data.Deck)}
				data.DealerHand = []Card{deal(&data.Deck), deal(&data.Deck)}

				dealerIDs, _ := sendHand(ctx, "Dealer's Hand:", data.DealerHand, false)
				playerIDs, _ := sendHand(ctx, "Your Hand:", data.PlayerHand, true)
				data.MessageIDs = append(data.MessageIDs, dealerIDs...)
				data.MessageIDs = append(data.MessageIDs, playerIDs...)

				// Check for immediate blackjack
				if calculateHandValue(data.PlayerHand) == 21 || calculateHandValue(data.DealerHand) == 21 {
					// The Prompt for EndGame will handle displaying the result.
					blackjackFlow.HandleUpdate(ctx) // This will trigger EndGame's prompt
					return nil                      // Stop this prompt, as EndGame will take over.
				}
			}

			playerValue := calculateHandValue(data.PlayerHand)
			prompt := fmt.Sprintf("Your hand value is %d. Hit or Stand?", playerValue)
			keyboard := teleflow.NewInlineKeyboard().AddRow(
				teleflow.ButtonCallback("Hit", "hit"),
				teleflow.ButtonCallback("Stand", "stand"),
			).Build()

			msg := tgbotapi.NewMessage(ctx.ChatID(), prompt)
			msg.ReplyMarkup = keyboard
			sentMsg, err := ctx.API().Send(msg)
			if err == nil {
				data.MessageIDs = append(data.MessageIDs, sentMsg.MessageID)
			}
			return err
		},
		Process: func(ctx *teleflow.Context, data *BlackjackData) (teleflow.FlowAction, error) {
			// This process function now handles both the check after dealing and player actions.
			// If hands have just been dealt and there was a blackjack, move to the end.
			if len(data.PlayerHand) == 2 && (calculateHandValue(data.PlayerHand) == 21 || calculateHandValue(data.DealerHand) == 21) {
				return teleflow.GoTo("EndGame"), nil
			}

			// Otherwise, process player's action from the inline keyboard.
			ctx.AnswerCallback()
			switch ctx.Text() {
			case "hit":
				newCard := deal(&data.Deck)
				data.PlayerHand = append(data.PlayerHand, newCard)

				newCardIDs, _ := sendHand(ctx, fmt.Sprintf("You drew a %s", newCard.String()), []Card{newCard}, true)
				data.MessageIDs = append(data.MessageIDs, newCardIDs...)

				if calculateHandValue(data.PlayerHand) > 21 {
					ctx.ReplyText("Bust! You lose.")
					return teleflow.GoTo("EndGame"), nil
				}
				return teleflow.Retry(), nil // Re-prompt for another action
			case "stand":
				ctx.ReplyText("Dealer's turn...")
				clearPreviousCards(ctx, data)

				// Show hands with dealer's card revealed
				dealerIDs, _ := sendHand(ctx, "Dealer's Hand:", data.DealerHand, true)
				playerIDs, _ := sendHand(ctx, "Your Hand:", data.PlayerHand, true)
				data.MessageIDs = append(data.MessageIDs, dealerIDs...)
				data.MessageIDs = append(data.MessageIDs, playerIDs...)

				time.Sleep(1 * time.Second)

				for calculateHandValue(data.DealerHand) < 17 {
					newCard := deal(&data.Deck)
					data.DealerHand = append(data.DealerHand, newCard)
					newCardIDs, _ := sendHand(ctx, fmt.Sprintf("Dealer draws a %s", newCard.String()), []Card{newCard}, true)
					data.MessageIDs = append(data.MessageIDs, newCardIDs...)
					time.Sleep(1 * time.Second) // Pause for dramatic effect
				}
				return teleflow.GoTo("EndGame"), nil
			default:
				// This case can be hit if the Process is triggered right after dealing.
				// We do nothing and let the Prompt handle the next step.
				return teleflow.Retry(), nil
			}
		},
	}

	// 4. End Game Step: Determines the winner and asks to play again.
	endGameStep := teleflow.Step[BlackjackData]{
		Name: "EndGame",
		Prompt: func(ctx *teleflow.Context, data *BlackjackData) error {
			playerValue := calculateHandValue(data.PlayerHand)
			dealerValue := calculateHandValue(data.DealerHand)

			result := ""
			if playerValue > 21 {
				result = "You busted! Dealer wins."
			} else if dealerValue > 21 {
				result = "Dealer busted! You win!"
			} else if playerValue == dealerValue {
				result = "Push! It's a tie."
			} else if playerValue > dealerValue {
				result = "You win!"
			} else {
				result = "Dealer wins."
			}

			finalMsg := fmt.Sprintf("Game over! Your hand: %d, Dealer's hand: %d. %s", playerValue, dealerValue, result)

			// Always clear previous cards and show the final hands for clarity.
			clearPreviousCards(ctx, data)
			dealerIDs, _ := sendHand(ctx, "Dealer's Final Hand:", data.DealerHand, true)
			playerIDs, _ := sendHand(ctx, "Your Final Hand:", data.PlayerHand, true)
			data.MessageIDs = append(data.MessageIDs, dealerIDs...)
			data.MessageIDs = append(data.MessageIDs, playerIDs...)

			keyboard := teleflow.NewInlineKeyboard().AddRow(
				teleflow.ButtonCallback("Yes", "play_again"),
				teleflow.ButtonCallback("No", "quit"),
			).Build()

			return ctx.ReplyInlineKeyboard(finalMsg+"\nPlay again?", keyboard)
		},
		Process: func(ctx *teleflow.Context, data *BlackjackData) (teleflow.FlowAction, error) {
			ctx.AnswerCallback()
			switch ctx.Text() {
			case "play_again":
				// Reset hands to trigger new deal in PlayerTurn prompt
				data.PlayerHand = []Card{}
				data.DealerHand = []Card{}
				return teleflow.GoTo("PlayerTurn"), nil
			case "quit":
				ctx.EditMessageText(ctx.Update().CallbackQuery.Message.MessageID, "Thanks for playing!")
				return teleflow.End(), nil
			}
			return teleflow.Retry(), nil
		},
	}

	blackjackFlow.
		AddStep(playerTurnStep).
		AddStep(endGameStep)

	router := teleflow.NewRouter()
	router.Handle("blackjack", blackjackFlow)
	router.Default(teleflow.Func(func(ctx *teleflow.Context) error {
		if blackjackFlow.IsUserInFlow(ctx.UserID()) {
			return blackjackFlow.HandleUpdate(ctx)
		}
		return ctx.ReplyText("Unknown command. Use /blackjack to play.")
	}))

	bot, err := teleflow.NewBot(token, router)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	bot.SetBotCommands(map[string]string{
		"blackjack": "Play a game of Blackjack",
	})

	log.Println("Blackjack bot starting...")
	log.Fatal(bot.Start())
}
