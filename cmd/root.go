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
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	serieNumber string
	titleNumber string
	neo         string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wsoffcli",
	Short: "Collect data from https://ws-tcg.com/ and https://en.ws-tcg.com/.",
	Long: `Collect data from https://ws-tcg.com/ and https://en.ws-tcg.com/.

Create a json file for each card with most information.

Example:
'wsoffcli fetch -n IMC' will fetch all cards with a code starting with 'IMC'

If you want more than one use '##' as seperator like 'wsoffcli fetch -n BD##IM'

'--expansion' uses a number in the official site that is unique for each expansion. '--title' uses a number in the official site that is unique for each title. Title and expansion numbers are distinct values and different between the English and Japanese sites. For example:
  English:
    Title 159 is "Tokyo Revengers"
    Expansion 159 is "BanG Dream! Girls Band Party Premium Booster"
  Japanese:
    Title numbers aren't supported
    Expansion 159 is "Monogatari Series: Second Season"
See doc/expansion_list.md for a list of expansion numbers.

To use environ variable, use the prefix 'WSOFF'.
	 `,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&serieNumber, "expansion", "", "", "expansion number")
	rootCmd.PersistentFlags().StringVarP(&titleNumber, "title", "t", "", "title number")
	rootCmd.PersistentFlags().StringVarP(&neo, "neo", "n", "", "Neo standar by set")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".wsoffcli" (without extension).
		viper.SetEnvPrefix("wsoff")
		viper.AddConfigPath(home)
		viper.SetConfigName(".wsoffcli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
