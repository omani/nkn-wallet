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
	restoreCmd.Flags().StringVarP(&alias, "alias", "a", "", "Delete account with given alias.")
}

func runRestore() error {
	if len(seed) == 0 {
		s, err := password.GetPassword("Seed")
		checkerr(err)
		seed = string(s)
	}
	if len(passwd) == 0 {
		pass, err := password.GetConfirmedPassword()
		checkerr(err)
		passwd = string(pass)
	}

	store := nknwallet.NewStore(path)
	seedecoded, err := hex.DecodeString(seed)
	checkerr(err)
	wallet, err := store.RestoreFromSeed(seedecoded, []byte(passwd), alias)
	checkerr(err)

	store.SaveWallet(wallet)

	return nil
}
