package certgen

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/rigdev/rig/cmd/rig-operator/log"
	"github.com/spf13/cobra"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	flagHosts     = "hosts"
	flagNamespace = "namespace"
)

func create(cmd *cobra.Command, args []string) error {
	k8s, err := newK8s()
	if err != nil {
		return err
	}

	namespace, err := cmd.Flags().GetString(flagNamespace)
	if err != nil {
		return err
	}
	name := args[0]

	log := log.New(false).WithValues("namespace", namespace, "name", name)

	log.Info("checking wether we should create secret...")
	if _, err := k8s.getSecret(cmd.Context(), namespace, name); err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("creating secret")
			hosts, err := cmd.Flags().GetStringSlice(flagHosts)
			if err != nil {
				return err
			}

			certs, err := GenerateCerts(hosts)
			if err != nil {
				return err
			}

			if err := k8s.createSecret(cmd.Context(), namespace, name, certs); err != nil {
				return err
			}
			log.Info("secret created")
			// create
		} else {
			return err
		}
	} else {
		log.Info("secret already exists")
	}

	return nil
}

func serialNumer() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("could not generate serial number: %w", err)
	}
	return serialNumber, nil
}

func encodeKey(key *ecdsa.PrivateKey) ([]byte, error) {
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal ECDSA private key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b}), nil
}

func encodeCert(derBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
}

type Certs struct {
	CA   []byte
	Cert []byte
	Key  []byte
}

func GenerateCerts(hosts []string) (*Certs, error) {
	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("could not generate root key: %w", err)
	}

	sn, err := serialNumer()
	if err != nil {
		return nil, err
	}

	notBefore := time.Now().Add(time.Minute * -5)
	notAfter := notBefore.Add(time.Hour * 24 * 365 * 100)

	rootTemplate := x509.Certificate{
		SerialNumber:          sn,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		Subject: pkix.Name{
			Organization: []string{"Rig.dev"},
			CommonName:   "rig-operator CA",
		},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		return nil, fmt.Errorf("could not create ca certificate: %w", err)
	}

	ca := encodeCert(derBytes)

	keyRaw, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("could not generate cert key: %w", err)
	}

	key, err := encodeKey(keyRaw)
	if err != nil {
		return nil, err
	}

	sn, err = serialNumer()
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: sn,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		Subject: pkix.Name{
			Organization: []string{"Rig.dev"},
			CommonName:   "rig-operator webhooks",
		},
	}

	for _, host := range hosts {
		if pip := net.ParseIP(host); pip != nil {
			template.IPAddresses = append(template.IPAddresses, pip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}

	derBytes, err = x509.CreateCertificate(rand.Reader, &template, &rootTemplate, &keyRaw.PublicKey, rootKey)
	if err != nil {
		return nil, fmt.Errorf("could not create certificate: %w", err)
	}

	return &Certs{
		CA:   ca,
		Key:  key,
		Cert: encodeCert(derBytes),
	}, nil
}
