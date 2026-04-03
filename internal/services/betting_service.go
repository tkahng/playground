package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

// BettingService orchestrates the wagering workflow for RPS games.
// All methods that mutate ledger state must be called inside a transaction.
type BettingService interface {
	// PlaceHostBet creates a pending escrow transfer for the host player.
	// Returns the pending LedgerTransfer whose ID should be stored on the game.
	PlaceHostBet(ctx context.Context, gameID, hostUserID uuid.UUID, amount int64) (*models.LedgerTransfer, error)

	// PlaceGuestAndSettle creates the guest's pending escrow transfer, posts both
	// pending holds, then distributes the escrow based on the game result.
	// Returns the guest's pending transfer ID for audit trail storage.
	PlaceGuestAndSettle(ctx context.Context, input PlaceGuestAndSettleInput) (uuid.UUID, error)

	// EnsureGuestCanAffordBet returns an error if the guest's available balance
	// is less than amount. Must be called before PlaceGuestAndSettle.
	EnsureGuestCanAffordBet(ctx context.Context, guestUserID uuid.UUID, amount int64) error

	// RefundHostBet voids the host's pending bet (e.g., guest declined or game expired).
	RefundHostBet(ctx context.Context, hostPendingTransferID uuid.UUID) error

	// RefundBothBets voids both pending bets (edge case: both placed but game expired).
	RefundBothBets(ctx context.Context, hostPendingTransferID, guestPendingTransferID uuid.UUID) error
}

// PlaceGuestAndSettleInput carries all data needed to finalise a bet when the guest responds.
type PlaceGuestAndSettleInput struct {
	GameID                 uuid.UUID
	GuestUserID            uuid.UUID
	HostUserID             uuid.UUID
	BetAmount              int64
	HostPendingTransferID  uuid.UUID
	HostResult             models.RpsParticipantResult
	GuestResult            models.RpsParticipantResult
}

type DbBettingService struct {
	adapter stores.StorageAdapterInterface
	ledger  LedgerService
}

var _ BettingService = (*DbBettingService)(nil)

func NewDbBettingService(adapter stores.StorageAdapterInterface, ledger LedgerService) *DbBettingService {
	return &DbBettingService{adapter: adapter, ledger: ledger}
}

// PlaceHostBet creates a pending escrow transfer for the host.
// Must be called inside a transaction.
func (s *DbBettingService) PlaceHostBet(ctx context.Context, gameID, hostUserID uuid.UUID, amount int64) (*models.LedgerTransfer, error) {
	hostWallet, err := s.ledger.GetOrCreateUserWallet(ctx, hostUserID)
	if err != nil {
		return nil, fmt.Errorf("host wallet: %w", err)
	}

	escrow, err := s.ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
	if err != nil {
		return nil, fmt.Errorf("escrow account: %w", err)
	}

	refType := models.ReferenceTypeRpsGame
	pending, err := s.ledger.CreatePendingTransfer(ctx, PostTransferInput{
		LedgerCode:      "POINTS",
		DebitAccountID:  hostWallet.ID,
		CreditAccountID: escrow.ID,
		Amount:          amount,
		TransferCode:    models.TransferCodeBetEscrow,
		ReferenceType:   &refType,
		ReferenceID:     &gameID,
	})
	if err != nil {
		return nil, fmt.Errorf("place host bet: %w", err)
	}
	return pending, nil
}

// PlaceGuestAndSettle finalises the bet: places guest's escrow, posts both holds, distributes funds.
// Returns the guest's pending transfer ID for audit trail storage.
// Must be called inside a transaction.
func (s *DbBettingService) PlaceGuestAndSettle(ctx context.Context, input PlaceGuestAndSettleInput) (uuid.UUID, error) {
	if input.BetAmount <= 0 {
		return uuid.Nil, errors.New("bet amount must be positive")
	}

	guestWallet, err := s.ledger.GetOrCreateUserWallet(ctx, input.GuestUserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("guest wallet: %w", err)
	}
	hostWallet, err := s.ledger.GetOrCreateUserWallet(ctx, input.HostUserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("host wallet: %w", err)
	}
	escrow, err := s.ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
	if err != nil {
		return uuid.Nil, fmt.Errorf("escrow account: %w", err)
	}

	refType := models.ReferenceTypeRpsGame

	// Step 1: create guest's pending escrow transfer.
	guestPending, err := s.ledger.CreatePendingTransfer(ctx, PostTransferInput{
		LedgerCode:      "POINTS",
		DebitAccountID:  guestWallet.ID,
		CreditAccountID: escrow.ID,
		Amount:          input.BetAmount,
		TransferCode:    models.TransferCodeBetEscrow,
		ReferenceType:   &refType,
		ReferenceID:     &input.GameID,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("place guest bet: %w", err)
	}

	// Step 2: post both pending holds (funds are now fully in escrow).
	if _, err = s.ledger.PostPendingTransfer(ctx, input.HostPendingTransferID); err != nil {
		return uuid.Nil, fmt.Errorf("post host pending: %w", err)
	}
	if _, err = s.ledger.PostPendingTransfer(ctx, guestPending.ID); err != nil {
		return uuid.Nil, fmt.Errorf("post guest pending: %w", err)
	}

	// Step 3: distribute escrow according to game result.
	totalPot := input.BetAmount * 2

	switch {
	case input.HostResult == models.RpsParticipantResultWin:
		// Host wins: all pot goes to host.
		_, err = s.ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  escrow.ID,
			CreditAccountID: hostWallet.ID,
			Amount:          totalPot,
			TransferCode:    models.TransferCodeBetWin,
			ReferenceType:   &refType,
			ReferenceID:     &input.GameID,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("pay host: %w", err)
		}

	case input.GuestResult == models.RpsParticipantResultWin:
		// Guest wins: all pot goes to guest.
		_, err = s.ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  escrow.ID,
			CreditAccountID: guestWallet.ID,
			Amount:          totalPot,
			TransferCode:    models.TransferCodeBetWin,
			ReferenceType:   &refType,
			ReferenceID:     &input.GameID,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("pay guest: %w", err)
		}

	default:
		// Tie: refund each player their original stake.
		_, err = s.ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  escrow.ID,
			CreditAccountID: hostWallet.ID,
			Amount:          input.BetAmount,
			TransferCode:    models.TransferCodeBetRefund,
			ReferenceType:   &refType,
			ReferenceID:     &input.GameID,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("refund host: %w", err)
		}
		_, err = s.ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  escrow.ID,
			CreditAccountID: guestWallet.ID,
			Amount:          input.BetAmount,
			TransferCode:    models.TransferCodeBetRefund,
			ReferenceType:   &refType,
			ReferenceID:     &input.GameID,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("refund guest: %w", err)
		}
	}

	return guestPending.ID, nil
}

