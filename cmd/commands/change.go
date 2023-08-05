package commands

import (
	"fmt"

	"github.com/nknorg/nkn/v2/util/password"
	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var changeCmd = &cobra.Command{
	Use:   "change",
	Short: "Change various information of an account in the wallet",
}
var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Change the password of an account in the wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(alias) == 0 && index == 0 {
			cobra.CheckErr("Need either index or alias flag.")
		}
		return runChangePassword()
	},
}
var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Change the alias of an account in the wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(alias) == 0 && index == 0 {
			cobra.CheckErr("Need either index or alias flag.")
		}
		return runChangeAlias()
	},
}

var ()

func init() {
	rootCmd.AddCommand(changeCmd)
	changeCmd.AddCommand(passwordCmd)
	changeCmd.AddCommand(aliasCmd)

	changeCmd.PersistentFlags().StringVarP(&alias, "alias", "a", "", "Alias of account.")
	changeCmd.PersistentFlags().IntVarP(&index, "index", "i", 0, "Index of account.")
}

func runChangePassword() error {
	if len(passwd) == 0 {
		pass, err := password.GetPassword("Current password")
		checkerr(err)
		passwd = string(pass)
	}

	store := nknwallet.NewStore(path)

	var wallet *nknwallet.Wallet
	var err error
	if len(alias) > 0 {
		wallet, err = store.GetWalletByAlias(alias, []byte(passwd))
	}
	if index > 0 {
		wallet, err = store.GetWalletByIndex(index, []byte(passwd))
	}
	checkerr(err)

	newpasswd, err := password.GetConfirmedPassword()
	checkerr(err)

	alias = wallet.Alias
	store.DeleteWalletByIndex(index)

	w, err := store.RestoreFromSeed(wallet.Seed(), newpasswd, alias)
	checkerr(err)

	err = store.SaveWallet(w)
	checkerr(err)

	fmt.Println("Successfully changed password of account.")

	return nil
}

func runChangeAlias() error {
	if len(passwd) == 0 {
		pass, err := password.GetPassword("Current password")
		checkerr(err)
		passwd = string(pass)
	}

	store := nknwallet.NewStore(path)

	var wallet *nknwallet.Wallet
	var err error
	if len(alias) > 0 {
		wallet, err = store.GetWalletByAlias(alias, []byte(passwd))
	}
	if index > 0 {
		wallet, err = store.GetWalletByIndex(index, []byte(passwd))
	}
	checkerr(err)

	store.DeleteWalletByIndex(index)

	w, err := store.RestoreFromSeed(wallet.Seed(), []byte(passwd), alias)
	checkerr(err)

	err = store.SaveWallet(w)
	checkerr(err)

	fmt.Println("Successfully changed alias of account.")

	return nil
}
