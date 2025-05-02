package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/tools/security"
)

// struct to save in the value of the param in the db
type EncryptedAuthOptions struct {
	EncryptedAuthOptions string `form:"encrypted_auth_options" json:"encrypted_auth_options"`
}

func GetOrSetEncryptedAuthOptions(ctx context.Context, dbx bob.Executor, encryptionKey string) (*AuthOptions, error) {
	var opts *AuthOptions
	var encryptedOpts *EncryptedAuthOptions
	// get the encrypted auth options from the db
	encryptedParams, err := queries.FindParams[EncryptedAuthOptions](ctx, dbx, EncryptedAuthOptionsKey)
	if err != nil {
		return nil, fmt.Errorf("error getting encrypted auth options from db: %w", err)
	}
	if encryptedParams != nil {
		encryptedOpts = &encryptedParams.Value.Val
	}
	// if the encrypted auth options are not nil, decrypt them
	if encryptedOpts != nil {
		decryptedOptString, err := security.Decrypt(encryptedOpts.EncryptedAuthOptions, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting auth options: %w", err)
		}
		var authOpts *AuthOptions
		err = json.Unmarshal(decryptedOptString, &authOpts)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling auth options: %w", err)
		}
		opts = authOpts
	}
	if opts == nil {
		opts = DefaultAuthSettings()
		optsStr, err := json.Marshal(opts)
		if err != nil {
			return nil, fmt.Errorf("error marshalling auth options: %w", err)
		}
		encryptedOptsStr, err := security.Encrypt(optsStr, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("error encrypting auth options: %w", err)
		}

		encryptedOpts = &EncryptedAuthOptions{
			EncryptedAuthOptions: encryptedOptsStr,
		}
		err = queries.SetParams(ctx, dbx, EncryptedAuthOptionsKey, encryptedOpts)
		if err != nil {
			return nil, err
		}

	}
	return opts, nil
}
