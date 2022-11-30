// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package logstash

import "github.com/elastic/cloud-on-k8s/v2/pkg/controller/common/name"

const (
	httpServiceSuffix = "http"
	configSuffix      = "config"
)

// LogstashNamer is a Namer that is configured with the defaults for resources related to a Logstash resource.
var LogstashNamer = name.NewNamer("logstash")

func HTTPService(logstashName string) string {
	return LogstashNamer.Suffix(logstashName, httpServiceSuffix)
}

func Deployment(logstashName string) string {
	return LogstashNamer.Suffix(logstashName)
}

func Config(logstashName string) string {
	return LogstashNamer.Suffix(logstashName, configSuffix)
}
