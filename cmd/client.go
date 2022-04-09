/*
Copyright © 2022 Tanner Storment

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
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
