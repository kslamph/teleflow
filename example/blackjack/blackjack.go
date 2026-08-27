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
	// Using a more reliable image optimization service with explicit resizing to 100x140
	ImageProxyURL = "https://wsrv.nl/?url=&w=100&h=140&fit=contain"
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
	// Use wsrv.nl (a more reliable version of images.weserv.nl) to optimize images
	originalURL := fmt.Sprintf("%s%s_%s.png", CardRepoURL, strings.ToLower(c.Suit), rank)
	return ImageProxyURL + originalURL
}

// BlackjackData holds the state for a single user's game.
type BlackjackData struct {
	Deck             []Card
	PlayerHand       []Card
	DealerHand       []Card
	PlayerMessageIDs []int // To keep track of player card messages
	DealerMessageIDs []int // To keep track of dealer card messages
	MessageIDs       []int // To keep track of other messages to delete later
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

// updateHand updates photos of the cards by editing existing messages or sending new ones.
func updateHand(ctx *teleflow.Context, title string, hand []Card, showHoleCard bool, existingMessageIDs []int) ([]int, error) {
	if len(hand) == 0 {
		// Delete any existing messages if hand is empty
		for _, id := range existingMessageIDs {
			ctx.DeleteMessage(id)
		}
		return []int{}, nil
	}

	// If no existing messages, send new ones
	if len(existingMessageIDs) == 0 {
		return sendHand(ctx, title, hand, showHoleCard)
	}

	// Edit existing messages and send new ones if needed
	messageIDs := make([]int, len(hand))

	// Edit existing messages
	for i := 0; i < len(hand) && i < len(existingMessageIDs); i++ {
		var photoURL string
		if i == 0 && !showHoleCard { // Dealer's first card is hidden
			// Use wsrv.nl to optimize the back card image
			originalURL := CardRepoURL + "back_dark.png"
			photoURL = ImageProxyURL + originalURL
		} else {
			photoURL = hand[i].ImageURL()
		}

		// Create media object for editing
		media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(photoURL))
		if i == 0 {
			media.Caption = title
		}

		// Try to edit the existing message
		err := ctx.EditMessageMedia(existingMessageIDs[i], media)
		if err != nil {
			// If editing fails, fall back to sending a new message
			photo := tgbotapi.NewPhoto(ctx.ChatID(), tgbotapi.FileURL(photoURL))
			if i == 0 {
				photo.Caption = title
			}
			msg, sendErr := ctx.API().Send(photo)
			if sendErr != nil {
				return nil, sendErr
			}
			messageIDs[i] = msg.MessageID
		} else {
			messageIDs[i] = existingMessageIDs[i]
		}
	}

	// Send new messages for additional cards
	for i := len(existingMessageIDs); i < len(hand); i++ {
		var photoURL string
		if i == 0 && !showHoleCard { // Dealer's first card is hidden
			// Use wsrv.nl to optimize the back card image
			originalURL := CardRepoURL + "back_dark.png"
			photoURL = ImageProxyURL + originalURL
		} else {
			photoURL = hand[i].ImageURL()
		}

		photo := tgbotapi.NewPhoto(ctx.ChatID(), tgbotapi.FileURL(photoURL))
		if i == 0 {
			photo.Caption = title
		}
		msg, err := ctx.API().Send(photo)
		if err != nil {
			// Fallback to original image URL
			originalURL := strings.TrimPrefix(photoURL, ImageProxyURL)
			photoFallback := tgbotapi.NewPhoto(ctx.ChatID(), tgbotapi.FileURL(originalURL))
			if i == 0 {
				photoFallback.Caption = title
			}
			msg, err = ctx.API().Send(photoFallback)
			if err != nil {
				return nil, err
			}
		}
		messageIDs[i] = msg.MessageID
	}

	// Delete excess messages
	for i := len(hand); i < len(existingMessageIDs); i++ {
		ctx.DeleteMessage(existingMessageIDs[i])
	}

	return messageIDs, nil
}

