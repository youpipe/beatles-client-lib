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
)

// buyCmd represents the buy command
var (
	licenseusername  string
	licenseuseremail string
	licenseusercell  string
)

var buyCmd = &cobra.Command{
	Use:   "buy",
	Short: "buy from eth with price",
	Long:  `buy from eth with price`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := cmdcommon.IsProcessStarted(); err != nil {
			log.Println(err)
			return
		}

		if licenseuseremail == "" {
			log.Println("please input email")
			return
		}

		var param []string

		param = append(param, licenseusername, licenseuseremail, licenseusercell)

		cmdclient.StringOpCmdSend("", cmdcommon.CMD_ETH_BUY, param)
	},
}

func init() {
	ethCmd.AddCommand(buyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	buyCmd.Flags().StringVarP(&licenseusername, "name", "n", "", "user name")
	buyCmd.Flags().StringVarP(&licenseuseremail, "email", "e", "", "user email address")
	buyCmd.Flags().StringVarP(&licenseusercell, "cell", "c", "", "user cell phone")

}
