// Copyright 2017 Pilosa Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctl

import (
	"crypto/tls"

	"github.com/pilosa/pilosa"
	"github.com/spf13/pflag"
)

// CommandWithTLSSupport is the interface for commands which has TLS settings
type CommandWithTLSSupport interface {
	TLSHost() string
	TLSConfiguration() pilosa.TLSConfig
}

// SetTLSConfig creates common TLS flags
func SetTLSConfig(flags *pflag.FlagSet, certificatePath *string, certificateKeyPath *string, skipVerify *bool) {
	flags.StringVarP(certificatePath, "tls.certificate", "", "", "TLS certificate path (usually has the .crt or .pem extension")
	flags.StringVarP(certificateKeyPath, "tls.key", "", "", "TLS certificate key path (usually has the .key extension")
	flags.BoolVarP(skipVerify, "tls.skip-verify", "", false, "Skip TLS certificate verification (not secure)")
}

// CommandClient returns a pilosa.InternalHTTPClient for the command
func CommandClient(cmd CommandWithTLSSupport) (*pilosa.InternalHTTPClient, error) {
	tlsConfig := cmd.TLSConfiguration()
	var TLSConfig *tls.Config
	if tlsConfig.CertificatePath != "" && tlsConfig.CertificateKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(tlsConfig.CertificatePath, tlsConfig.CertificateKeyPath)
		if err != nil {
			return nil, err
		}
		TLSConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: tlsConfig.SkipVerify,
		}
	}
	client, err := pilosa.NewInternalHTTPClient(cmd.TLSHost(), pilosa.GetHTTPClient(TLSConfig))
	if err != nil {
		return nil, err
	}
	return client, err
}
