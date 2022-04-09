/*
Copyright © 2022 Tanner Storment

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Listen for UDP packets and Acknowledge them",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("server called")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("local", "l", ":6666", "Local address to listen on")
	serverCmd.Flags().StringP("key", "k", "", "Key to use for HMAC")

	viper.BindPFlag("local", serverCmd.PersistentFlags().Lookup("local"))
	viper.BindPFlag("key", serverCmd.PersistentFlags().Lookup("key"))
}
