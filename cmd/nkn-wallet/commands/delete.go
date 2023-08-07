package commands

import (
	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an account from the wallet",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.MarkFlagRequired("index")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDelete()
	},
}

var ()

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete() error {
	store, err := nknwallet.NewStore(path)
	checkerr(err)
	store.DeleteWalletByIndex(index)

	return nil
}
