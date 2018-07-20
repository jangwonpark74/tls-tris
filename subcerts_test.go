// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"testing"
	"time"
)

// These test keys were generated with the following program, available in the
// crypto/tls directory:
//
//		go run generate_cert.go -ecdsa-curve P256 -host 127.0.0.1 -dc
//
// To get a certificate without the DelegationUsage extension, remove the `-dc`
// parameter.
var delegatorCertPEM = `-----BEGIN CERTIFICATE-----
MIIBdzCCAR2gAwIBAgIQLVIvEpo0/0TzRja4ImvB1TAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE4MDcwMzE2NTE1M1oXDTE5MDcwMzE2NTE1M1ow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABOhB
U6adaAgliLaFc1PAo9HBO4Wish1G4df3IK5EXLy+ooYfmkfzT1FxqbNLZufNYzve
25fmpal/1VJAjpVyKq2jVTBTMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggr
BgEFBQcDATAMBgNVHRMBAf8EAjAAMA8GA1UdEQQIMAaHBH8AAAEwDQYJKwYBBAGC
2kssBAAwCgYIKoZIzj0EAwIDSAAwRQIhAPNwRk6cygm6zO5rjOzohKYWS+1KuWCM
OetDIvU4mdyoAiAGN97y3GJccYn9ZOJS4UOqhr9oO8PuZMLgdq4OrMRiiA==
-----END CERTIFICATE-----
`

var delegatorKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIJDVlo+sJolMcNjMkfCGDUjMJcE4UgclcXGCrOtbJAi2oAoGCCqGSM49
AwEHoUQDQgAE6EFTpp1oCCWItoVzU8Cj0cE7haKyHUbh1/cgrkRcvL6ihh+aR/NP
UXGps0tm581jO97bl+alqX/VUkCOlXIqrQ==
-----END EC PRIVATE KEY-----
`

var nonDelegatorCertPEM = `-----BEGIN CERTIFICATE-----
MIIBaTCCAQ6gAwIBAgIQSUo+9uaip3qCW+1EPeHZgDAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE4MDYxMjIzNDAyNloXDTE5MDYxMjIzNDAyNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABLf7
fiznPVdc3V5mM3ymswU2/IoJaq/deA6dgdj50ozdYyRiAPjxzcz9zRsZw1apTF/h
yNfiLhV4EE1VrwXcT5OjRjBEMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggr
BgEFBQcDATAMBgNVHRMBAf8EAjAAMA8GA1UdEQQIMAaHBH8AAAEwCgYIKoZIzj0E
AwIDSQAwRgIhANXG0zmrVtQBK0TNZZoEGMOtSwxmiZzXNe+IjdpxO3TiAiEA5VYx
0CWJq5zqpVXbJMeKVMASo2nrXZoA6NhJvFQ97hw=
-----END CERTIFICATE-----
`

var nonDelegatorKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMw9DiOfGI1E/XZrrW2huZSjYi0EKwvVjAe+dYtyFsSloAoGCCqGSM49
AwEHoUQDQgAEt/t+LOc9V1zdXmYzfKazBTb8iglqr914Dp2B2PnSjN1jJGIA+PHN
zP3NGxnDVqlMX+HI1+IuFXgQTVWvBdxPkw==
-----END EC PRIVATE KEY-----
`

var dcTestDCsPEM = `-----BEGIN DC TEST DATA-----
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
HJbn52hmB027JEIMPWsxKxPkr7udk7Q=
-----END DC TEST DATA-----
`

type dcTestDC struct {
	Name       string
	Version    int
	Scheme     int
	DC         []byte
	PrivateKey []byte
}

var dcTestDCs []dcTestDC
var dcTestConfig *Config
var dcTestDelegationCert Certificate
var dcTestCert Certificate
var dcTestNow time.Time

