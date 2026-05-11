package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type DbGamingStoreDecorator struct {
	Delegate                  *DBGamingStore
	CountFriendshipsFunc      func(ctx context.Context, filter *FriendshipFilter) (int64, error)
	CountPlayersFunc          func(ctx context.Context, filter *PlayersFilter) (int64, error)
	CreateFriendshipFunc      func(ctx context.Context, friendship *models.Friendship) (*models.Friendship, error)
	CreatePlayerFunc          func(ctx context.Context, player *models.Player) (*models.Player, error)
	DeleteFriendshipsFunc     func(ctx context.Context, filter *FriendshipFilter) (int64, error)
	DeletePlayersFunc         func(ctx context.Context, filter *PlayersFilter) (int64, error)
	FindFriendshipFunc        func(ctx context.Context, filter *FriendshipFilter) (*models.Friendship, error)
	FindFriendshipsFunc       func(ctx context.Context, filter *FriendshipFilter) ([]*models.Friendship, error)
	FindPlayerFunc            func(ctx context.Context, filter *PlayersFilter) (*models.Player, error)
	FindPlayersFunc           func(ctx context.Context, filter *PlayersFilter) ([]*models.Player, error)
	UpdateFriendshipFunc      func(ctx context.Context, player *models.Friendship) (*models.Friendship, error)
	UpdatePlayerFunc          func(ctx context.Context, player *models.Player) (*models.Player, error)
	UpdatePlayerLastSeenFunc  func(ctx context.Context, playerID uuid.UUID) error
	CreateRpsGameFunc              func(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error)
	UpdateRpsGameFunc              func(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error)
	FindRpsGameFunc                func(ctx context.Context, filter *RpsGameFilter) (*models.RpsGame, error)
	FindRpsGamesFunc               func(ctx context.Context, filter *RpsGameFilter) ([]*models.RpsGame, error)
	CountRpsGamesFunc              func(ctx context.Context, filter *RpsGameFilter) (int64, error)
	FindRpsGameForUpdateFunc       func(ctx context.Context, gameID uuid.UUID) (*models.RpsGame, error)
	FindExpiredPendingBetGamesFunc func(ctx context.Context) ([]*models.RpsGame, error)
	CountRpsParticipantsFunc  func(ctx context.Context, filter *RpsParticipantFilter) (int64, error)
	CreateRpsParticipantFunc  func(ctx context.Context, participant *models.RpsParticipant) (*models.RpsParticipant, error)
	CreateRpsParticipantsFunc func(ctx context.Context, participant []*models.RpsParticipant) ([]*models.RpsParticipant, error)
	FindRpsParticipantFunc    func(ctx context.Context, filter *RpsParticipantFilter) (*models.RpsParticipant, error)
	FindRpsParticipantsFunc   func(ctx context.Context, filter *RpsParticipantFilter) ([]*models.RpsParticipant, error)
	UpdateRpsParticipantFunc  func(ctx context.Context, participant *models.RpsParticipant) (*models.RpsParticipant, error)
	CountRpsGameInvitesFunc   func(ctx context.Context, filter *RpsGameInviteFilter) (int64, error)
	CreateRpsGameInviteFunc   func(ctx context.Context, invite *models.RpsGameInvite) (*models.RpsGameInvite, error)
	DeleteRpGameInvitesFunc   func(ctx context.Context, filter *RpsGameInviteFilter) (int64, error)
	UpdateRpsGameInviteFunc   func(ctx context.Context, player *models.RpsGameInvite) (*models.RpsGameInvite, error)
	FindRpsGameInviteFunc     func(ctx context.Context, filter *RpsGameInviteFilter) (*models.RpsGameInvite, error)
	FindRpsGameInvitesFunc    func(ctx context.Context, filter *RpsGameInviteFilter) ([]*models.RpsGameInvite, error)

	CreateRpsRematchRequestFunc      func(ctx context.Context, req *models.RpsRematchRequest) (*models.RpsRematchRequest, error)
	FindRpsRematchRequestFunc        func(ctx context.Context, filter *RpsRematchFilter) (*models.RpsRematchRequest, error)
	UpdateRpsRematchRequestFunc      func(ctx context.Context, req *models.RpsRematchRequest) (*models.RpsRematchRequest, error)
	FindExpiredPendingRpsRematchesFunc func(ctx context.Context) ([]*models.RpsRematchRequest, error)

	WithTxFunc func(db database.Dbx) *DBGamingStore
}

