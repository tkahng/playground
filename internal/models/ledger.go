package models

import (
	"time"

	"github.com/google/uuid"
)

// LedgerAccountFlags is a bitfield controlling account behaviour.
const (
	// AccountFlagDebitsMustNotExceedCredits prevents overdraft.
	// Use on user-facing wallet accounts.
	AccountFlagDebitsMustNotExceedCredits int = 1 << 0
	// AccountFlagCreditsMustNotExceedDebits is the inverse constraint.
	// Use on system issuance accounts.
	AccountFlagCreditsMustNotExceedDebits int = 1 << 1
)

// LedgerAccount is a node in the double-entry ledger.
//
// Balances are tracked as cumulative sums rather than a single mutable column
// to produce an unambiguous audit trail and support two-phase (pending) transfers.
//
//   - Balance           = credits_posted - debits_posted
//   - Available balance = Balance - debits_pending
type LedgerAccount struct {
	_              struct{}   `db:"accounts" schema:"ledger" json:"-"`
	ID             uuid.UUID  `db:"id,pk" json:"id"`
	Code           string     `db:"code" json:"code"`
	EntityType     string     `db:"entity_type" json:"entity_type"`
	EntityID       *uuid.UUID `db:"entity_id" json:"entity_id,omitempty"`
	LedgerCode     string     `db:"ledger_code" json:"ledger_code"`
	Flags          int        `db:"flags" json:"flags"`
	DebitsPending  int64      `db:"debits_pending" json:"debits_pending"`
	CreditsPending int64      `db:"credits_pending" json:"credits_pending"`
	DebitsPosted   int64      `db:"debits_posted" json:"debits_posted"`
	CreditsPosted  int64      `db:"credits_posted" json:"credits_posted"`
	Metadata       []byte     `db:"metadata" json:"metadata"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

// Balance returns the net settled balance (credits_posted - debits_posted).
func (a *LedgerAccount) Balance() int64 {
	return a.CreditsPosted - a.DebitsPosted
}

// AvailableBalance returns the balance minus any pending holds.
func (a *LedgerAccount) AvailableBalance() int64 {
	return a.Balance() - a.DebitsPending
}

// LedgerTransferStatus is the lifecycle state of a transfer.
type LedgerTransferStatus string

const (
	LedgerTransferStatusPending LedgerTransferStatus = "pending"
	LedgerTransferStatusPosted  LedgerTransferStatus = "posted"
	LedgerTransferStatusVoided  LedgerTransferStatus = "voided"
)

// Transfer-flag bit positions.
const (
	// TransferFlagPending marks a two-phase pending transfer (holds available balance).
	TransferFlagPending int = 1 << 0
	// TransferFlagPostPending converts a pending transfer to posted.
	TransferFlagPostPending int = 1 << 1
	// TransferFlagVoidPending releases a pending transfer with no net effect.
	TransferFlagVoidPending int = 1 << 2
	// TransferFlagLinked marks a transfer as part of an atomic linked group.
	TransferFlagLinked int = 1 << 3
)

// Well-known transfer codes.
const (
	TransferCodePurchase  = "purchase"   // Points bought via Stripe
	TransferCodeBetEscrow = "bet_escrow" // Bet placed, funds moved to escrow (pending)
	TransferCodeBetWin    = "bet_win"    // Winner receives escrow funds
	TransferCodeBetRefund = "bet_refund" // Tie: escrow returned to each player
	TransferCodeBetVoid   = "bet_void"   // Void a pending bet (decline / expiry)
)

// Well-known reference types.
const (
	ReferenceTypeRpsGame        = "rps_game"
	ReferenceTypeStripeCheckout = "stripe_checkout"
)

// Well-known system account codes.
const (
	SystemAccountPointsIssuance = "system:points_issuance"
	SystemAccountGameEscrow     = "system:game_escrow"
)

// UserWalletCode returns the ledger account code for a user's points wallet.
func UserWalletCode(userID uuid.UUID) string {
	return "user:" + userID.String() + ":wallet"
}

// LedgerTransfer represents a single value movement between two accounts.
// Every transfer debits one account and credits another.
type LedgerTransfer struct {
	_               struct{}             `db:"transfers" schema:"ledger" json:"-"`
	ID              uuid.UUID            `db:"id,pk" json:"id"`
	LedgerCode      string               `db:"ledger_code" json:"ledger_code"`
	DebitAccountID  uuid.UUID            `db:"debit_account_id" json:"debit_account_id"`
	CreditAccountID uuid.UUID            `db:"credit_account_id" json:"credit_account_id"`
	Amount          int64                `db:"amount" json:"amount"`
	PendingID       *uuid.UUID           `db:"pending_id" json:"pending_id,omitempty"`
	Flags           int                  `db:"flags" json:"flags"`
	Status          LedgerTransferStatus `db:"status" json:"status"`
	TransferCode    string               `db:"transfer_code" json:"transfer_code"`
	ReferenceType   *string              `db:"reference_type" json:"reference_type,omitempty"`
	ReferenceID     *uuid.UUID           `db:"reference_id" json:"reference_id,omitempty"`
	TimeoutAt       *time.Time           `db:"timeout_at" json:"timeout_at,omitempty"`
	Metadata        []byte               `db:"metadata" json:"metadata"`
	CreatedAt       time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `db:"updated_at" json:"updated_at"`
	// Relations (not persisted; populated by queries)
	DebitAccount  *LedgerAccount `db:"debit_account" src:"debit_account_id" dest:"id" table:"ledger.accounts" json:"debit_account,omitempty"`
	CreditAccount *LedgerAccount `db:"credit_account" src:"credit_account_id" dest:"id" table:"ledger.accounts" json:"credit_account,omitempty"`
}
