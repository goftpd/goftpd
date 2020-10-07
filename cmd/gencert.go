package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	var host, tlsCertFile, tlsKeyFile string

	var gencertCmd = &cobra.Command{
		Use:   "gencert",
		Short: "Create a TLS certificate for goftpd.",
		RunE: func(cmd *cobra.Command, args []string) error {

			priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				return errors.WithMessage(err, "Failed to generate private key")
			}

			keyUsage := x509.KeyUsageDigitalSignature

			notBefore := time.Now()
			notAfter := notBefore.Add(365 * 24 * time.Hour)

			serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

			serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
			if err != nil {
				return errors.WithMessage(err, "Failed to generate serial number")
			}

			template := x509.Certificate{
				SerialNumber: serialNumber,
				Subject: pkix.Name{
					Organization: []string{"Acme Co"},
				},
				NotBefore:             notBefore,
				NotAfter:              notAfter,
				KeyUsage:              keyUsage,
				ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
				BasicConstraintsValid: true,
			}

			hosts := strings.Split(host, ",")

			for _, h := range hosts {
				if ip := net.ParseIP(h); ip != nil {
					template.IPAddresses = append(template.IPAddresses, ip)
				} else {
					template.DNSNames = append(template.DNSNames, h)
				}
			}

			derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
			if err != nil {
				return errors.WithMessage(err, "Failed to create certificate")
			}

			certOut, err := os.Create(tlsCertFile)
			if err != nil {
				return errors.WithMessagef(err, "Failed to open %s for writing", tlsCertFile)
			}

			if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
				return errors.WithMessagef(err, "Failed to write data to %s", tlsCertFile)
			}

			if err := certOut.Close(); err != nil {
				return errors.WithMessagef(err, "Error closing %s", tlsCertFile)
			}

			keyOut, err := os.OpenFile(tlsKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				return errors.WithMessagef(err, "Failed to open %s for writing", tlsKeyFile)
			}

			privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
			if err != nil {
				return errors.WithMessage(err, "Unable to marshal private key")
			}

			if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
				return errors.WithMessage(err, "Failed to write data to key.pem")
			}

			if err := keyOut.Close(); err != nil {
				return errors.WithMessage(err, "Error closing key.pem")
			}

			return nil
		},
	}

	gencertCmd.Flags().StringVarP(&host, "host", "", "", "Comma-separated hostnames and IPs to generate a certificate for")
	gencertCmd.Flags().StringVarP(&tlsCertFile, "tlsCert", "c", "site/config/cert.pem", "Location to save TLS Certificate file")
	gencertCmd.Flags().StringVarP(&tlsKeyFile, "tlsKey", "k", "site/config/key.pem", "Location to save TLS Key file")

	gencertCmd.MarkFlagRequired("host")

	rootCmd.AddCommand(gencertCmd)
}
