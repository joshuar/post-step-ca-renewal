/*
Copyright © 2023 Joshua Rich <joshua.rich@gmail.com>
*/
package cmd

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joshuar/post-step-ca-renewal/internal/actions"
	"github.com/joshuar/post-step-ca-renewal/internal/certs"
	"github.com/joshuar/post-step-ca-renewal/internal/config"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	defaultConfigFile = "/etc/post-step-ca-renewal/config.yml"
	configFile        string
	debugFlag         bool
	profileFlag       bool
)

var rootCmd = &cobra.Command{
	Use:   "post-step-ca-renewal",
	Short: "Copy certs elsewhere when step-ca updates them.",
	Long:  `A daemon that watches for cert renewals from step-ca and then copies them elsewhere for use by other services.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if debugFlag {
			log.SetLevel(log.DebugLevel)
			log.Debug("debug logging enabled")
		}
		if profileFlag {
			go func() {
				log.Info(http.ListenAndServe("localhost:6060", nil))
			}()
			log.Info("profiling is enabled and available at localhost:6060")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		certConfig := config.GetCertDetails()
		allActions := config.GetActions()
		watchCerts(allActions, certConfig)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", defaultConfigFile, "config file")
	rootCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "debug output (default is false)")
	rootCmd.Flags().BoolVarP(&profileFlag, "profile", "p", false, "enable profiling (default is false)")
	cobra.OnInitialize(initConfig)

}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home + ".post-step-ca-renewal")
		viper.AddConfigPath("/etc/post-step-ca-renewal/")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("using config file: %s", viper.ConfigFileUsed())
	} else {
		log.Fatalf("error reading config file %s: %v", viper.ConfigFileUsed(), err)
	}
}

func performActions(changed string, certConfig *certs.CertConfig, allActions *actions.AllActions) {

	doWork := func(done <-chan interface{}, action *actions.Action, wg *sync.WaitGroup, result chan<- string) {
		started := time.Now()
		defer wg.Done()

		if action.PreCommand != nil {
			for _, cmd := range action.PreCommand {
				log.Debugf("%s: running pre-command %s", action.Name, cmd)
				run(cmd)
			}
		}
		// switch filepath.Base(changed) {
		// case filepath.Base(certConfig.Cert):
		log.Debugf("%s - copying cert %s to %s", action.Name, certConfig.Cert, action.Cert)
		_, err := copy(certConfig.Cert, action.Cert)
		if err != nil {
			log.Errorf("%s: failed to copy, %v", action.Name, err)
		}
		// case filepath.Base(certConfig.Key):
		log.Debugf("%s: copying key %s to %s", action.Name, certConfig.Key, action.Key)
		_, err = copy(certConfig.Key, action.Key)
		if err != nil {
			log.Errorf("%s: failed to copy, %v", action.Name, err)
		}
		// }
		if action.FullChain != "" {
			log.Debugf("%s: creating fullchain file %s", action.Name, action.FullChain)
			_, err := concat(action.FullChain, certConfig.Cert, certConfig.CA)
			if err != nil {
				log.Errorf("%s: failed to create fullchain, %v", action.Name, err)
			}
		}
		if action.PostCommand != nil {
			for _, cmd := range action.PostCommand {
				log.Debugf("%s: running post-command %s", action.Name, cmd)
				run(cmd)
			}
		}
		select {
		case <-done:
		case result <- action.Name:
		}

		took := time.Since(started)

		log.Debugf("action %s took %v\n", action.Name, took)
	}

	done := make(chan interface{})
	result := make(chan string)

	var wg sync.WaitGroup
	wg.Add(allActions.Count())

	for i := 0; i < allActions.Count(); i++ {
		go doWork(done, allActions.GetActionByIndex(i), &wg, result)
	}

	close(done)
	wg.Wait()
}

func copy(src, dst string) (int64, error) {
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func concat(dst string, srcs ...string) (int64, error) {
	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	var totalBytes int64
	for _, s := range srcs {
		source, err := os.Open(s)
		if err != nil {
			return 0, err
		}
		defer source.Close()
		nBytes, err := io.Copy(destination, source)
		if err != nil {
			return nBytes, err
		} else {
			totalBytes += nBytes
		}
	}
	return totalBytes, err
}

func run(command string) {
	cmdArray := strings.Split(command, " ")
	cmd := exec.Command(cmdArray[0], cmdArray[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("error running command %v: %s", err, string(out))
	} else {
		log.Debugf("%s\n", string(out))
	}
}

func watchCerts(allActions *actions.AllActions, certs *certs.CertConfig) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("creating a new watcher: %s", err)
	}
	defer w.Close()

	go fileLoop(w, certs, allActions)

	err = w.Add(certs.Cert)
	if err != nil {
		log.Errorf("%q: %s", certs.Cert, err)
	} else {
		log.Debugf("added watch for %s", certs.Cert)
	}

	// for _, p := range []string{certs.Cert, certs.Key} {
	// 	err = w.Add(p)
	// 	if err != nil {
	// 		log.Errorf("%q: %s", p, err)
	// 	} else {
	// 		log.Debugf("added watch for %s", p)
	// 	}
	// }

	<-make(chan struct{}) // Block forever
}

func fileLoop(w *fsnotify.Watcher, certs *certs.CertConfig, allActions *actions.AllActions) {
	for {
		select {
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Errorf("problem watching file: %s", err)
		case e, ok := <-w.Events:
			if !ok {
				return
			}

			switch {
			case e.Op.Has(fsnotify.Write):
				performActions(e.Name, certs, allActions)
			}

			log.Debugf("observed event %s", e)
		}
	}
}
