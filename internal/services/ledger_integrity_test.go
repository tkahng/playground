package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestLedgerService_AvailableBalance_DecreasesWithPendingHold(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		// Re-fetch wallet to get updated counters.
		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 200 {
			t.Errorf("Balance() = %d, want 200", wallet.Balance())
		}
		if wallet.AvailableBalance() != 150 {
			t.Errorf("AvailableBalance() = %d, want 150", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_AvailableBalance_RestoresAfterVoid(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		pending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		if _, err = ledger.VoidPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("VoidPendingTransfer: %v", err)
		}

		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 200 {
			t.Errorf("Balance() = %d, want 200 after void", wallet.Balance())
		}
		if wallet.AvailableBalance() != 200 {
			t.Errorf("AvailableBalance() = %d, want 200 after void", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_AvailableBalance_DecreasesAfterPendingPost(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		pending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		if _, err = ledger.PostPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("PostPendingTransfer: %v", err)
		}

		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 150 {
			t.Errorf("Balance() = %d, want 150 after post", wallet.Balance())
		}
		if wallet.AvailableBalance() != 150 {
			t.Errorf("AvailableBalance() = %d, want 150 after post", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_MoneyConservation_BetSettle_HostWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		// Record escrow balance before bet.
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount escrow: %v", err)
		}
		escrowBefore := escrow.Balance()
		hostBalanceBefore, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance hostBefore: %v", err)
		}
		guestBalanceBefore, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance guestBefore: %v", err)
		}
		totalBefore := hostBalanceBefore + guestBalanceBefore + escrowBefore

		// Place host bet.
		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		// Settle: host wins.
		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             betAmount,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultWin,
			GuestResult:           models.RpsParticipantResultLose,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle: %v", err)
		}

		// Assert final balances.
		hostBalance, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBalance, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}

		if hostBalance != 600 {
			t.Errorf("host balance = %d, want 600 (500 + 100 winnings)", hostBalance)
		}
		if guestBalance != 400 {
			t.Errorf("guest balance = %d, want 400 (500 - 100 bet)", guestBalance)
		}

		// Escrow must be zero net (no funds stranded).
		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		if escrow.Balance() != escrowBefore {
			t.Errorf("escrow net balance = %d, want %d (must return to prior state)", escrow.Balance(), escrowBefore)
		}
		totalAfter := hostBalance + guestBalance + escrow.Balance()
		if totalAfter != totalBefore {
			t.Errorf("total system balance = %d, want %d (money conservation violated)", totalAfter, totalBefore)
		}
	})
}

func TestLedgerService_MoneyConservation_BetSettle_Tie(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount escrow: %v", err)
		}
		escrowBefore := escrow.Balance()
		hostBalanceBefore, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance hostBefore: %v", err)
		}
		guestBalanceBefore, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance guestBefore: %v", err)
		}
		totalBefore := hostBalanceBefore + guestBalanceBefore + escrowBefore

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             betAmount,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultTie,
			GuestResult:           models.RpsParticipantResultTie,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle: %v", err)
		}

		hostBalance, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBalance, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}

		if hostBalance != 500 {
			t.Errorf("host balance = %d, want 500 (tie: refunded)", hostBalance)
		}
		if guestBalance != 500 {
			t.Errorf("guest balance = %d, want 500 (tie: refunded)", guestBalance)
		}

		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		if escrow.Balance() != escrowBefore {
			t.Errorf("escrow balance = %d, want %d", escrow.Balance(), escrowBefore)
		}
		totalAfter := hostBalance + guestBalance + escrow.Balance()
		if totalAfter != totalBefore {
			t.Errorf("total system balance = %d, want %d (money conservation violated)", totalAfter, totalBefore)
		}
	})
}

func TestLedgerService_MoneyConservation_BetRefund_BothVoided(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		// Place host bet, then place guest bet.
		guestWallet, err := ledger.GetOrCreateUserWallet(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet guest: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount escrow: %v", err)
		}
		escrowBefore := escrow.Balance()
		hostBalanceBefore, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance hostBefore: %v", err)
		}
		guestBalanceBefore, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance guestBefore: %v", err)
		}
		totalBefore := hostBalanceBefore + guestBalanceBefore + escrowBefore

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		// Create guest pending manually (simulates guest escrow placed before game result).
		guestPending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  guestWallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          betAmount,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer guest: %v", err)
		}

		// Refund both.
		if err := betting.RefundBothBets(ctx, hostPending.ID, guestPending.ID); err != nil {
			t.Fatalf("RefundBothBets: %v", err)
		}

		// Both balances must return to 500.
		hostBalance, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBalance, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}
		if hostBalance != 500 {
			t.Errorf("host balance = %d, want 500 after full refund", hostBalance)
		}
		if guestBalance != 500 {
			t.Errorf("guest balance = %d, want 500 after full refund", guestBalance)
		}

		// Available balances also 500 (no pending holds).
		hostAvail, err := ledger.GetUserAvailableBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance host: %v", err)
		}
		guestAvail, err := ledger.GetUserAvailableBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance guest: %v", err)
		}
		if hostAvail != 500 {
			t.Errorf("host available = %d, want 500", hostAvail)
		}
		if guestAvail != 500 {
			t.Errorf("guest available = %d, want 500", guestAvail)
		}

		// Escrow must return to its prior balance (net=0 for this game).
		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		if escrow.Balance() != escrowBefore {
			t.Errorf("escrow balance = %d, want %d (escrow must net to zero)", escrow.Balance(), escrowBefore)
		}
		totalAfter := hostBalance + guestBalance + escrow.Balance()
		if totalAfter != totalBefore {
			t.Errorf("total system balance = %d, want %d (money conservation violated)", totalAfter, totalBefore)
		}
	})
}
