/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/const-tmp/consul-watch-notify/consul"
	"github.com/const-tmp/consul-watch-notify/notify"
	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/consul/api/watch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

// consulCmd represents the consul command
var consulCmd = &cobra.Command{
	Use:   "consul",
	Short: "consul watch",
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(viper.BindEnv("consul.address", "CONSUL_HTTP_ADDR"))
		cobra.CheckErr(viper.BindEnv("consul.token", "CONSUL_HTTP_TOKEN"))
		viper.SetDefault("consul.address", "http://localhost:8500")
		token := viper.GetString("consul.token")
		address := viper.GetString("consul.address")
		viper.WatchConfig()

		plans := make(map[string]*watch.Plan)
		for _, watchType := range viper.GetStringSlice("consul.watches") {
			plan, err := watch.Parse(map[string]interface{}{
				"token": token,
				"type":  watchType,
			})
			if err != nil {
				log.Fatalf("parse %s plan error: %s", watchType, err.Error())
			}
			plans[watchType] = plan
		}

		setNotifiersForPlans(plans, viper.Sub("notifiers"))

		errCh := make(chan struct {
			name string
			err  error
		})

		viper.OnConfigChange(func(_ fsnotify.Event) {
			log.Println("reloading config")

			var currentWatches []string
			for watchType := range plans {
				currentWatches = append(currentWatches, watchType)
			}

			newWatches := viper.GetStringSlice("consul.watches")

			var plansToStart []*watch.Plan
			for _, watchType := range newWatches {
				// add watch
				if !keyExists(watchType, currentWatches) {
					plan, err := watch.Parse(map[string]interface{}{
						"token": token,
						"type":  watchType,
					})
					if err != nil {
						log.Printf("parse %s plan error: %s", watchType, err.Error())
						continue
					}
					plans[watchType] = plan
					plansToStart = append(plansToStart, plan)
				}
			}

			// remove watch
			for _, watchType := range currentWatches {
				if !keyExists(watchType, newWatches) {
					log.Printf("stop and remove %s watch", watchType)
					plans[watchType].Stop()
					delete(plans, watchType)
				}
			}

			setNotifiersForPlans(plans, viper.Sub("notifiers"))

			for _, plan := range plansToStart {
				go func(p *watch.Plan) {
					errCh <- struct {
						name string
						err  error
					}{name: p.Type, err: p.Run(address)}
				}(plan)
			}
		})

		for _, plan := range plans {
			go func(p *watch.Plan) {
				errCh <- struct {
					name string
					err  error
				}{name: p.Type, err: p.Run(address)}
			}(plan)
		}

		for ch := range errCh {
			log.Printf("%s error: %s", ch.name, ch.err.Error())
		}
	},
}

func keyExists(key string, slice []string) bool {
	for _, s := range slice {
		if key == s {
			return true
		}
	}
	return false
}

func newNotifyConfig(templates *viper.Viper, notifier notify.Interface, notifierName, watchType string) consul.NotifyConfig {
	notifyConfig := consul.NotifyConfig{Name: notifierName}
	notifyConfig.RegisterTemplate, notifyConfig.ChangeTemplate, notifyConfig.DeregisterTemplate = consul.TemplatesFromViper(
		fmt.Sprintf("%s.%s", notifierName, watchType),
		templates.Sub(watchType),
	)
	notifyConfig.Notifier = notifier
	return notifyConfig
}

func setPlanHandler(plan *watch.Plan, templates *viper.Viper, notifier notify.Interface, notifierName, watchType string) {
	nc := newNotifyConfig(templates, notifier, notifierName, watchType)
	switch watchType {
	case "checks":
		w := consul.NewChecksWatcher()
		w.Notifiers = append(w.Notifiers, nc)
		plan.Handler = w.HandlerFactory()
	case "services":
		w := consul.NewServicesWatcher()
		w.Notifiers = append(w.Notifiers, nc)
		plan.Handler = w.HandlerFactory()
	case "nodes":
		w := consul.NewNodesWatcher()
		w.Notifiers = append(w.Notifiers, nc)
		plan.Handler = w.HandlerFactory()
	default:
		log.Printf("unknown watch type %s", watchType)
	}
}

func setNotifiersForPlans(plans map[string]*watch.Plan, notifiers *viper.Viper) {
	for notifierName := range notifiers.AllSettings() {
		notifierCfg := notifiers.Sub(notifierName)
		notifier, err := notify.NewFromViper(notifierCfg)
		if err != nil {
			log.Fatalf("notifier %s error: %s", notifierName, err.Error())
		}
		templates := notifierCfg.Sub("templates")
		for watchType, plan := range plans {
			setPlanHandler(plan, templates, notifier, notifierName, watchType)
		}
	}
}

func init() {
	rootCmd.AddCommand(consulCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// consulCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// consulCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
