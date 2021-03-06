/*
Copyright 2017 The Kubernetes Authors.

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

// Refresh retries Github status updates for stale PR statuses.
package main

import (
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
)

var (
	configPath        = flag.String("config-path", "/etc/config/config", "Path to config.yaml.")
	port              = flag.Int("port", 8888, "Port to listen on.")
	dryRun            = flag.Bool("dry-run", true, "Dry run for testing. Uses API tokens but does not mutate.")
	webhookSecretFile = flag.String("hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing the GitHub HMAC secret.")
	githubTokenFile   = flag.String("github-token-file", "/etc/github/oauth", "Path to the file containing the GitHub OAuth token.")
	githubEndpoint    = flag.String("github-endpoint", "https://api.github.com", "GitHub's API endpoint.")
	prowURL           = flag.String("prow-url", "", "Prow frontend URL.")
)

func validateFlags() error {
	if *prowURL == "" {
		return errors.New("--prow-url needs to be specified")
	}
	_, err := url.Parse(*githubEndpoint)
	return err
}

func main() {
	flag.Parse()
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Ignore SIGTERM so that we don't drop hooks when the pod is removed.
	// We'll get SIGTERM first and then SIGKILL after our graceful termination
	// deadline.
	signal.Ignore(syscall.SIGTERM)

	if err := validateFlags(); err != nil {
		logrus.WithError(err).Fatal("Error validating flags.")
	}

	configAgent := &config.Agent{}
	if err := configAgent.Start(*configPath); err != nil {
		logrus.WithError(err).Fatal("Error starting config agent.")
	}

	webhookSecretRaw, err := ioutil.ReadFile(*webhookSecretFile)
	if err != nil {
		logrus.WithError(err).Fatal("Could not read webhook secret file.")
	}
	webhookSecret := bytes.TrimSpace(webhookSecretRaw)

	oauthSecretRaw, err := ioutil.ReadFile(*githubTokenFile)
	if err != nil {
		logrus.WithError(err).Fatal("Could not read oauth secret file.")
	}
	oauthSecret := string(bytes.TrimSpace(oauthSecretRaw))

	ghc := github.NewClient(oauthSecret, *githubEndpoint)
	if *dryRun {
		ghc = github.NewDryRunClient(oauthSecret, *githubEndpoint)
	}

	server := NewServer(oauthSecret, webhookSecret, ghc, *prowURL, configAgent)

	http.Handle("/", server)
	logrus.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