func init() {

	// Parse the PEM-encoded DER block containing the test DCs.
	block, _ := pem.Decode([]byte(dcTestDCsPEM))
	if block == nil {
		panic("failed to decode DC tests PEM block")
	}

	// Parse the DER-encoded test DCs.
	_, err := asn1.Unmarshal(block.Bytes, &dcTestDCs)
	if err != nil {
		panic("failed to unmarshal DC test ASN.1 data")
	}

	// Use a static time for testing. This is the point at which the test DCs
	// were generated.
	dcTestNow = time.Date(2018, 07, 03, 18, 0, 0, 234234, time.UTC)

	// The base configuration for the client and server.
	dcTestConfig = &Config{
		Time: func() time.Time {
			return dcTestNow
		},
		Rand:         zeroSource{},
		Certificates: nil,
		MinVersion:   VersionTLS10,
		MaxVersion:   VersionTLS13Draft22,
		CipherSuites: allCipherSuites(),
	}

	// The delegation certificate.
	dcTestDelegationCert, err = X509KeyPair([]byte(delegatorCertPEM), []byte(delegatorKeyPEM))
	if err != nil {
		panic(err)
	}
	dcTestDelegationCert.Leaf, err = x509.ParseCertificate(dcTestDelegationCert.Certificate[0])
	if err != nil {
		panic(err)
	}

	// A certificate without the the DelegationUsage extension for X.509.
	dcTestCert, err = X509KeyPair([]byte(nonDelegatorCertPEM), []byte(nonDelegatorKeyPEM))
	if err != nil {
		panic(err)
	}
	dcTestCert.Leaf, err = x509.ParseCertificate(dcTestCert.Certificate[0])
	if err != nil {
		panic(err)
	}

	// Make these roots of these certificates the client's trusted CAs.
	dcTestConfig.RootCAs = x509.NewCertPool()

	raw := dcTestDelegationCert.Certificate[len(dcTestDelegationCert.Certificate)-1]
	root, err := x509.ParseCertificate(raw)
	if err != nil {
		panic(err)
	}
	dcTestConfig.RootCAs.AddCert(root)

	raw = dcTestCert.Certificate[len(dcTestCert.Certificate)-1]
	root, err = x509.ParseCertificate(raw)
	if err != nil {
		panic(err)
	}
	dcTestConfig.RootCAs.AddCert(root)
}

// Tests the handshake and one round of application data. Returns true if the
// connection used a DC.
func testConnWithDC(t *testing.T, clientMsg, serverMsg string, clientConfig, serverConfig *Config) (bool, error) {
	ln := newLocalListener(t)
	defer ln.Close()

	// Listen for and serve a single client connection.
	srvCh := make(chan *Conn, 1)
	var serr error
	go func() {
		sconn, err := ln.Accept()
		if err != nil {
			serr = err
			srvCh <- nil
			return
		}
		srv := Server(sconn, serverConfig)
		if err := srv.Handshake(); err != nil {
			serr = fmt.Errorf("handshake: %v", err)
			srvCh <- nil
			return
		}
		srvCh <- srv
	}()

	// Dial the server.
	cli, err := Dial("tcp", ln.Addr().String(), clientConfig)
	if err != nil {
		return false, err
	}
	defer cli.Close()

	srv := <-srvCh
	if srv == nil {
		return false, serr
	}

	bufLen := len(clientMsg)
	if len(serverMsg) > len(clientMsg) {
		bufLen = len(serverMsg)
	}
	buf := make([]byte, bufLen)

	// Client sends a message to the server, which is read by the server.
	cli.Write([]byte(clientMsg))
	n, err := srv.Read(buf)
	if n != len(clientMsg) || string(buf[:n]) != clientMsg {
		return false, fmt.Errorf("Server read = %d, buf= %q; want %d, %s", n, buf, len(clientMsg), clientMsg)
	}

	// Server reads a message from the client, which is read by the client.
	srv.Write([]byte(serverMsg))
	n, err = cli.Read(buf)
	if n != len(serverMsg) || err != nil || string(buf[:n]) != serverMsg {
		return false, fmt.Errorf("Client read = %d, %v, data %q; want %d, nil, %s", n, err, buf, len(serverMsg), serverMsg)
	}

	// Return true if the client's conn.dc structure was instantiated.
	return (cli.verifiedDc != nil), nil
}