// sendHand sends photos of the cards as a media group and returns their message IDs.
func sendHand(ctx *teleflow.Context, title string, hand []Card, showHoleCard bool) ([]int, error) {
	if len(hand) == 0 {
		return []int{}, nil
	}

	media := make([]interface{}, len(hand))
	for i, card := range hand {
		var photoURL string
		if i == 0 && !showHoleCard { // Dealer's first card is hidden
			// Use wsrv.nl to optimize the back card image
			originalURL := CardRepoURL + "back_dark.png"
			photoURL = ImageProxyURL + originalURL
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
		// Try sending with optimized image first, fallback to original if it fails
		msg, err := ctx.API().Send(photo)
		if err != nil {
			// Fallback to original image URL
			originalURL := strings.TrimPrefix(hand[0].ImageURL(), ImageProxyURL)
			photoFallback := tgbotapi.NewPhoto(ctx.ChatID(), tgbotapi.FileURL(originalURL))
			photoFallback.Caption = title
			msg, err = ctx.API().Send(photoFallback)
			if err != nil {
				return nil, err
			}
		}
		return []int{msg.MessageID}, nil
	}

	// Try sending media group with optimized images first
	msgs, err := ctx.API().SendMediaGroup(tgbotapi.NewMediaGroup(ctx.ChatID(), media))
	if err != nil {
		// Fallback to original image URLs
		mediaFallback := make([]interface{}, len(hand))
		for i, card := range hand {
			var photoURL string
			if i == 0 && !showHoleCard { // Dealer's first card is hidden
				photoURL = CardRepoURL + "back_dark.png"
			} else {
				// Get original URL by removing proxy prefix
				photoURL = strings.TrimPrefix(card.ImageURL(), ImageProxyURL)
			}

			p := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(photoURL))
			if i == 0 {
				p.Caption = title
			}
			mediaFallback[i] = p
		}
		msgs, err = ctx.API().SendMediaGroup(tgbotapi.NewMediaGroup(ctx.ChatID(), mediaFallback))
		if err != nil {
			return nil, err
		}
	}

	messageIDs := make([]int, len(msgs))
	for i, msg := range msgs {
		messageIDs[i] = msg.MessageID
	}
	return messageIDs, nil
}

// clearPreviousCards deletes the messages for the previous hand.
func clearPreviousCards(ctx *teleflow.Context, data *BlackjackData) {
	// Clear player card messages
	for _, id := range data.PlayerMessageIDs {
		ctx.DeleteMessage(id)
	}
	data.PlayerMessageIDs = []int{}

	// Clear dealer card messages
	for _, id := range data.DealerMessageIDs {
		ctx.DeleteMessage(id)
	}
	data.DealerMessageIDs = []int{}

	// Clear other messages
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
				data.DealerMessageIDs = dealerIDs
				data.PlayerMessageIDs = playerIDs

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

				// Update player's hand with the new card
				playerIDs, _ := updateHand(ctx, fmt.Sprintf("You drew a %s", newCard.String()), data.PlayerHand, true, data.PlayerMessageIDs)
				data.PlayerMessageIDs = playerIDs

				if calculateHandValue(data.PlayerHand) > 21 {
					ctx.ReplyText("Bust! You lose.")
					return teleflow.GoTo("EndGame"), nil
				}
				return teleflow.Retry(), nil // Re-prompt for another action
			case "stand":
				ctx.ReplyText("Dealer's turn...")
				clearPreviousCards(ctx, data)

				// Show hands with dealer's card revealed
				dealerIDs, _ := updateHand(ctx, "Dealer's Hand:", data.DealerHand, true, data.DealerMessageIDs)
				playerIDs, _ := updateHand(ctx, "Your Hand:", data.PlayerHand, true, data.PlayerMessageIDs)
				data.DealerMessageIDs = dealerIDs
				data.PlayerMessageIDs = playerIDs

				time.Sleep(1 * time.Second)

				for calculateHandValue(data.DealerHand) < 17 {
					newCard := deal(&data.Deck)
					data.DealerHand = append(data.DealerHand, newCard)
					// Update dealer's hand with the new card
					dealerIDs, _ := updateHand(ctx, fmt.Sprintf("Dealer draws a %s", newCard.String()), data.DealerHand, true, data.DealerMessageIDs)
					data.DealerMessageIDs = dealerIDs
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

			// Show the final hands for clarity.
			dealerIDs, _ := updateHand(ctx, "Dealer's Final Hand:", data.DealerHand, true, data.DealerMessageIDs)
			playerIDs, _ := updateHand(ctx, "Your Final Hand:", data.PlayerHand, true, data.PlayerMessageIDs)
			data.DealerMessageIDs = dealerIDs
			data.PlayerMessageIDs = playerIDs

			keyboard := teleflow.NewInlineKeyboard().AddRow(
				teleflow.ButtonCallback("Yes", "play_again"),
				teleflow.ButtonCallback("No", "quit"),
			).Build()

			return ctx.ReplyWithKeyboard(finalMsg+"\nPlay again?", keyboard)
		},
		Process: func(ctx *teleflow.Context, data *BlackjackData) (teleflow.FlowAction, error) {
			ctx.AnswerCallback()
			switch ctx.Text() {
			case "play_again":
				// Clear card messages and reset hands to trigger new deal in PlayerTurn prompt
				clearPreviousCards(ctx, data)
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

	for _, s := range []teleflow.Step[BlackjackData]{playerTurnStep, endGameStep} {
		if _, err := blackjackFlow.AddStep(s); err != nil {
			log.Fatal(err)
		}
	}

	router := teleflow.NewRouter()
	router.Handle("blackjack", blackjackFlow)
	router.Default(teleflow.Func(func(ctx *teleflow.Context) error {
		if blackjackFlow.IsUserInFlow(ctx.ChatID(), ctx.UserID()) {
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
