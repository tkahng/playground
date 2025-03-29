package main

import (
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/tkahng/authgo/internal/tools/security"
)

func main() {

	otp := security.GenerateOtp(6)
	fmt.Println(otp)

	tokenHash := security.GenerateTokenHash("tkahng@gmail.com", otp)
	fmt.Println(tokenHash)
	tokenHash2, _ := security.CreateHash("tkahng@gmail.com"+otp, argon2id.DefaultParams)
	tokenHash3, _ := security.CreateHash("tkahng@gmail.com"+otp, argon2id.DefaultParams)
	// tokenHash2 := security.GenerateTokenHash2("tkahng@gmail.com", otp)
	fmt.Println(tokenHash2)

	// encryptionKey := security.RandomString(32)
	fmt.Println(tokenHash3)

	fmt.Println(security.ComparePasswordAndHash("tkahng@gmail.com"+otp, tokenHash3))
	fmt.Println(security.ComparePasswordAndHash("tkahng@gmail.com"+otp, tokenHash2))
	// fmt.Println(security.ComparePasswordAndHash("tkahng@gmail.com"+otp, tokenHash3))
	// opts := core.DefaultAuthSettings()
	// optsStr, err := json.Marshal(opts)
	// if err != nil {
	// 	panic(err)
	// }
	// encryptedOpts, err := security.Encrypt(optsStr, encryptionKey)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(encryptedOpts)
	// decryptedOpts, err := security.Decrypt(encryptedOpts, encryptionKey)
	// if err != nil {
	// 	panic(err)
	// }
	// println(encryptedOpts)
	// println(string(decryptedOpts))
}
