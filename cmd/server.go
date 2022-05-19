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
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stormentt/packetloss/server"
	"golang.org/x/crypto/blake2b"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Listen for UDP packets and Acknowledge them",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		localStr := viper.GetString("local")
		laddr, err := net.ResolveUDPAddr("udp", localStr)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":        err,
				"LocalAddress": localStr,
			}).Fatal("could not resolve listen addr")
		}

		hkey := blake2b.Sum512([]byte(viper.GetString("key")))

		log.WithFields(log.Fields{
			"hkey": fmt.Sprintf("%X", hkey),
		}).Debug("using mac key")

		err = server.Listen(hkey[:], laddr)

		if err != nil {
			log.WithFields(log.Fields{
				"Error":        err,
				"LocalAddress": localStr,
			}).Fatal("could not listen")
		}

		log.Info("finished")
	},
}

func init() {
	serverCmd.Flags().StringP("local", "l", ":6666", "Local address to listen on")
	serverCmd.Flags().StringP("key", "k", "", "Key to use for HMAC")
	serverCmd.Flags().Duration("cull-time", time.Minute*10, "time between culling server stats")

	viper.BindPFlag("local", serverCmd.Flags().Lookup("local"))
	viper.BindPFlag("key", serverCmd.Flags().Lookup("key"))
	viper.BindPFlag("cull_time", serverCmd.Flags().Lookup("cull-time"))

	rootCmd.AddCommand(serverCmd)
}
