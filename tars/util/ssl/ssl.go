package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"reflect"

	"github.com/TarsCloud/TarsGo/tars/util/rogger"
)

var log = rogger.GetLogger("ssl")

func NewServerTlsConfig(ca, cert, key string, verifyClient bool, ciphers string) (tlsConfig *tls.Config, err error) {
	tlsConfig = &tls.Config{}
	if ca != "" {
		certBytes, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		clientCAs := x509.NewCertPool()
		clientCAs.AppendCertsFromPEM(certBytes)
		tlsConfig.ClientCAs = clientCAs
	}
	if ciphers == "" {
		if err = appendKeyPair(tlsConfig, cert, key); err != nil {
			return nil, err
		}
	} else {
		if err = appendKeyPairWithPassword(tlsConfig, cert, key, []byte(ciphers)); err != nil {
			return nil, err
		}
	}
	if verifyClient {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsConfig, nil
}

func NewClientTlsConfig(ca, cert, key string, ciphers string) (tlsConfig *tls.Config, err error) {
	tlsConfig = &tls.Config{}
	insecureSkipVerify := true
	if ca != "" {
		certBytes, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		rootCAs := x509.NewCertPool()
		rootCAs.AppendCertsFromPEM(certBytes)
		tlsConfig.RootCAs = rootCAs
		insecureSkipVerify = false
	}
	tlsConfig.InsecureSkipVerify = insecureSkipVerify
	if ciphers == "" && cert != "" {
		if err = appendKeyPair(tlsConfig, cert, key); err != nil {
			return nil, err
		}
	} else if cert != "" {
		if err = appendKeyPairWithPassword(tlsConfig, cert, key, []byte(ciphers)); err != nil {
			return nil, err
		}
	}
	return tlsConfig, nil
}

// https://github.com/outbrain/orchestrator/blob/master/go/ssl/ssl.go
// appendKeyPair loads the given TLS key pair and appends it to tlsConfig.Certificates.
func appendKeyPair(tlsConfig *tls.Config, certFile string, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	return nil
}

// appendKeyPairWithPassword Read in a keypair where the key is password protected
func appendKeyPairWithPassword(tlsConfig *tls.Config, certFile string, keyFile string, pemPass []byte) error {
	// Certificates aren't usually password protected, but we're kicking the password
	// along just in case.  It won't be used if the file isn't encrypted
	certData, err := ReadPEMData(certFile, pemPass)
	if err != nil {
		return err
	}
	keyData, err := ReadPEMData(keyFile, pemPass)
	if err != nil {
		return err
	}
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return err
	}
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	return nil
}

// ReadPEMData Read a PEM file and ask for a password to decrypt it if needed
func ReadPEMData(pemFile string, pemPass []byte) ([]byte, error) {
	pemData, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return pemData, err
	}

	// We should really just get the pem.Block back here, if there's other
	// junk on the end, warn about it.
	pemBlock, rest := pem.Decode(pemData)
	if len(rest) > 0 {
		log.Info("Didn't parse all of", pemFile)
	}

	if x509.IsEncryptedPEMBlock(pemBlock) {
		// Decrypt and get the ASN.1 DER bytes here
		pemData, err = x509.DecryptPEMBlock(pemBlock, pemPass)
		if err != nil {
			return pemData, err
		} else {
			log.Info("Decrypted", pemFile, "successfully")
		}
		// Shove the decrypted DER bytes into a new pem Block with blank headers
		var newBlock pem.Block
		newBlock.Type = pemBlock.Type
		newBlock.Bytes = pemData
		// This is now like reading in an uncrypted key from a file and stuffing it
		// into a byte stream
		pemData = pem.EncodeToMemory(&newBlock)
	}
	return pemData, nil
}

func FDFromTLSConn(conn *tls.Conn) uintptr {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem().Elem()
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return uintptr(pfdVal.FieldByName("Sysfd").Int())
}
