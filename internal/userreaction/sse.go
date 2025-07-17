package userreaction

type LatestUserReactionSseEvent struct {
	UserReaction *UserReaction `json:"user_reaction"`
}

type LatestUserReactionStatsSseEvent struct {
	UserReactionStats *UserReactionStats `json:"user_reaction_stats"`
}
