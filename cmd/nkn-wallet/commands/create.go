package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/nknorg/nkn/v2/util/password"
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

	createCmd.Flags().StringVar(&alias, "alias", "", "Alias for the new account.")
	createCmd.Flags().BoolVarP(&save, "save", "s", false, "Save new account to wallet.")
}

func checkerr(err error) {
	if err != nil {
		cobra.CheckErr(err)
	}
}

func runCreate() error {
	store := nknwallet.NewStore(path)
	if ok := store.IsExistWalletByAlias(alias); ok {
		cobra.CheckErr(fmt.Sprintf("Account with alias %s already exists.", alias))
	}
	if len(passwd) == 0 {
		pass, err := password.GetConfirmedPassword()
		if err != nil {
			return err
		}
		passwd = string(pass)
	}

	wallet, err := store.NewWallet([]byte(passwd), alias, nil)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Account information:")
	fmt.Printf("Seed: %s\n", hex.EncodeToString(wallet.Seed()))
	fmt.Printf("Address: %s\n", wallet.Address())
	if len(alias) > 0 {
		fmt.Printf("Alias: %s\n", alias)
	}

	if save {
		err = store.SaveWallet(wallet)
		checkerr(err)
	}
	fmt.Println("Account saved successfully.")

	return nil
}
