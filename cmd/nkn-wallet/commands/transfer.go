package commands

import (
	"fmt"

	"github.com/nknorg/nkn/v2/common"
	"github.com/nknorg/nkn/v2/util/password"
	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer funds to another NKN address",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(alias) == 0 && index == 0 {
			cobra.CheckErr("Need either index or alias flag.")
		}
		return runTransfer()
	},
}

var (
	to     string
	amount string
	fee    string
)

func init() {
	rootCmd.AddCommand(transferCmd)

	transferCmd.Flags().StringVar(&to, "to", "", "NKN Address of recipient.")
	transferCmd.Flags().StringVar(&amount, "amount", "", "Amount of funds to transfer.")
	transferCmd.Flags().StringVar(&fee, "fee", "", "Make miners happy by specifying an optional fee for the transaction.")
	transferCmd.PersistentFlags().StringVarP(&alias, "alias", "a", "", "Alias of account.")
	transferCmd.PersistentFlags().IntVarP(&index, "index", "i", 0, "Index of account.")

	transferCmd.MarkFlagRequired("amount")
	transferCmd.MarkFlagRequired("to")
}

func runTransfer() error {
	_, err := common.ToScriptHash(to)
	checkerr(err)

	if len(passwd) == 0 {
		pass, err := password.GetPassword("")
		checkerr(err)
		passwd = string(pass)
	}

	store := nknwallet.NewStore(path)

	var wallet *nknwallet.Wallet
	if len(alias) > 0 {
		wallet, err = store.GetWalletByAlias(alias, []byte(passwd))
	}
	if index > 0 {
		wallet, err = store.GetWalletByIndex(index, []byte(passwd))
	}
	checkerr(err)

	if amount == "all" {
		a, err := wallet.Balance()
		checkerr(err)
		amount = a.String()
	}
	a, err := common.StringToFixed64(amount)
	checkerr(err)
	if a == 0 || amount == "0" {
		cobra.CheckErr("Trying to send amount of 0. Aborting!")
	}

	txhash, err := wallet.Transfer(to, a.String(), nil)
	checkerr(err)

	fmt.Printf("Successfully sent %s NKN from %s to %s. txHash: %s\n", a, wallet.Address(), to, txhash)

	return nil
}
