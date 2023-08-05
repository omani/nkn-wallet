package commands

import (
	"math/rand"
	"time"

	"github.com/spf13/cobra"
)

// Globals
var (
	path  string
	ip    string
	index int
)

var rootCmd = &cobra.Command{
	Use:     "nkn-wallet",
	Version: "1.0",
	Short:   "nkn-wallet - A next generation wallet for NKN.",
	Long: `nkn-wallet v1.0
---------------------------------------------------------------------------
nkn-wallet is a library that implements nkn-sdk-go with a new wallet
functionality utilizing age encryption and using the NKN OpenAPI for
querying the NKN blockchain.

The following releases will include the possibility for encryption by
SSH public key ("ssh-ed25519 AAAA...", "ssh-rsa AAAA...") stored on disk
or by fetching the keys from Github (github.com/[user].keys).

URL: https://github.com/omani/nkn-wallet
MIT license. Copyright (c) 2023 HAH! Sun

[This message will be removed with the next version]
---------------------------------------------------------------------------
`,
}

func RootCmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rand.Seed(time.Now().UnixNano())

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringVarP(&path, "path", "p", "./nkn-wallet.json", "path to wallet file")
	rootCmd.PersistentFlags().StringVar(&ip, "ip", "mainnet-seed-0001.org", "DNS/IP of NKN remote node")
	rootCmd.PersistentFlags().StringVar(&passwd, "password", "", "Password for the new account.")

	rootCmd.PersistentFlags().MarkHidden("password")
}
