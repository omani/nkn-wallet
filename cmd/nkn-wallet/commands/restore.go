package commands

import (
	"encoding/hex"

	"github.com/nknorg/nkn/v2/util/password"
	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore an account from a seed",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRestore()
	},
}

var (
	seed string
)

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVar(&seed, "seed", "", "Seed of the account to be restored.")
}

func runRestore() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)

	if len(seed) == 0 {
		s, err := password.GetPassword("Seed")
		checkerr(err)
		seed = string(s)

	}
	seedbyte, err := hex.DecodeString(seed)
	checkerr(err)

	var wallet *nknwallet.Wallet

	if len(ageIdentity) > 0 {
		wallet, err = store.RestoreFromSeedByIdentity(seedbyte, ageIdentity)
	} else if len(ageRecipientFile) > 0 {
		wallet, err = store.RestoreFromSeedByIdentity(seedbyte, ageIdentity)
	} else if len(ageRecipient) > 0 {
		wallet, err = store.RestoreFromSeedByIdentity(seedbyte, ageIdentity)
	} else {
		wallet, err = store.RestoreFromSeedByPassword(seedbyte)
	}
	checkerr(err)
	store.SaveWallet(wallet)

	return nil
}
