package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	log "github.com/sirupsen/logrus"
)

type Foremand struct {
	Prefix  string
	Cli     *clientv3.Client
	Manager map[string]*exec.Cmd
}

func InitForemand(config Config) (*Foremand, error) {
	log.WithFields(log.Fields{
		"endpoints": config.Endpoints,
		"prefix":    config.Prefix,
	}).Info("Initialing foremand")
	cfg := clientv3.Config{
		Endpoints:            config.Endpoints,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
	}

	if config.BasicAuth {
		cfg.Username = config.Username
		cfg.Password = config.Password
	}

	if config.Cert != "" && config.Key != "" {
		config.TLSEnabled = true
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if config.CaCert != "" {
		certBytes, err := ioutil.ReadFile(config.CaCert)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
	}

	if config.Cert != "" && config.Key != "" {
		tlsCert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
	}

	if config.TLSEnabled {
		cfg.TLS = tlsConfig
	}

	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	fmd := &Foremand{
		Prefix:  config.Prefix,
		Cli:     cli,
		Manager: map[string]*exec.Cmd{},
	}
	return fmd, nil
}

func (fmd *Foremand) Start(doneChan chan bool, errChan chan error) {
	log.Info("Starting foremand")
	resp, err := fmd.Cli.Get(context.Background(), config.Prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("etcd not reachable")
		doneChan <- true
		return
	}
	log.Debug("reading etcd")
	for _, kv := range resp.Kvs {
		app := string(kv.Key)
		shell := string(kv.Value)

		if err := fmd.fork(app, shell); err != nil {
			errChan <- err
		}
	}

	log.Debug("watching etcd")
	watchChan := fmd.Cli.Watch(context.Background(), config.Prefix)
	for watchResp := range watchChan {
		for _, evt := range watchResp.Events {
			app := string(evt.Kv.Key)
			shell := string(evt.Kv.Value)
			switch evt.Type {
			case mvccpb.PUT:
				if err := fmd.fork(app, shell); err != nil {
					errChan <- err
				}

			case mvccpb.DELETE:
				fmd.tryKill(app)
			}
		}
	}
}

func (fmd *Foremand) Stop() {
	log.Info("Stopping foremand")
	fmd.Cli.Close()
	for _, proc := range fmd.Manager {
		proc.Process.Kill()
	}
}

func (fmd *Foremand) fork(app, shell string) error {
	log.WithFields(log.Fields{
		"app":   app,
		"shell": shell,
	}).Info("forking")

	fmd.tryKill(app)

	cmd := exec.Command("bash", "-c", shell)
	cmd.Env = os.Environ()
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.ExtraFiles = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	fmd.Manager[app] = cmd
	return cmd.Start()
}

func (fmd *Foremand) tryKill(app string) {
	if fmd.Manager[app] != nil {
		log.WithFields(log.Fields{
			"app": app,
			"pid": fmd.Manager[app].Process.Pid,
		}).Info("killing")
		fmd.Manager[app].Process.Kill()
		delete(fmd.Manager, app)
	}
}
