// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"
	"github.com/giantliao/beatles-client-lib/bootstrap"
	"github.com/howeyc/gopass"
	"os"

	"github.com/giantliao/beatles-client-lib/app/cmdcommon"
	"github.com/giantliao/beatles-client-lib/config"

	"github.com/giantliao/beatles-client-lib/app/cmdservice"

	"github.com/spf13/cobra"
	"log"
)

////var cfgFile string
//
var (
	cmdconfigfilename string
)

var keypassword string

func inputpassword() (password string, err error) {
	passwd, err := gopass.GetPasswdPrompt("Please Enter Password: ", true, os.Stdin, os.Stdout)
	if err != nil {
		return "", err
	}

	if len(passwd) < 1 {
		return "", errors.New("Please input valid password")
	}

	return string(passwd), nil
}

func inputChoose() (choose string, err error) {
	c, err := gopass.GetPasswdPrompt("Do you reinit config[yes/no]: ", true, os.Stdin, os.Stdout)
	if err != nil {
		return "", err
	}

	return string(c), nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cc",
	Short: "start beatles client in current shell",
	Long:  `start beatles client in current shell`,
	Run: func(cmd *cobra.Command, args []string) {

		_, err := cmdcommon.IsProcessCanStarted()
		if err != nil {
			log.Println(err)
			return
		}

		InitCfg()
		cfg := config.GetCBtlc()
		cfg.Save()

		//if keypassword == "" {
		//	if keypassword, err = inputpassword(); err != nil {
		//		log.Println(err)
		//		return
		//	}
		//}

		if len(cfg.Miners) == 0 {
			err := bootstrap.UpdateBootstrap()
			if err != nil {
				log.Println(err)
				return
			}
		}

		cmdservice.GetCmdServerInst().StartCmdService()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func InitCfg() {
	if cmdconfigfilename != "" {
		cfg := config.LoadFromCfgFile(cmdconfigfilename)
		if cfg == nil {
			return
		}
	} else {
		config.LoadFromCmd(cfginit)
	}
	//Set2SmartContract()
}

func cfginit(bc *config.BtlClientConf) *config.BtlClientConf {
	cfg := bc
	//if remoteethaccesspoint != "" {
	//	cfg.EthAccessPoint = remoteethaccesspoint
	//}
	//if cmdroottcpport > 0 && cmdroottcpport < 65535 {
	//	cfg.TcpPort = cmdroottcpport
	//}
	//if cmdropstennap != "" {
	//	cfg.RopstenNAP = cmdropstennap
	//}
	//if cmdbastokenaddr != "" {
	//	cfg.TokenAddr = cmdbastokenaddr
	//}
	//if cmdbasmgraddr != "" {
	//	cfg.MgrAddr = cmdbasmgraddr
	//}
	//if cmddohserverport > 0 && cmddohserverport < 65535 {
	//	cfg.DohServerPort = cmddohserverport
	//}
	//
	//if cmdcertfile != "" {
	//	cfg.CertFile = cmdcertfile
	//}
	//if cmdkeyfile != "" {
	//	cfg.KeyFile = cmdkeyfile
	//}
	//if cmdquerydnstimeout != 0 {
	//	cfg.TimeOut = cmdquerydnstimeout
	//}
	//if cmdquerydnstrytimes != 0 {
	//	cfg.TryTimes = cmdquerydnstrytimes
	//}

	return cfg

}

func init() {
	//cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.app.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//rootCmd.Flags().IntVarP(&cmdroottcpport, "tcp-listen-port", "t", 65566, "local tcp listen port")
	//rootCmd.Flags().IntVarP(&cmdrootudpport, "udp-listen-port", "u", 65566, "local udp listen port")
	//rootCmd.Flags().StringVarP(&cmdropstennap, "ropsten-network-access-point", "r", "", "ropsten network access point")
	//rootCmd.Flags().StringVarP(&cmdbastokenaddr, "bas-token-address", "a", "", "bas token address")
	//rootCmd.Flags().StringVarP(&cmdbasmgraddr, "bas-mgr-address", "m", "", "bas manager address")
	rootCmd.Flags().StringVarP(&cmdconfigfilename, "config-file-name", "c", "", "configuration file name")
	//rootCmd.Flags().StringVarP(&keypassword, "password", "p", "", "password for key encrypt")
	//rootCmd.Flags().IntVarP(&cmddohserverport, "doh-listen-port", "p", 65566, "local doh server listen port")
	//rootCmd.Flags().StringVarP(&cmdcertfile, "cert-file", "f", "", "certificate file for tls")
	//rootCmd.Flags().StringVarP(&cmdkeyfile, "key-file", "k", "", "private key file for tls")
	//rootCmd.Flags().StringVarP(&cmddnspath, "dns-query-path", "q", "", "path for dns query")
	//rootCmd.Flags().IntVarP(&cmdquerydnstimeout, "dns-query-time", "o", 0, "max time for wait remote dns server reply")
	//rootCmd.Flags().IntVarP(&cmdquerydnstrytimes, "dns-query-times", "s", 0, "max times for sending dns to remote dns server ")
}