func (s *DbGamingStoreDecorator) CountRpsGames(ctx context.Context, filter *RpsGameFilter) (int64, error) {
	if s.CountRpsGamesFunc != nil {
		return s.CountRpsGamesFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, fmt.Errorf("Gaming store decorator CountRpsGames %w", ErrDelegateNil)
	}
	return s.Delegate.CountRpsGames(ctx, filter)
}

// FindRpsGameInvite implements [GamingStore].
func (s *DbGamingStoreDecorator) FindRpsGameInvite(ctx context.Context, filter *RpsGameInviteFilter) (*models.RpsGameInvite, error) {
	if s.FindRpsGameInviteFunc != nil {
		return s.FindRpsGameInviteFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindRpsGameInvite %w", ErrDelegateNil)
	}
	return s.Delegate.FindRpsGameInvite(ctx, filter)
}

// FindRpsGameInvites implements [GamingStore].
func (s *DbGamingStoreDecorator) FindRpsGameInvites(ctx context.Context, filter *RpsGameInviteFilter) ([]*models.RpsGameInvite, error) {
	if s.FindRpsGameInvitesFunc != nil {
		return s.FindRpsGameInvitesFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindRpsGameInvites %w", ErrDelegateNil)
	}
	return s.Delegate.FindRpsGameInvites(ctx, filter)
}

// CountRpsGameInvites implements [GamingStore].
func (s *DbGamingStoreDecorator) CountRpsGameInvites(ctx context.Context, filter *RpsGameInviteFilter) (int64, error) {
	if s.CountRpsGameInvitesFunc != nil {
		return s.CountRpsGameInvitesFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, fmt.Errorf("Gaming store decorator CountRpsGameInvites %w", ErrDelegateNil)
	}
	return s.Delegate.CountRpsGameInvites(ctx, filter)
}

// CreateRpsGameInvite implements [GamingStore].
func (s *DbGamingStoreDecorator) CreateRpsGameInvite(ctx context.Context, invite *models.RpsGameInvite) (*models.RpsGameInvite, error) {
	if s.CreateRpsGameInviteFunc != nil {
		return s.CreateRpsGameInviteFunc(ctx, invite)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator CreateRpsGameInvite %w", ErrDelegateNil)
	}
	return s.Delegate.CreateRpsGameInvite(ctx, invite)
}

// DeleteRpGameInvites implements [GamingStore].
func (s *DbGamingStoreDecorator) DeleteRpGameInvites(ctx context.Context, filter *RpsGameInviteFilter) (int64, error) {
	if s.DeleteRpGameInvitesFunc != nil {
		return s.DeleteRpGameInvitesFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, fmt.Errorf("Gaming store decorator DeleteRpGameInvites %w", ErrDelegateNil)
	}
	return s.Delegate.DeleteRpGameInvites(ctx, filter)
}

// UpdateRpsGameInvite implements [GamingStore].
func (s *DbGamingStoreDecorator) UpdateRpsGameInvite(ctx context.Context, player *models.RpsGameInvite) (*models.RpsGameInvite, error) {
	if s.UpdateRpsGameInviteFunc != nil {
		return s.UpdateRpsGameInviteFunc(ctx, player)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator UpdateRpsGameInvite %w", ErrDelegateNil)
	}
	return s.Delegate.UpdateRpsGameInvite(ctx, player)
}

// CreateRpsParticipants implements [GamingStore].
func (s *DbGamingStoreDecorator) CreateRpsParticipants(ctx context.Context, participant []*models.RpsParticipant) ([]*models.RpsParticipant, error) {
	if s.CreateRpsParticipantsFunc != nil {
		return s.CreateRpsParticipantsFunc(ctx, participant)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator CreateRpsParticipants %w", ErrDelegateNil)
	}
	return s.Delegate.CreateRpsParticipants(ctx, participant)
}

// CountRpsParticipants implements [GamingStore].
func (s *DbGamingStoreDecorator) CountRpsParticipants(ctx context.Context, filter *RpsParticipantFilter) (int64, error) {
	if s.CountRpsParticipantsFunc != nil {
		return s.CountRpsParticipantsFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, fmt.Errorf("Gaming store decorator CountRpsParticipants %w", ErrDelegateNil)
	}
	return s.Delegate.CountRpsParticipants(ctx, filter)
}

// CreateRpsParticipant implements [GamingStore].
func (s *DbGamingStoreDecorator) CreateRpsParticipant(ctx context.Context, participant *models.RpsParticipant) (*models.RpsParticipant, error) {
	if s.CreateRpsParticipantFunc != nil {
		return s.CreateRpsParticipantFunc(ctx, participant)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator CreateRpsParticipant %w", ErrDelegateNil)
	}
	return s.Delegate.CreateRpsParticipant(ctx, participant)
}

