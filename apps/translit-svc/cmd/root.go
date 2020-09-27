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
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/EmilLaursen/tls/libraries/transliteration"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "tls",
	Short: "Transliterates stdin to ASCII, and also preserves æøå§, then outputs to stdout",
	Long:  `Transliterates stdin to ASCII, and also preserves æøå§, then outputs to stdout`,
	Args: func(cmd *cobra.Command, args []string) error {

		HIGH_NUMBER_OF_GOROUTINES := 512

		if len(args) > HIGH_NUMBER_OF_GOROUTINES {
			return errors.Errorf("Relax with the number of files already...")
		}

		for _, arg := range args {
			if _, err := os.Stat(arg); err != nil {
				return errors.Errorf("arguments file %s does not exist", arg)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		out := viper.GetString("output-dir")
		var getOutputDir func(arg string) string

		if len(out) > 0 {
			getOutputDir = func(arg string) string {
				return out
			}
		} else {
			getOutputDir = func(arg string) string {
				return filepath.Dir(arg)
			}
		}

		daTrans := transliteration.NewDanishTransliterator()

		var wg sync.WaitGroup

		for _, arg := range args {
			outputDir := getOutputDir(arg)
			outputDir, err := filepath.Abs(outputDir)
			if err != nil {
				log.Fatalf("Absolute path error: %+v", err)
			}

			os.MkdirAll(outputDir, 0755)

			extension := filepath.Ext(arg)
			filename := strings.TrimSuffix(filepath.Base(arg), extension)

			transliteratedFile := fmt.Sprintf("%s-transliterated%s", filename, extension)

			outputFilePath := filepath.Join(outputDir, transliteratedFile)

			wg.Add(1)
			go func(inputFilePath, outputFilePath string) {
				defer wg.Done()

				inputFile, err := os.Open(inputFilePath)
				if err != nil {
					log.Fatalf("Error opening file %s: %+v", inputFilePath, err)
				}

				defer inputFile.Close()

				f, err := os.Create(outputFilePath)
				if err != nil {
					log.Fatalf("failed to create file: %+v", errors.Wrap(err, ""))
				}
				defer f.Close()

				daTrans.Process(bufio.NewReader(inputFile), bufio.NewWriter(f))
			}(arg, outputFilePath)
		}
		wg.Wait()

		if len(args) > 0 {
			os.Exit(0)
		}

		fi, err := os.Stdin.Stat()
		if err != nil {
			log.Panic("file.stat()", err)
		}

		if fi.Mode()&os.ModeNamedPipe == 0 || fi.Size() <= 0 {
			err := cmd.Help()
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		stdin := bufio.NewReader(cmd.InOrStdin())
		stdout := bufio.NewWriter(cmd.OutOrStdout())

		daTrans.Process(stdin, stdout)
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

func init() {
	rootCmd.PersistentFlags().StringP("output-dir", "o", "", "output directory")
	viper.BindPFlag("output-dir", rootCmd.PersistentFlags().Lookup("output-dir"))

	cobra.OnInitialize(initConfig)
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

		// Search config in home directory with name ".testcli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".tls")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
