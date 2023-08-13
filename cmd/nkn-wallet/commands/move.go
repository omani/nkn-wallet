package commands

import (
	"fmt"

	"github.com/nknorg/nkn/v2/common"
	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move funds between accounts in the wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMove()
	},
}

var (
	fromID int
	toID   int
)

func init() {
	rootCmd.AddCommand(moveCmd)

	moveCmd.Flags().IntVar(&fromID, "from-id", 0, "NKN Address of recipient.")
	moveCmd.Flags().IntVar(&toID, "to-id", 0, "NKN Address of recipient.")
	moveCmd.Flags().StringVar(&amount, "amount", "", "Amount of funds to transfer. Use 'all' to transfer all funds.")
	moveCmd.Flags().StringVar(&fee, "fee", "", "Make miners happy by specifying an optional fee for the transaction.")

	moveCmd.MarkFlagRequired("amount")
	moveCmd.MarkFlagRequired("from-id")
	moveCmd.MarkFlagRequired("to-id")
	moveCmd.MarkFlagsRequiredTogether("from-id", "to-id")
}

func runMove() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)
	wallet, err := getWallet(store, fromID)
	checkerr(err)
	recipientwallet, err := store.GetWalletByIndex(toID)
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
	to := recipientwallet.Address()
	txhash, err := wallet.Transfer(to, amount, nil)
	checkerr(err)

	fmt.Printf("Successfully sent %s NKN from %s to %s. txHash: %s\n", a, wallet.Address(), recipientwallet.Address(), txhash)

	return nil
}
