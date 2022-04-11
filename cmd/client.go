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
	"github.com/stormentt/packetloss/client"
	"golang.org/x/crypto/blake2b"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Send UDP packets to server and record acknowledgements",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		remoteStr := viper.GetString("remote")
		raddr, err := net.ResolveUDPAddr("udp", remoteStr)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":         err,
				"RemoteAddress": remoteStr,
			}).Fatal("could not resolve remote addr")
		}

		log.WithFields(log.Fields{
			"RemoteAddress": remoteStr,
		}).Info("sending packets")

		hkey := blake2b.Sum512([]byte(viper.GetString("key")))
		log.WithFields(log.Fields{
			"hkey": fmt.Sprintf("%X", hkey),
		}).Debug("using mac key")

		err = client.Start(hkey[:], raddr)

		if err != nil {
			log.WithFields(log.Fields{
				"Error":         err,
				"RemoteAddress": remoteStr,
			}).Fatal("could not send packets")
		}

		log.Info("finished")
	},
}

func init() {
	clientCmd.Flags().StringP("remote", "r", "localhost:6666", "Remote address to send packets to")
	clientCmd.Flags().StringP("key", "k", "", "Key to use for HMAC")
	clientCmd.Flags().DurationP("packet-time", "t", 100*time.Millisecond, "Time to wait between sending packets")

	viper.BindPFlag("remote", clientCmd.Flags().Lookup("remote"))
	viper.BindPFlag("key", clientCmd.Flags().Lookup("key"))
	viper.BindPFlag("packet-time", clientCmd.Flags().Lookup("packet-time"))

	rootCmd.AddCommand(clientCmd)
}
