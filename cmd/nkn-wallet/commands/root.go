package commands

import (
	"errors"
	"math/rand"
	"time"

	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

// Globals
var (
	path             string
	ip               string
	index            int
	ageRecipient     string
	ageRecipientFile string
	ageIdentity      string
)

var rootCmd = &cobra.Command{
	Use:     "nkn-wallet",
	Version: "1.0",
	Short:   "nkn-wallet - A next generation wallet for NKN.",
	Long: `nkn-wallet v1.0
---------------------------------------------------------------------------
nkn-wallet is a library that implements nkn-sdk-go with a new wallet
functionality utilizing age encryption and using the NKN OpenAPI for
querying the NKN blockchain.

Upcoming releases will include the possibility for encryption by
SSH public key ("ssh-ed25519 AAAA...", "ssh-rsa AAAA...") stored on disk
or by fetching the keys from a Github user profile (github.com/[user].keys).

URL: https://github.com/omani/nkn-wallet
MIT license. Copyright (c) 2023 HAH! Sun

[This message will be removed with the next version]
---------------------------------------------------------------------------
`,
}

func RootCmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rand.Seed(time.Now().UnixNano())

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringVarP(&path, "path", "p", "./nkn-wallet.json", "path to wallet file")
	rootCmd.PersistentFlags().StringVar(&ip, "ip", "mainnet-seed-0001.org", "DNS/IP of NKN remote node")
	rootCmd.PersistentFlags().StringVarP(&ageRecipient, "age-recipient", "r", "", "Use recipient for age encryption ['ssh-', 'age1'].")
	rootCmd.PersistentFlags().StringVarP(&ageRecipientFile, "age-recipient-file", "R", "", "Use recipient file for age encryption [ssh public-key, age recipient].")
	rootCmd.PersistentFlags().StringVarP(&ageIdentity, "age-identity", "i", "", "Use identity file for age decryption [ssh private key, age identity file].")

	rootCmd.PersistentFlags().IntVar(&index, "index", 0, "Use account with index.")

	rootCmd.MarkFlagsMutuallyExclusive("age-recipient", "age-recipient-file", "age-identity")
}

func getWallet(store *nknwallet.Store, index int) (*nknwallet.Wallet, error) {
	var wallet *nknwallet.Wallet
	var err error

	if len(ageIdentity) > 0 {
		wallet, err = store.NewWalletByIdentity(ageIdentity, index, nil)
	} else if len(ageRecipientFile) > 0 {
		wallet, err = store.NewWalletByRecipientFile(ageRecipientFile, index, nil)
	} else if len(ageRecipient) > 0 {
		wallet, err = store.NewWalletByRecipient(ageRecipient, index, nil)
	} else {
		wallet, err = store.NewWalletByPassword(index, nil)
	}
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}
	return nil, errors.New("Error: No wallet could be fetched.")
}
