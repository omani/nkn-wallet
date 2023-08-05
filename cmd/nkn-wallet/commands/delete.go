package commands

import (
	nknwallet "github.com/omani/nkn-wallet"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an account from the wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDelete()
	},
}

var ()

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&alias, "alias", "a", "", "Delete account with given alias.")
	deleteCmd.Flags().IntVarP(&index, "index", "i", 0, "Delete account with given index.")

	deleteCmd.MarkFlagsMutuallyExclusive("index", "alias")
}

func runDelete() error {
	if len(alias) == 0 && index == 0 {
		cobra.CheckErr("Need either index or alias flag.")
	}

	store := nknwallet.NewStore(path)

	if len(alias) > 0 {
		store.DeleteWalletByAlias(alias)
	} else if index > 0 {
		store.DeleteWalletByIndex(index)
	}

	return nil
}