// FindRpsParticipant implements [GamingStore].
func (s *DbGamingStoreDecorator) FindRpsParticipant(ctx context.Context, filter *RpsParticipantFilter) (*models.RpsParticipant, error) {
	if s.FindRpsParticipantFunc != nil {
		return s.FindRpsParticipantFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindRpsParticipant %w", ErrDelegateNil)
	}
	return s.Delegate.FindRpsParticipant(ctx, filter)
}

// FindRpsParticipants implements [GamingStore].
func (s *DbGamingStoreDecorator) FindRpsParticipants(ctx context.Context, filter *RpsParticipantFilter) ([]*models.RpsParticipant, error) {
	if s.FindRpsParticipantsFunc != nil {
		return s.FindRpsParticipantsFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindRpsParticipants %w", ErrDelegateNil)
	}
	return s.Delegate.FindRpsParticipants(ctx, filter)
}

// UpdateRpsParticipant implements [GamingStore].
func (s *DbGamingStoreDecorator) UpdateRpsParticipant(ctx context.Context, participant *models.RpsParticipant) (*models.RpsParticipant, error) {
	if s.UpdateRpsParticipantFunc != nil {
		return s.UpdateRpsParticipantFunc(ctx, participant)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator UpdateRpsParticipant %w", ErrDelegateNil)
	}
	return s.Delegate.UpdateRpsParticipant(ctx, participant)
}

// FindRpsGame implements [GamingStore].
func (s *DbGamingStoreDecorator) FindRpsGame(ctx context.Context, filter *RpsGameFilter) (*models.RpsGame, error) {
	if s.FindRpsGameFunc != nil {
		return s.FindRpsGameFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindGame %w", ErrDelegateNil)
	}
	return s.Delegate.FindRpsGame(ctx, filter)
}

// FindRpsGames implements [GamingStore].
func (s *DbGamingStoreDecorator) FindRpsGames(ctx context.Context, filter *RpsGameFilter) ([]*models.RpsGame, error) {
	if s.FindRpsGamesFunc != nil {
		return s.FindRpsGamesFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindGames %w", ErrDelegateNil)
	}
	return s.Delegate.FindRpsGames(ctx, filter)
}

// CreateRpsGame implements [GamingStore].
func (s *DbGamingStoreDecorator) CreateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error) {
	if s.CreateRpsGameFunc != nil {
		return s.CreateRpsGameFunc(ctx, game)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator CreateGame %w", ErrDelegateNil)
	}
	return s.Delegate.CreateRpsGame(ctx, game)
}

// UpdateRpsGame implements [GamingStore].
func (s *DbGamingStoreDecorator) UpdateRpsGame(ctx context.Context, game *models.RpsGame) (*models.RpsGame, error) {
	if s.UpdateRpsGameFunc != nil {
		return s.UpdateRpsGameFunc(ctx, game)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator UpdateGame %w", ErrDelegateNil)
	}
	return s.Delegate.UpdateRpsGame(ctx, game)
}

// CountPlayers implements [GamingStore].
func (s *DbGamingStoreDecorator) CountPlayers(ctx context.Context, filter *PlayersFilter) (int64, error) {
	if s.CountPlayersFunc != nil {
		return s.CountPlayersFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, fmt.Errorf("Gaming store decorator CountPlayers %w", ErrDelegateNil)
	}
	return s.Delegate.CountPlayers(ctx, filter)
}

// CreateFriendship implements [GamingStore].
func (s *DbGamingStoreDecorator) CreateFriendship(ctx context.Context, friendship *models.Friendship) (*models.Friendship, error) {
	if s.CreateFriendshipFunc != nil {
		return s.CreateFriendshipFunc(ctx, friendship)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator CreateFriendship %w", ErrDelegateNil)
	}
	return s.Delegate.CreateFriendship(ctx, friendship)
}

// CreatePlayer implements [GamingStore].
func (s *DbGamingStoreDecorator) CreatePlayer(ctx context.Context, player *models.Player) (*models.Player, error) {
	if s.CreatePlayerFunc != nil {
		return s.CreatePlayerFunc(ctx, player)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator CreatePlayer %w", ErrDelegateNil)
	}
	return s.Delegate.CreatePlayer(ctx, player)
}

// DeleteFriendships implements [GamingStore].
func (s *DbGamingStoreDecorator) DeleteFriendships(ctx context.Context, filter *FriendshipFilter) (int64, error) {
	if s.DeleteFriendshipsFunc != nil {
		return s.DeleteFriendshipsFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, fmt.Errorf("Gaming store decorator DeleteFriendships %w", ErrDelegateNil)
	}
	return s.Delegate.DeleteFriendships(ctx, filter)
}

