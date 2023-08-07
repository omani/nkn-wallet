package commands

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show various information to an account in the wallet",
}
var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show balance of an account in the wallet",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("index")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runShowBalance()
	},
}
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show account information",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("index")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runShowInfo()
	},
}
var txnCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Show the last 250 transactions of an account (from those 250 only of type TRANSFER_ASSET_TYPE)",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("index")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runShowTxn()
	},
}

var ()

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.AddCommand(balanceCmd)
	showCmd.AddCommand(infoCmd)
	showCmd.AddCommand(txnCmd)
}

func runShowBalance() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)
	wallet, err := getWallet(store, index)
	checkerr(err)
	balance, err := wallet.OpenAPI().GetBalance()
	checkerr(err)
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.Style().Options.DrawBorder = false
	t.SetOutputMirror(os.Stdout)
	t.SetAlign([]text.Align{text.AlignCenter, text.AlignCenter})
	t.AppendHeader(table.Row{"id", "alias", "address", "balance"})
	t.AppendRow(table.Row{wallet.ID, wallet.Alias, wallet.Address(), balance})
	t.Render()

	return nil
}

func runShowInfo() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)
	wallet, err := getWallet(store, index)
	checkerr(err)
	fmt.Println(wallet.ID)
	checkerr(err)
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.Style().Options.DrawBorder = false
	t.SetOutputMirror(os.Stdout)
	t.SetAlign([]text.Align{text.AlignCenter, text.AlignCenter})
	t.AppendHeader(table.Row{"id", "alias", "address", "pubkey", "seed"})
	pubkey := hex.EncodeToString(wallet.PubKey())
	t.AppendRow(table.Row{wallet.ID, wallet.Alias, wallet.Address(), pubkey, wallet.ShowSeed()})
	t.Render()

	return nil
}

func runShowTxn() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)
	wallet, err := getWallet(store, index)
	checkerr(err)
	txn, err := wallet.OpenAPI().GetTransactions()
	checkerr(err)

	if txn == nil {
		fmt.Println("Account has no transactions.")
		return nil
	}
	if len(txn.Data) == 0 {
		fmt.Println("Account has no transactions.")
		return nil
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.Style().Options.DrawBorder = false
	t.SetOutputMirror(os.Stdout)
	t.SetAlign([]text.Align{text.AlignCenter, text.AlignCenter})
	t.AppendHeader(table.Row{"created at", "block height", "txn hash", "sender", "recipient", "amount"})

	for _, tx := range txn.Data {
		if tx.TxType != "TRANSFER_ASSET_TYPE" {
			continue
		}
		t.AppendRow(table.Row{tx.CreatedAt, tx.BlockHeight, tx.Hash, tx.Payload.SenderWallet, tx.Payload.RecipientWallet, tx.Payload.Amount})
	}
	t.Render()

	balance, err := wallet.OpenAPI().GetBalance()
	checkerr(err)
	fmt.Printf("Total balance of account: %s NKN\n", balance)

	return nil
}
