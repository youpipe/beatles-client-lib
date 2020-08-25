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
	"fmt"

	"github.com/spf13/cobra"
)

var transactionUsedForLicense bool

// transactionCmd represents the transaction command
var ethtransactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "show all transaction",
	Long:  `show all transaction`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("transaction called")
	},
}

func init() {
	ethCmd.AddCommand(ethtransactionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transactionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transactionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	ethtransactionCmd.Flags().BoolVarP(&transactionUsedForLicense, "not", "n", false, "transaction not used for license")

}