// DeletePlayers implements [GamingStore].
func (s *DbGamingStoreDecorator) DeletePlayers(ctx context.Context, filter *PlayersFilter) (int64, error) {
	if s.DeletePlayersFunc != nil {
		return s.DeletePlayersFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, fmt.Errorf("Gaming store decorator DeletePlayers %w", ErrDelegateNil)
	}
	return s.Delegate.DeletePlayers(ctx, filter)
}

// NewDbGamingStoreDecorator creates a new DbGamingStoreDecorator.
func NewDbGamingStoreDecorator(db database.Dbx) *DbGamingStoreDecorator {
	return &DbGamingStoreDecorator{
		Delegate: NewDBGamingStore(db),
	}
}

// FindFriendship implements [GamingStore].
func (s *DbGamingStoreDecorator) FindFriendship(ctx context.Context, filter *FriendshipFilter) (*models.Friendship, error) {
	if s.FindFriendshipFunc != nil {
		return s.FindFriendshipFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindFriendship %w", ErrDelegateNil)
	}
	return s.Delegate.FindFriendship(ctx, filter)
}

// FindFriendships implements [GamingStore].
func (s *DbGamingStoreDecorator) FindFriendships(ctx context.Context, filter *FriendshipFilter) ([]*models.Friendship, error) {
	if s.FindFriendshipsFunc != nil {
		return s.FindFriendshipsFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindFriendships %w", ErrDelegateNil)
	}
	return s.Delegate.FindFriendships(ctx, filter)
}

// FindHousePlayer implements [GamingStore].
func (s *DbGamingStoreDecorator) FindHousePlayer(ctx context.Context) (*models.Player, error) {
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindHousePlayer %w", ErrDelegateNil)
	}
	return s.Delegate.FindHousePlayer(ctx)
}

// GetHouseGameAggregates implements [GamingStore].
func (s *DbGamingStoreDecorator) GetHouseGameAggregates(ctx context.Context, housePlayerID uuid.UUID) (*HouseGameAggregates, error) {
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator GetHouseGameAggregates %w", ErrDelegateNil)
	}
	return s.Delegate.GetHouseGameAggregates(ctx, housePlayerID)
}

// GetPlayerGameAggregates implements [GamingStore].
func (s *DbGamingStoreDecorator) GetPlayerGameAggregates(ctx context.Context, playerID uuid.UUID) (*PlayerGameAggregates, error) {
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator GetPlayerGameAggregates %w", ErrDelegateNil)
	}
	return s.Delegate.GetPlayerGameAggregates(ctx, playerID)
}

// FindPlayer implements [GamingStore].
func (s *DbGamingStoreDecorator) FindPlayer(ctx context.Context, filter *PlayersFilter) (*models.Player, error) {
	if s.FindPlayerFunc != nil {
		return s.FindPlayerFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindPlayer %w", ErrDelegateNil)
	}
	return s.Delegate.FindPlayer(ctx, filter)
}

// FindPlayers implements [GamingStore].
func (s *DbGamingStoreDecorator) FindPlayers(ctx context.Context, filter *PlayersFilter) ([]*models.Player, error) {
	if s.FindPlayersFunc != nil {
		return s.FindPlayersFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindPlayers %w", ErrDelegateNil)
	}
	return s.Delegate.FindPlayers(ctx, filter)
}

// UpdateFriendship implements [GamingStore].
func (s *DbGamingStoreDecorator) UpdateFriendship(ctx context.Context, player *models.Friendship) (*models.Friendship, error) {
	if s.UpdateFriendshipFunc != nil {
		return s.UpdateFriendshipFunc(ctx, player)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator UpdateFriendship %w", ErrDelegateNil)
	}
	return s.Delegate.UpdateFriendship(ctx, player)
}

// UpdatePlayer implements [GamingStore].
func (s *DbGamingStoreDecorator) UpdatePlayer(ctx context.Context, player *models.Player) (*models.Player, error) {
	if s.UpdatePlayerFunc != nil {
		return s.UpdatePlayerFunc(ctx, player)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator UpdatePlayer %w", ErrDelegateNil)
	}
	return s.Delegate.UpdatePlayer(ctx, player)
}

// UpdatePlayerLastSeen implements [GamingStore].
func (s *DbGamingStoreDecorator) UpdatePlayerLastSeen(ctx context.Context, playerID uuid.UUID) error {
	if s.UpdatePlayerLastSeenFunc != nil {
		return s.UpdatePlayerLastSeenFunc(ctx, playerID)
	}
	if s.Delegate == nil {
		return fmt.Errorf("Gaming store decorator UpdatePlayerLastSeen %w", ErrDelegateNil)
	}
	return s.Delegate.UpdatePlayerLastSeen(ctx, playerID)
}

