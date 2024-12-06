// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prober

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/blackbox_exporter/config"
)

type ProbeResult struct {
	success        bool
	failureReason  string
	failureDetails []string
}

// Creates the ProbeResult for a failed probe.
//
// Expects an odd number of string arguments.
//
// Example:
// Calling probeFailure("problem", "label1", "value1", "label2", "value2")
// will result in the metric:
//
// `probe_failure_info{reason="problem", label1="value1", label2="value2"}`
//
// The corresponding gauge can be obtained with the ProbeResult.failureInfoGauge()
// method.
func ProbeFailure(reason string, details ...string) ProbeResult {
	if len(details)%2 != 0 {
		panic("Must be called with an odd number of string arguments.")
	}

	return ProbeResult{success: false, failureReason: reason, failureDetails: details}
}

func ProbeSuccess() ProbeResult {
	return ProbeResult{success: true, failureReason: "", failureDetails: nil}
}

func (r *ProbeResult) failureInfoGauge() *prometheus.GaugeVec {
	labels := append([]string{"reason", r.failureReason}, r.failureDetails...)
	return prometheus.NewGaugeVec(probeFailureInfo, labels)
}

func (r *ProbeResult) log(logger *slog.Logger, duration float64) {
	if r.success {
		logger.Info("Probe succeeded", "duration_seconds", duration)
	} else {
		// converting the []string slice to an []any slice is a bit finicky
		logDetails := make([]any, 0, len(r.failureDetails)+2)
		for _, d := range r.failureDetails {
			logDetails = append(logDetails, d)
		}
		logger.Error(r.failureReason, logDetails...)
		logger.Error("Probe failed", "duration_seconds", duration)
	}
}

type ProbeFn func(ctx context.Context, target string, config config.Module, registry *prometheus.Registry, logger *slog.Logger) ProbeResult

const (
	helpSSLEarliestCertExpiry     = "Returns last SSL chain expiry in unixtime"
	helpSSLChainExpiryInTimeStamp = "Returns last SSL chain expiry in timestamp"
	helpProbeTLSInfo              = "Returns the TLS version used or NaN when unknown"
	helpProbeTLSCipher            = "Returns the TLS cipher negotiated during handshake"
	helpProbeFailureInfo          = "Return the reason a Probe failed"
)

var (
	sslEarliestCertExpiryGaugeOpts = prometheus.GaugeOpts{
		Name: "probe_ssl_earliest_cert_expiry",
		Help: helpSSLEarliestCertExpiry,
	}

	sslChainExpiryInTimeStampGaugeOpts = prometheus.GaugeOpts{
		Name: "probe_ssl_last_chain_expiry_timestamp_seconds",
		Help: helpSSLChainExpiryInTimeStamp,
	}

	probeTLSInfoGaugeOpts = prometheus.GaugeOpts{
		Name: "probe_tls_version_info",
		Help: helpProbeTLSInfo,
	}

	probeTLSCipherGaugeOpts = prometheus.GaugeOpts{
		Name: "probe_tls_cipher_info",
		Help: helpProbeTLSCipher,
	}

	probeFailureInfo = prometheus.GaugeOpts{
		Name: "probe_failure_info",
		Help: helpProbeFailureInfo,
	}
)
