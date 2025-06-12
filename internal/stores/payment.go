package stores

type DbPaymentStore struct {
	*DbStripeStore
	*DbRbacStore
	*DbTeamStore
}
