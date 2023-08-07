package commands

import (
	"fmt"

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
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("index")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangePassword()
	},
}
var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Change the alias of an account in the wallet",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("index")
		cmd.MarkFlagRequired("newalias")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangeAlias()
	},
}

var (
	newalias string
)

func init() {
	rootCmd.AddCommand(changeCmd)

	changeCmd.AddCommand(passwordCmd)
	changeCmd.AddCommand(aliasCmd)

	changeCmd.PersistentFlags().StringVar(&newalias, "newalias", "", "New alias of account.")
}

func runChangePassword() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)
	wallet, err := getWallet(store, index)
	checkerr(err)

	err = store.SetPassword(wallet)
	checkerr(err)

	return nil
}

func runChangeAlias() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)

	if len(alias) > 0 {
		if ok := store.IsExistWalletByAlias(alias); !ok {
			cobra.CheckErr(fmt.Sprintf("Wallet with alias %s does not exist.", alias))
		}
	}
	if ok := store.IsExistWalletByAlias(newalias); ok {
		cobra.CheckErr(fmt.Sprintf("Account with alias %s already exists.", alias))
	}
	wallet, err := getWallet(store, index)
	checkerr(err)
	store.SetAlias(wallet, newalias)
	checkerr(err)

	return nil
}