// EnsureGuestCanAffordBet returns an error if the guest's available balance is less than amount.
func (s *DbBettingService) EnsureGuestCanAffordBet(ctx context.Context, guestUserID uuid.UUID, amount int64) error {
	wallet, err := s.ledger.GetOrCreateUserWallet(ctx, guestUserID)
	if err != nil {
		return fmt.Errorf("guest wallet: %w", err)
	}
	if wallet.AvailableBalance() < amount {
		return fmt.Errorf("insufficient balance: need %d pts but have %d pts available", amount, wallet.AvailableBalance())
	}
	return nil
}

// RefundHostBet voids the host's pending escrow transfer.
// Must be called inside a transaction.
func (s *DbBettingService) RefundHostBet(ctx context.Context, hostPendingTransferID uuid.UUID) error {
	if _, err := s.ledger.VoidPendingTransfer(ctx, hostPendingTransferID); err != nil {
		return fmt.Errorf("void host bet: %w", err)
	}
	return nil
}

// RefundBothBets voids both pending escrow transfers.
// Must be called inside a transaction.
func (s *DbBettingService) RefundBothBets(ctx context.Context, hostPendingTransferID, guestPendingTransferID uuid.UUID) error {
	if _, err := s.ledger.VoidPendingTransfer(ctx, hostPendingTransferID); err != nil {
		return fmt.Errorf("void host bet: %w", err)
	}
	if _, err := s.ledger.VoidPendingTransfer(ctx, guestPendingTransferID); err != nil {
		return fmt.Errorf("void guest bet: %w", err)
	}
	return nil
}

// PointsPurchaseFulfillInput holds the data needed to credit a user after a Stripe purchase.
type PointsPurchaseFulfillInput struct {
	UserID          uuid.UUID
	PointsAmount    int64
	StripeSessionID string
}

// FulfillPointsPurchase credits a user's wallet after a successful Stripe checkout.
// Idempotent: skips if a transfer with the same reference already exists.
// Must be called inside a transaction.
func FulfillPointsPurchase(ctx context.Context, adapter stores.StorageAdapterInterface, ledger LedgerService, input PointsPurchaseFulfillInput) error {
	refType := models.ReferenceTypeStripeCheckout
	sessionID, err := uuid.Parse(input.StripeSessionID)
	if err != nil {
		// Stripe session IDs are not UUIDs; store as opaque metadata instead.
		// Use a deterministic UUID derived from the session ID for idempotency.
		sessionID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(input.StripeSessionID))
	}

	// Idempotency: check if this session has already been fulfilled.
	existing, err := adapter.Ledger().FindTransfer(ctx, &stores.LedgerTransferFilter{
		ReferenceTypes: []string{refType},
		ReferenceIds:   []uuid.UUID{sessionID},
		TransferCodes:  []string{models.TransferCodePurchase},
	})
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}
	if existing != nil {
		// Already fulfilled; this is a duplicate webhook delivery.
		return nil
	}

	issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
	if err != nil {
		return fmt.Errorf("issuance account: %w", err)
	}
	wallet, err := ledger.GetOrCreateUserWallet(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("user wallet: %w", err)
	}

	_, err = ledger.PostTransfer(ctx, PostTransferInput{
		LedgerCode:      "POINTS",
		DebitAccountID:  issuance.ID,
		CreditAccountID: wallet.ID,
		Amount:          input.PointsAmount,
		TransferCode:    models.TransferCodePurchase,
		ReferenceType:   &refType,
		ReferenceID:     types.Pointer(sessionID),
	})
	return err
}
