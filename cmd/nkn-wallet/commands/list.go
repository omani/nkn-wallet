package commands

import (
	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accounts of the wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)
	checkerr(store.ListWallets())

	return nil
}
