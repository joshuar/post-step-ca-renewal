/*
Copyright © 2023 Joshua Rich <joshua.rich@gmail.com>
*/
package config

import (
	"os"

	"github.com/joshuar/post-step-ca-renewal/internal/actions"
	"github.com/joshuar/post-step-ca-renewal/internal/certs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GetCertDetails() *certs.CertConfig {
	cert := viper.GetString("cert")
	log.Debugf("using %s for cert", cert)
	key := viper.GetString("key")
	log.Debugf("using %s for key", key)
	ca := viper.GetString("ca")
	log.Debugf("using %s for ca", cert)

	for _, f := range []string{cert, key, ca} {
		sourceFileStat, err := os.Stat(f)
		if err != nil {
			log.Fatalf("%s: %v", f, err)
		}
		if !sourceFileStat.Mode().IsRegular() {
			log.Fatalf("%s is not a regular file", f)
		}
	}

	return &certs.CertConfig{
		Cert: cert,
		Key:  key,
		CA:   ca,
	}
}

func GetActions() *actions.AllActions {
	a := &actions.AllActions{}
	viper.UnmarshalKey("actions", &a.ActionList)
	if a.Count() == 0 {
		log.Info("no actions have been defined, no need to do anything.")
		os.Exit(0)
	}
	log.Debugf("found %d actions to process.", a.Count())
	return a
}
