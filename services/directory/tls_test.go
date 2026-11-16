// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"crypto/tls"
	ber "github.com/go-asn1-ber/asn1-ber"
	"net"
	"net/http"
	"net/http/httptest"
	"svnova/models"
	"testing"
	"time"
)

func TestLDAPTLSVerificationIsPerSource(t *testing.T) {
	certificateServer := httptest.NewTLSServer(http.NotFoundHandler())
	certificate := certificateServer.TLS.Certificates[0]
	certificateServer.Close()
	for _, scheme := range []string{"ldaps", "ldap"} {
		for _, skip := range []bool{false, true} {
			t.Run(scheme+map[bool]string{false: "-verify", true: "-skip"}[skip], func(t *testing.T) {
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				defer listener.Close()
				done := make(chan struct{})
				go func() {
					defer close(done)
					conn, err := listener.Accept()
					if err != nil {
						return
					}
					defer conn.Close()
					conn.SetDeadline(time.Now().Add(3 * time.Second))
					if scheme == "ldap" {
						packet, err := ber.ReadPacket(conn)
						if err != nil || len(packet.Children) != 2 {
							return
						}
						response := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "")
						response.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, packet.Children[0].Value, ""))
						extended := ber.Encode(ber.ClassApplication, ber.TypeConstructed, 24, nil, "")
						extended.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, 0, ""))
						extended.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", ""))
						extended.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", ""))
						response.AppendChild(extended)
						if _, err = conn.Write(response.Bytes()); err != nil {
							return
						}
					}
					secure := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{certificate}, MinVersion: tls.VersionTLS12})
					if secure.Handshake() == nil {
						var buffer [1]byte
						_, _ = secure.Read(buffer[:])
					}
				}()
				source := &models.LDAPConfig{URL: scheme + "://" + listener.Addr().String(), SkipVerify: skip, BaseDN: "dc=example", Filter: "(objectClass=person)", UsernameAttribute: "uid", KeyAttribute: "entryUUID"}
				connection, err := (&Directory{}).connect(source)
				if connection != nil {
					connection.Close()
				}
				if (err == nil) != skip {
					t.Fatalf("TLS verification skip=%v result=%v", skip, err)
				}
				listener.Close()
				<-done
			})
		}
	}
}
