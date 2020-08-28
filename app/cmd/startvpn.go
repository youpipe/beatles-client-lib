/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/giantliao/beatles-client-lib/app/cmdclient"
	"github.com/giantliao/beatles-client-lib/app/cmdcommon"
	"log"
	"strconv"

	"github.com/spf13/cobra"
)

var startVpnIndex int

// vpnCmd represents the vpn command
var startvpnCmd = &cobra.Command{
	Use:   "vpn",
	Short: "start vpn",
	Long:  `start vpn`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := cmdcommon.IsProcessStarted(); err != nil {
			log.Println(err)
			return
		}

		var param []string

		param = append(param, strconv.Itoa(startVpnIndex))

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_START_VPN, param)
	},
}

func init() {
	startCmd.AddCommand(startvpnCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// vpnCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// vpnCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	startvpnCmd.Flags().IntVarP(&startVpnIndex, "miner-index", "m", 0, "choose a miner index to start vpn")

}
