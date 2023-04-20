package main

import (
	"context"
	"flag"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/signal"
	"syscall"

	"net/http"
	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

var (
	version = "0.0.0"
)

func main() {
	cfgFile := flag.String("config", "", "Path to a config file")
	flag.Parse()

	if *cfgFile == "" {
		log.Fatalf("Config file required")
	}

	cfg, err := configLoad(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.ListenPprof != "" {
		go func() {
			if err := http.ListenAndServe(cfg.ListenPprof, nil); err != nil {
				log.Fatalf("Unable to listen on %s: %s", cfg.ListenPprof, err)
			}
		}()
	}

	if cfg.LogLevel != "" {
		lvl, err := log.ParseLevel(cfg.LogLevel)
		if err != nil {
			log.Fatalf("Unable to parse log level: %s", err)
		}

		log.SetLevel(lvl)
	}

	fmt.Printf("k8s_api: %s\n", cfg.K8s_api)
	fmt.Printf("k8s_serviceaccount: %s\n", cfg.K8s_serviceaccount)

	/*
		k8sconfig, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(k8sconfig)
		if err != nil {
			panic(err.Error())
		}*/

	clientConfig := rest.Config{}
	clientConfig.BearerToken = cfg.K8s_token
	clientConfig.Host = cfg.K8s_api

	clientset, err := kubernetes.NewForConfig(&clientConfig)
	if err != nil {
		fmt.Printf("error getting Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	if err != nil {
		fmt.Printf("error getting namespaces: %v\n", err)
		os.Exit(1)
	}
	for _, namespace := range namespaces.Items {
		fmt.Printf("namespace %s\n", namespace.Name)
		if org, found := namespace.Labels["appuio.io/organization"]; found {
			fmt.Printf("  org %s\n", org)
		}
	}

	proc := newProcessor(*cfg)

	if err = proc.run(); err != nil {
		log.Fatalf("Unable to start: %s", err)
	}

	log.Warnf("Listening on %s, sending to %s", cfg.Listen, cfg.Target)
	log.Warnf("Started v%s", version)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, os.Interrupt)
	<-ch

	log.Warn("Shutting down, draining requests")
	if err = proc.close(); err != nil {
		log.Errorf("Error during shutdown: %s", err)
	}

	log.Warnf("Finished")
}
