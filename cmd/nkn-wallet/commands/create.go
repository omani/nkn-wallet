package commands

import (
	"encoding/hex"
	"fmt"

	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an account in the wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCreate()
	},
}

var (
	passwd string
	alias  string
	save   bool
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().BoolVarP(&save, "save", "s", false, "Save new account to wallet.")
}

func checkerr(err error) {
	if err != nil {
		cobra.CheckErr(err)
	}
}

func runCreate() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)
	wallet, err := getWallet(store, index)
	checkerr(err)

	fmt.Println("Account information:")
	fmt.Printf("ID: %d\n", wallet.ID)
	fmt.Printf("Address: %s\n", wallet.Address())
	fmt.Printf("Seed: %s\n", hex.EncodeToString(wallet.Seed()))
	if len(alias) > 0 {
		fmt.Printf("Alias: %s\n", alias)
	}

	if save {
		err = store.SaveWallet(wallet)
		checkerr(err)
		fmt.Println("Account saved successfully.")
	}

	return nil
}
