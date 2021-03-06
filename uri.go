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

package pilosa

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pilosa/pilosa/internal"
)

var schemeRegexp = regexp.MustCompile("^[+a-z]+$")
var hostRegexp = regexp.MustCompile("^[0-9a-z.-]+$|^\\[[:0-9a-fA-F]+\\]$")
var addressRegexp = regexp.MustCompile("^(([+a-z]+):\\/\\/)?([0-9a-z.-]+|\\[[:0-9a-fA-F]+\\])?(:([0-9]+))?$")

// URI represents a Pilosa URI.
// A Pilosa URI consists of three parts:
// 1) Scheme: Protocol of the URI. Default: http.
// 2) Host: Hostname or IP URI. Default: localhost. IPv6 addresses should be written in brackets, e.g., `[fd42:4201:f86b:7e09:216:3eff:fefa:ed80]`.
// 3) Port: Port of the URI. Default: 10101.
//
// All parts of the URI are optional. The following are equivalent:
// 	http://localhost:10101
// 	http://localhost
// 	http://:10101
// 	localhost:10101
// 	localhost
// 	:10101
type URI struct {
	scheme string `json:"scheme"`
	host   string `json:"host"`
	port   uint16 `json:"port"`
}

// DefaultURI creates and returns the default URI.
func DefaultURI() *URI {
	return &URI{
		scheme: "http",
		host:   "localhost",
		port:   10101,
	}
}

type URIs []URI

func (u URIs) HostPortStrings() []string {
	s := make([]string, len(u))
	for i, a := range u {
		s[i] = a.HostPort()
	}
	return s
}

// NewURIFromHostPort returns a URI with specified host and port.
func NewURIFromHostPort(host string, port uint16) (*URI, error) {
	uri := DefaultURI()
	err := uri.SetHost(host)
	if err != nil {
		return nil, err
	}
	uri.SetPort(port)
	return uri, nil
}

// NewURIFromAddress parses the passed address and returns a URI.
func NewURIFromAddress(address string) (*URI, error) {
	return parseAddress(address)
}

// Scheme returns the scheme of this URI.
func (u *URI) Scheme() string {
	return u.scheme
}

// SetScheme sets the scheme of this URI.
func (u *URI) SetScheme(scheme string) error {
	m := schemeRegexp.FindStringSubmatch(scheme)
	if m == nil {
		return errors.New("invalid scheme")
	}
	u.scheme = scheme
	return nil
}

// Host returns the host of this URI.
func (u *URI) Host() string {
	return u.host
}

// SetHost sets the host of this URI.
func (u *URI) SetHost(host string) error {
	m := hostRegexp.FindStringSubmatch(host)
	if m == nil {
		return errors.New("invalid host")
	}
	u.host = host
	return nil
}

// Port returns the port of this URI.
func (u *URI) Port() uint16 {
	return u.port
}

// SetPort sets the port of this URI.
func (u *URI) SetPort(port uint16) {
	u.port = port
}

// HostPort returns `Host:Port`
func (u *URI) HostPort() string {
	// XXX: The following is just to make TestHandler_Status; remove it
	if u == nil {
		return ""
	}
	s := fmt.Sprintf("%s:%d", u.host, u.port)
	return s
}

// Normalize returns the address in a form usable by a HTTP client.
func (u *URI) Normalize() string {
	scheme := u.scheme
	index := strings.Index(scheme, "+")
	if index >= 0 {
		scheme = scheme[:index]
	}
	return fmt.Sprintf("%s://%s:%d", scheme, u.host, u.port)
}

// String returns the address as a string.
func (u URI) String() string {
	return fmt.Sprintf("%s://%s:%d", u.scheme, u.host, u.port)
}

// Equals returns true if the checked URI is equivalent to this URI.
func (u URI) Equals(other *URI) bool {
	if other == nil {
		return false
	}
	return u == *other
}

// Path returns URI with path
func (u *URI) Path(path string) string {
	return fmt.Sprintf("%s%s", u.Normalize(), path)
}

// The following methods are required to implement pflag Value interface.

// Set sets the time quantum value.
func (u *URI) Set(value string) error {
	uri, err := NewURIFromAddress(value)
	if err != nil {
		return err
	}
	*u = *uri
	return nil
}

// Type returns the type of a time quantum value.
func (u URI) Type() string {
	return "URI"
}

func parseAddress(address string) (uri *URI, err error) {
	m := addressRegexp.FindStringSubmatch(address)
	if m == nil {
		return nil, errors.New("invalid address")
	}
	scheme := "http"
	if m[2] != "" {
		scheme = m[2]
	}
	host := "localhost"
	if m[3] != "" {
		host = m[3]
	}
	var port = 10101
	if m[5] != "" {
		port, err = strconv.Atoi(m[5])
		if err != nil {
			return nil, errors.New("converting port string to int")
		}
	}
	uri = &URI{
		scheme: scheme,
		host:   host,
		port:   uint16(port),
	}
	return uri, nil
}

// Encode converts o into its internal representation.
func (u URI) Encode() *internal.URI {
	return encodeURI(u)
}

func encodeURI(u URI) *internal.URI {
	return &internal.URI{
		Scheme: u.scheme,
		Host:   u.host,
		Port:   uint32(u.port),
	}
}

func DecodeURI(i *internal.URI) URI {
	return decodeURI(i)
}

func decodeURI(i *internal.URI) URI {
	if i == nil {
		return URI{}
	}
	return URI{
		scheme: i.Scheme,
		host:   i.Host,
		port:   uint16(i.Port),
	}
}

func encodeURIs(a []URI) []*internal.URI {
	if len(a) == 0 {
		return nil
	}
	other := make([]*internal.URI, len(a))
	for i := range a {
		other[i] = encodeURI(a[i])
	}
	return other
}

func decodeURIs(a []*internal.URI) []URI {
	if len(a) == 0 {
		return nil
	}
	other := make([]URI, len(a))
	for i := range a {
		other[i] = decodeURI(a[i])
	}
	return other
}

// MarshalJSON marshals URI into a JSON-encoded byte slice.
func (u *URI) MarshalJSON() ([]byte, error) {
	var output struct {
		Scheme string `json:"scheme,omitempty"`
		Host   string `json:"host,omitempty"`
		Port   uint16 `json:"port,omitempty"`
	}
	output.Scheme = u.scheme
	output.Host = u.host
	output.Port = u.port

	return json.Marshal(output)
}

func (u *URI) UnmarshalJSON(b []byte) error {
	var input struct {
		Scheme string `json:"scheme,omitempty"`
		Host   string `json:"host,omitempty"`
		Port   uint16 `json:"port,omitempty"`
	}
	if err := json.Unmarshal(b, &input); err != nil {
		return err
	}
	u.scheme = input.Scheme
	u.host = input.Host
	u.port = input.Port
	return nil
}
