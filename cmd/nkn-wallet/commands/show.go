package commands

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"github.com/nknorg/nkn/v2/util/password"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(alias) == 0 && index == 0 {
			cobra.CheckErr("Need either index or alias flag.")
		}
		return runShowBalance()
	},
}
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show account information",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(alias) == 0 && index == 0 {
			cobra.CheckErr("Need either index or alias flag.")
		}
		return runShowInfo()
	},
}
var txnCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Show the last 250 transactions of an account (from those 250 only of type TRANSFER_ASSET_TYPE)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(alias) == 0 && index == 0 {
			cobra.CheckErr("Need either index or alias flag.")
		}
		return runShowTxn()
	},
}

var ()

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.AddCommand(balanceCmd)
	showCmd.AddCommand(infoCmd)
	showCmd.AddCommand(txnCmd)

	showCmd.PersistentFlags().StringVarP(&alias, "alias", "a", "", "Show balance of account with given alias.")
	showCmd.PersistentFlags().IntVarP(&index, "index", "i", 0, "Show balance of account with given index.")
}

func runShowBalance() error {
	if len(passwd) == 0 {
		pass, err := password.GetPassword("")
		if err != nil {
			return err
		}
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
	if len(passwd) == 0 {
		pass, err := password.GetPassword("")
		if err != nil {
			return err
		}
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
	if len(passwd) == 0 {
		pass, err := password.GetPassword("")
		if err != nil {
			return err
		}
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
