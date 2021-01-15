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
	"github.com/spf13/cobra"
	"log"
	"strconv"
)

var buyLicenseMonth int
var payLicenseType int
var receiverBeatlesAddr string

// priceCmd represents the price command
var ethPriceCmd = &cobra.Command{
	Use:   "price",
	Short: "show current eth price",
	Long:  `show current eth price`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := cmdcommon.IsProcessStarted(); err != nil {
			log.Println(err)
			return
		}

		var param []string

		param = append(param, strconv.Itoa(buyLicenseMonth))
		param = append(param, strconv.Itoa(payLicenseType))
		param = append(param, receiverBeatlesAddr)

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_SHOW_ETH_PRICE, param)
	},
}

func init() {
	ethCmd.AddCommand(ethPriceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// priceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// priceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	ethPriceCmd.Flags().IntVarP(&buyLicenseMonth, "month", "m", 6, "month to buy")
	ethPriceCmd.Flags().IntVarP(&payLicenseType, "typ", "t", 0, "pay type [0:eth,default,1:btlc]")
	ethPriceCmd.Flags().StringVarP(&receiverBeatlesAddr, "receiver", "r", "", "license receiver, default is current user")
}