// Checks that the client suppports a version >= 1.2 and accepts delegated
// credentials. If so, it returns the delegation certificate; otherwise it
// returns a plain certificate.
func testServerGetCertificate(ch *ClientHelloInfo) (*Certificate, error) {
	versOk := false
	for _, vers := range ch.SupportedVersions {
		versOk = versOk || (vers >= uint16(VersionTLS12))
	}

	if versOk && ch.AcceptsDelegatedCredential {
		return &dcTestDelegationCert, nil
	}
	return &dcTestCert, nil
}

// Various test cases for handshakes involving DCs.
var dcTests = []struct {
	clientDC         bool
	serverDC         bool
	clientSkipVerify bool
	clientMaxVers    uint16
	serverMaxVers    uint16
	nowOffset        time.Duration
	dcTestName       string
	expectSuccess    bool
	expectDC         bool
	name             string
}{
	{true, true, false, VersionTLS12, VersionTLS12, 0, "tls12", true, true, "tls12"},
	{true, true, false, VersionTLS13, VersionTLS13, 0, "tls13", true, true, "tls13"},
	{true, true, false, VersionTLS12, VersionTLS12, 0, "malformed12", false, false, "tls12, malformed dc"},
	{true, true, false, VersionTLS13, VersionTLS13, 0, "malformed13", false, false, "tls13, malformed dc"},
	{true, true, true, VersionTLS12, VersionTLS12, 0, "invalid", true, true, "tls12, invalid dc, skip verify"},
	{true, true, true, VersionTLS13, VersionTLS13, 0, "invalid", true, true, "tls13, invalid dc, skip verify"},
	{false, true, false, VersionTLS12, VersionTLS12, 0, "tls12", true, false, "client no dc"},
	{true, false, false, VersionTLS12, VersionTLS12, 0, "tls12", true, false, "server no dc"},
	{true, true, false, VersionTLS11, VersionTLS12, 0, "tls12", true, false, "client old"},
	{true, true, false, VersionTLS12, VersionTLS11, 0, "tls12", true, false, "server old"},
	{true, true, false, VersionTLS13, VersionTLS13, dcMaxTTL, "tls13", false, false, "expired dc"},
}

// Tests the handshake with the delegated credential extension for each test
// case in dcTests.
func TestDCHandshake(t *testing.T) {
	serverMsg := "hello"
	clientMsg := "world"

	clientConfig := dcTestConfig.Clone()
	serverConfig := dcTestConfig.Clone()
	serverConfig.GetCertificate = testServerGetCertificate

	for i, test := range dcTests {
		clientConfig.MaxVersion = test.clientMaxVers
		serverConfig.MaxVersion = test.serverMaxVers
		clientConfig.InsecureSkipVerify = test.clientSkipVerify
		clientConfig.AcceptDelegatedCredential = test.clientDC
		clientConfig.Time = func() time.Time {
			return dcTestNow.Add(time.Duration(test.nowOffset))
		}

		if test.serverDC {
			serverConfig.GetDelegatedCredential = func(
				ch *ClientHelloInfo, vers uint16) (*DelegatedCredential, crypto.PrivateKey, error) {
				for _, t := range dcTestDCs {
					if t.Name == test.dcTestName {
						dc, err := UnmarshalDelegatedCredential(t.DC)
						if err != nil {
							return nil, nil, err
						}
						sk, err := x509.ParseECPrivateKey(t.PrivateKey)
						if err != nil {
							return nil, nil, err
						}
						return dc, sk, nil
					}
				}
				return nil, nil, fmt.Errorf("Test DC with name '%s' not found", test.dcTestName)
			}
		} else {
			serverConfig.GetDelegatedCredential = nil
		}

		usedDC, err := testConnWithDC(t, clientMsg, serverMsg, clientConfig, serverConfig)
		if err != nil && test.expectSuccess {
			t.Errorf("test #%d (%s) fails: %s", i+1, test.name, err)
		} else if err == nil && !test.expectSuccess {
			t.Errorf("test #%d (%s) succeeds; expected failure", i+1, test.name)
		}

		if usedDC != test.expectDC {
			t.Errorf("test #%d (%s) usedDC = %v; expected %v", i+1, test.name, usedDC, test.expectDC)
		}
	}
}
