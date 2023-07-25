/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/const-tmp/consul-watch-notify/check"
	"github.com/const-tmp/consul-watch-notify/config"
	"github.com/const-tmp/consul-watch-notify/notify"
	"github.com/const-tmp/consul-watch-notify/utils"
	"github.com/const-tmp/consul-watch-notify/watcher"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"html/template"
	"log"
	"os"
	"os/signal"
	"time"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "watcher",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("period", 15*time.Second)

		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt)
		ctx := context.Background()

		var notifiers []watcher.NotifyConfig

		notifierCfg := viper.Sub("notifiers")
		for k := range notifierCfg.AllSettings() {
			fmt.Println(k)
			cfg, err := config.NewFromViper[watcher.NotifyConfig](notifierCfg.Sub(k), func(v *viper.Viper, w *watcher.NotifyConfig) error {
				tmplText := v.GetString("template")
				if tmplText == "" {
					return fmt.Errorf("%s: template is required", k)
				}
				tmpl, err := template.New("watcher").
					Funcs(utils.FuncMap).
					Parse(tmplText)
				if err != nil {
					return fmt.Errorf("template parse error: %w", err)
				}
				if err := viper.Unmarshal(w); err != nil {
					return fmt.Errorf("%s: unmarshal error: %w", k, err)
				}
				notifier, err := notify.NewFromViper(v)
				if err != nil {
					return fmt.Errorf("create notifier error: %w", err)
				}
				w.Notifier = notifier
				w.Template = tmpl
				w.Name = k
				return nil
			})
			if err != nil {
				log.Fatal("notifier error:", err)
			}
			log.Println(cfg)
			notifiers = append(notifiers, cfg)
		}

		for s := range viper.GetStringMap("checks") {
			v := viper.Sub("checks." + s)
			v.SetDefault("period", viper.GetDuration("period"))

			chk, err := check.NewFromViper(v)
			if err != nil {
				log.Printf("create %s check error: %s", s, err.Error())
				os.Exit(1)
			}

			go watcher.Watcher{
				Name:      s,
				Period:    v.GetDuration("period"),
				Check:     chk,
				Notifiers: notifiers,
			}.Start(ctx)
		}

		log.Printf("received %s", <-ch)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.watcher.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".watcher" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".watcher")
	}

	log.Default().SetFlags(log.LstdFlags | log.Lmsgprefix | log.Lshortfile)
	log.Default().SetPrefix("[ watcher ]\t")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
