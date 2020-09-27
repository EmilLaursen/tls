package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/EmilLaursen/tls/libraries/transliteration"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use: "tls file1.txt file2.txt",
	Example: `cat file1.txt | tls
tls -v -o /my/awesome/dir file1.txt file2.txt file3.txt
	`,
	Short: "Transliterates stdin to ASCII, and also preserves æøå§, then outputs to stdout",
	Long: `Transliterates stdin to ASCII, and also preserves æøå§, then outputs to stdout.
Supplied file arguments are processed concurrently. There is no concurrency bound
so do not supply a large amount of files, unless you want to use alot of file
descriptors.

		tls original/path/file1.txt

will place the transliterated file next to file1.txt at

		original/path/file1-transliterated.txt
`,
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
		MB := 1000 * 1000

		verbose := viper.GetBool("verbose")
		out := viper.GetString("output-dir")

		doProcessStdIn := len(args) <= 0
		if doProcessStdIn {
			ProcessStdIn(cmd)
			os.Exit(0)
		}

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

		p := mpb.New(
			mpb.WithWaitGroup(&wg),
			mpb.WithWidth(90),
			mpb.WithRefreshRate(250*time.Millisecond),
		)

		start := time.Now()
		totalBytes := int64(0)

		for _, arg := range args {
			outputDir := getOutputDir(arg)
			outputDir, err := filepath.Abs(outputDir)
			if err != nil {
				log.Fatalf("Absolute path error: %+v", errors.Wrap(err, ""))
			}

			os.MkdirAll(outputDir, 0755)

			extension := filepath.Ext(arg)
			filename := strings.TrimSuffix(filepath.Base(arg), extension)

			transliteratedFile := fmt.Sprintf("%s-transliterated%s", filename, extension)

			outputFilePath := filepath.Join(outputDir, transliteratedFile)

			// Verbose related...
			fi, err := os.Stat(arg)
			if err != nil {
				log.Fatal("file stat error %+v", errors.Wrap(err, ""))
			}

			fileSize := fi.Size()
			atomic.AddInt64(&totalBytes, fileSize)

			bar := p.AddBar(fi.Size(), mpb.BarStyle("[=>-|"),
				mpb.PrependDecorators(
					decor.Name(filepath.Base(arg)),
					decor.CountersKiloByte(" % .2f / % .2f"),
				),
				mpb.AppendDecorators(
					decor.AverageETA(decor.ET_STYLE_GO),
					decor.Name(" ] "),
					decor.AverageSpeed(decor.UnitKB, "% .1f"),
				),
			)
			// Verbose related...

			wg.Add(1)
			go func(inputFilePath, outputFilePath string) {
				defer wg.Done()

				inputFile, err := os.Open(inputFilePath)
				if err != nil {
					log.Fatalf("Error opening file %s: %+v", inputFilePath, errors.Wrap(err, ""))
				}

				defer inputFile.Close()

				f, err := os.Create(outputFilePath)
				if err != nil {
					log.Fatalf("failed to create file: %+v", errors.Wrap(err, ""))
				}
				defer f.Close()

				var reader *bufio.Reader
				if verbose {
					reader = bufio.NewReader(bar.ProxyReader(inputFile))
				} else {
					reader = bufio.NewReader(inputFile)
				}

				daTrans.Process(reader, bufio.NewWriter(f))
			}(arg, outputFilePath)
		}

		p.Wait()

		if verbose {
			totalTime := time.Since(start)
			mbPerSec := float64(totalBytes/int64(MB)) / totalTime.Seconds()
			fmt.Printf("Processing speed: %.1f MB/s", mbPerSec)
		}

		os.Exit(0)
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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show progress bars")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.PersistentFlags().StringP("output-dir", "o", "", "custom output directory where transliterated files are placed")
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

func ProcessStdIn(cmd *cobra.Command) {

	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Panic("file.stat()", err)
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		err := cmd.Help()
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	stdin := bufio.NewReader(cmd.InOrStdin())
	stdout := bufio.NewWriter(cmd.OutOrStdout())

	daTrans := transliteration.NewDanishTransliterator()

	daTrans.Process(stdin, stdout)
}
