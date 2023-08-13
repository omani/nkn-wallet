package commands

import (
	"fmt"

	"github.com/nknorg/nkn/v2/common"
	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer funds to another NKN address",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("index")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
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

	transferCmd.MarkFlagRequired("amount")
	transferCmd.MarkFlagRequired("to")
}

func runTransfer() error {
	_, err := common.ToScriptHash(to)
	checkerr(err)

	store, err := nknwallet.NewStore(path)
	checkerr(err)
	wallet, err := getWallet(store, index)
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
