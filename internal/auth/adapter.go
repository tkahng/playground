package auth

import "github.com/tkahng/authgo/internal/shared"

type Adapter interface {
	CreateUser(user *shared.User) (*shared.User, error)
	GetUser(id string) (*shared.User, error)
	GetUserByEmail(email string) (*shared.User, error)
	GetUserByAccount(providerAccountId string, provider string) (*shared.User, error)
	UpdateUser(user *shared.User) (*shared.User, error)
	DeleteUser(id string) error
	LinkAccount(account *shared.UserAccount) error
	UnlinkAccount(providerAccountId string, provider string) error
	// CreateSession(session *Session) (*Session, error)
	// GetSession(sessionToken string) (*Session, error)
	// UpdateSession(session *Session) (*Session, error)
	// DeleteSession(sessionToken string) error
	// CreateVerificationToken(verificationToken *VerificationToken) (*VerificationToken, error)
	// GetVerificationToken(identifier string, token string) (*VerificationToken, error)
	// GetAccount(providerAccountId string, provider string) (*Account, error)
	// GetAuthenticator(credentialID string) (*Authenticator, error)
	// CreateAuthenticator(authenticator *Authenticator) (*Authenticator, error)
	// ListAuthenticatorsByUserId(userId string) ([]*Authenticator, error)
	// UpdateAuthenticatorCounter(credentialID string, newCounter string) (*Authenticator, error)
}

// export interface Adapter {
//   /**
//    * Creates a user in the database and returns it.
//    *
//    * See also [User management](https://authjs.dev/guides/creating-a-database-adapter#user-management)
//    */
//   createUser?(user: AdapterUser): Awaitable<AdapterUser>
//   /**
//    * Returns a user from the database via the user id.
//    *
//    * See also [User management](https://authjs.dev/guides/creating-a-database-adapter#user-management)
//    */
//   getUser?(id: string): Awaitable<AdapterUser | null>
//   /**
//    * Returns a user from the database via the user's email address.
//    *
//    * See also [Verification tokens](https://authjs.dev/guides/creating-a-database-adapter#verification-tokens)
//    */
//   getUserByEmail?(email: string): Awaitable<AdapterUser | null>
//   /**
//    * Using the provider id and the id of the user for a specific account, get the user.
//    *
//    * See also [User management](https://authjs.dev/guides/creating-a-database-adapter#user-management)
//    */
//   getUserByAccount?(
//     providerAccountId: Pick<AdapterAccount, "provider" | "providerAccountId">
//   ): Awaitable<AdapterUser | null>
//   /**
//    * Updates a user in the database and returns it.
//    *
//    * See also [User management](https://authjs.dev/guides/creating-a-database-adapter#user-management)
//    */
//   updateUser?(
//     user: Partial<AdapterUser> & Pick<AdapterUser, "id">
//   ): Awaitable<AdapterUser>
//   /**
//    * @todo This method is currently not invoked yet.
//    *
//    * See also [User management](https://authjs.dev/guides/creating-a-database-adapter#user-management)
//    */
//   deleteUser?(
//     userId: string
//   ): Promise<void> | Awaitable<AdapterUser | null | undefined>
//   /**
//    * This method is invoked internally (but optionally can be used for manual linking).
//    * It creates an [Account](https://authjs.dev/reference/core/adapters#models) in the database.
//    *
//    * See also [User management](https://authjs.dev/guides/creating-a-database-adapter#user-management)
//    */
//   linkAccount?(
//     account: AdapterAccount
//   ): Promise<void> | Awaitable<AdapterAccount | null | undefined>
//   /** @todo This method is currently not invoked yet. */
//   unlinkAccount?(
//     providerAccountId: Pick<AdapterAccount, "provider" | "providerAccountId">
//   ): Promise<void> | Awaitable<AdapterAccount | undefined>
//   /**
//    * Creates a session for the user and returns it.
//    *
//    * See also [Database Session management](https://authjs.dev/guides/creating-a-database-adapter#database-session-management)
//    */
//   createSession?(session: {
//     sessionToken: string
//     userId: string
//     expires: Date
//   }): Awaitable<AdapterSession>
//   /**
//    * Returns a session and a userfrom the database in one go.
//    *
//    * :::tip
//    * If the database supports joins, it's recommended to reduce the number of database queries.
//    * :::
//    *
//    * See also [Database Session management](https://authjs.dev/guides/creating-a-database-adapter#database-session-management)
//    */
//   getSessionAndUser?(
//     sessionToken: string
//   ): Awaitable<{ session: AdapterSession; user: AdapterUser } | null>
//   /**
//    * Updates a session in the database and returns it.
//    *
//    * See also [Database Session management](https://authjs.dev/guides/creating-a-database-adapter#database-session-management)
//    */
//   updateSession?(
//     session: Partial<AdapterSession> & Pick<AdapterSession, "sessionToken">
//   ): Awaitable<AdapterSession | null | undefined>
//   /**
//    * Deletes a session from the database. It is preferred that this method also
//    * returns the session that is being deleted for logging purposes.
//    *
//    * See also [Database Session management](https://authjs.dev/guides/creating-a-database-adapter#database-session-management)
//    */
//   deleteSession?(
//     sessionToken: string
//   ): Promise<void> | Awaitable<AdapterSession | null | undefined>
//   /**
//    * Creates a verification token and returns it.
//    *
//    * See also [Verification tokens](https://authjs.dev/guides/creating-a-database-adapter#verification-tokens)
//    */
//   createVerificationToken?(
//     verificationToken: VerificationToken
//   ): Awaitable<VerificationToken | null | undefined>
//   /**
//    * Return verification token from the database and deletes it
//    * so it can only be used once.
//    *
//    * See also [Verification tokens](https://authjs.dev/guides/creating-a-database-adapter#verification-tokens)
//    */
//   useVerificationToken?(params: {
//     identifier: string
//     token: string
//   }): Awaitable<VerificationToken | null>
//   /**
//    * Get account by provider account id and provider.
//    *
//    * If an account is not found, the adapter must return `null`.
//    */
//   getAccount?(
//     providerAccountId: AdapterAccount["providerAccountId"],
//     provider: AdapterAccount["provider"]
//   ): Awaitable<AdapterAccount | null>
//   /**
//    * Returns an authenticator from its credentialID.
//    *
//    * If an authenticator is not found, the adapter must return `null`.
//    */
//   getAuthenticator?(
//     credentialID: AdapterAuthenticator["credentialID"]
//   ): Awaitable<AdapterAuthenticator | null>
//   /**
//    * Create a new authenticator.
//    *
//    * If the creation fails, the adapter must throw an error.
//    */
//   createAuthenticator?(
//     authenticator: AdapterAuthenticator
//   ): Awaitable<AdapterAuthenticator>
//   /**
//    * Returns all authenticators from a user.
//    *
//    * If a user is not found, the adapter should still return an empty array.
//    * If the retrieval fails for some other reason, the adapter must throw an error.
//    */
//   listAuthenticatorsByUserId?(
//     userId: AdapterAuthenticator["userId"]
//   ): Awaitable<AdapterAuthenticator[]>
//   /**
//    * Updates an authenticator's counter.
//    *
//    * If the update fails, the adapter must throw an error.
//    */
//   updateAuthenticatorCounter?(
//     credentialID: AdapterAuthenticator["credentialID"],
//     newCounter: AdapterAuthenticator["counter"]
//   ): Awaitable<AdapterAuthenticator>
// }