// WithTx implements [GamingStore].
func (s *DbGamingStoreDecorator) WithTx(db database.Dbx) *DBGamingStore {
	if s.WithTxFunc != nil {
		return s.WithTxFunc(db)
	}
	if s.Delegate == nil {
		return nil
	}
	return s.Delegate.WithTx(db)
}

func (s *DbGamingStoreDecorator) CountFriendships(ctx context.Context, filter *FriendshipFilter) (int64, error) {
	if s.CountFriendshipsFunc != nil {
		return s.CountFriendshipsFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, fmt.Errorf("Gaming store decorator CountFriendships %w", ErrDelegateNil)
	}
	return s.Delegate.CountFriendships(ctx, filter)
}

// FindRpsGameForUpdate implements [GamingStore].
func (s *DbGamingStoreDecorator) FindRpsGameForUpdate(ctx context.Context, gameID uuid.UUID) (*models.RpsGame, error) {
	if s.FindRpsGameForUpdateFunc != nil {
		return s.FindRpsGameForUpdateFunc(ctx, gameID)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindRpsGameForUpdate %w", ErrDelegateNil)
	}
	return s.Delegate.FindRpsGameForUpdate(ctx, gameID)
}

// FindExpiredPendingBetGames implements [GamingStore].
func (s *DbGamingStoreDecorator) FindExpiredPendingBetGames(ctx context.Context) ([]*models.RpsGame, error) {
	if s.FindExpiredPendingBetGamesFunc != nil {
		return s.FindExpiredPendingBetGamesFunc(ctx)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindExpiredPendingBetGames %w", ErrDelegateNil)
	}
	return s.Delegate.FindExpiredPendingBetGames(ctx)
}

// FindPendingGamesExpiringWithin implements [GamingStore].
func (s *DbGamingStoreDecorator) FindPendingGamesExpiringWithin(ctx context.Context, within time.Duration) ([]*models.RpsGame, error) {
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindPendingGamesExpiringWithin %w", ErrDelegateNil)
	}
	return s.Delegate.FindPendingGamesExpiringWithin(ctx, within)
}

// MarkRpsGameExpirySent implements [GamingStore].
func (s *DbGamingStoreDecorator) MarkRpsGameExpirySent(ctx context.Context, game *models.RpsGame) error {
	if s.Delegate == nil {
		return fmt.Errorf("Gaming store decorator MarkRpsGameExpirySent %w", ErrDelegateNil)
	}
	return s.Delegate.MarkRpsGameExpirySent(ctx, game)
}

func (s *DbGamingStoreDecorator) CreateRpsRematchRequest(ctx context.Context, req *models.RpsRematchRequest) (*models.RpsRematchRequest, error) {
	if s.CreateRpsRematchRequestFunc != nil {
		return s.CreateRpsRematchRequestFunc(ctx, req)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator CreateRpsRematchRequest %w", ErrDelegateNil)
	}
	return s.Delegate.CreateRpsRematchRequest(ctx, req)
}

func (s *DbGamingStoreDecorator) FindRpsRematchRequest(ctx context.Context, filter *RpsRematchFilter) (*models.RpsRematchRequest, error) {
	if s.FindRpsRematchRequestFunc != nil {
		return s.FindRpsRematchRequestFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindRpsRematchRequest %w", ErrDelegateNil)
	}
	return s.Delegate.FindRpsRematchRequest(ctx, filter)
}

func (s *DbGamingStoreDecorator) UpdateRpsRematchRequest(ctx context.Context, req *models.RpsRematchRequest) (*models.RpsRematchRequest, error) {
	if s.UpdateRpsRematchRequestFunc != nil {
		return s.UpdateRpsRematchRequestFunc(ctx, req)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator UpdateRpsRematchRequest %w", ErrDelegateNil)
	}
	return s.Delegate.UpdateRpsRematchRequest(ctx, req)
}

func (s *DbGamingStoreDecorator) FindExpiredPendingRpsRematches(ctx context.Context) ([]*models.RpsRematchRequest, error) {
	if s.FindExpiredPendingRpsRematchesFunc != nil {
		return s.FindExpiredPendingRpsRematchesFunc(ctx)
	}
	if s.Delegate == nil {
		return nil, fmt.Errorf("Gaming store decorator FindExpiredPendingRpsRematches %w", ErrDelegateNil)
	}
	return s.Delegate.FindExpiredPendingRpsRematches(ctx)
}

var _ GamingStore = (*DbGamingStoreDecorator)(nil)
