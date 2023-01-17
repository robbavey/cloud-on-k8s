// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import (
	common_name "github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/name"
)

const (
	httpServiceSuffix = "http"
	configSuffix      = "config"
	deploymentSuffix  = "deployment"
)

// Namer is a Namer that is configured with the defaults for resources related to an Agent resource.
var Namer = common_name.NewNamer("logstash")

// ConfigSecretName returns the name of a secret used to storage Logstash configuration data.
func ConfigSecretName(name string) string {
	return Namer.Suffix(name, "secret")
}

func ConfigMapName(name string) string {
	return Namer.Suffix(name, "config")
}

// Name returns the name of Logstash.
func Name(name string) string {
	return Namer.Suffix(name)
}

// HTTPServiceName returns the name of the HTTP service for a given Logstash name.
func HTTPServiceName(name string) string {
	return Namer.Suffix(name, httpServiceSuffix)
}

// EnvVarsSecretName returns the name of the secret used to storage environmental variables
// for a given Logstash name.
func EnvVarsSecretName(name string) string {
	return Namer.Suffix(name, "envvars")
}

func DeploymentName(name string) string {
	return Namer.Suffix(name, deploymentSuffix)
}

func ConfigName(name string) string {
	return Namer.Suffix(name, configSuffix)
}
