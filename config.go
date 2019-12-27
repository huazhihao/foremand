package main

import (
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	LogLevel   string `toml:"log-level"`
	ConfigFile string `toml:"config"`
	Prefix     string
	Endpoints  []string

	BasicAuth bool `toml:"auth"`
	Username  string
	Password  string

	TLSEnabled bool   `toml:"tls"`
	CaCert     string `toml:"ca-cert"`
	Cert       string `toml:"cert"`
	Key        string `toml:"key"`
}

var config Config

func init() {
	flag.StringVar(&config.LogLevel, "log-level", "info", "debug,info,warning,debug,panic,fatal")
	flag.StringVar(&config.ConfigFile, "config-file", "/etc/foremand/foremand.toml", "config file")
	flag.StringVar(&config.Prefix, "prefix", "foremand", "etcd path prefix for watching")
	endpoints := flag.String("endpoints", "http://127.0.0.1:2379", "etcd endpoints")
	config.Endpoints = strings.Split(*endpoints, ",")

	flag.BoolVar(&config.BasicAuth, "auth", false, "enabling etcd basic auth")
	flag.StringVar(&config.Username, "username", "", "etcd basic auth username")
	flag.StringVar(&config.Password, "password", "", "etcd basic auth password")

	flag.BoolVar(&config.TLSEnabled, "tls", false, "enabling etcd tls")
	flag.StringVar(&config.CaCert, "ca-cert", "", "etcd tls ca cert")
	flag.StringVar(&config.Cert, "cert", "", "etcd tls cert")
	flag.StringVar(&config.Key, "key", "", "etcd tls key")

}

func initConfig() error {
	_, err := os.Stat(config.ConfigFile)
	if os.IsNotExist(err) {
		log.Debug("Skipping foremand config file.")
	} else {
		log.Debug("Loading " + config.ConfigFile)
		configBytes, err := ioutil.ReadFile(config.ConfigFile)
		if err != nil {
			return err
		}

		_, err = toml.Decode(string(configBytes), &config)
		if err != nil {
			return err
		}
	}

	if config.LogLevel != "" {
		logLevel, err := log.ParseLevel(config.LogLevel)
		if err != nil {
			return err
		}
		log.SetLevel(logLevel)
	}
	return nil
}
