/*
Copyright © 2022 Tanner Storment

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Send UDP packets to server and record acknowledgements",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client called")
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.Flags().StringP("remote", "r", "localhost:6666", "Remote address to send packets to")
	clientCmd.Flags().StringP("key", "k", "", "Key to use for HMAC")

	viper.BindPFlag("remote", clientCmd.PersistentFlags().Lookup("remote"))
	viper.BindPFlag("key", clientCmd.PersistentFlags().Lookup("key"))
}
