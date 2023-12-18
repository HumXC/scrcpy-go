package tls

import (
	"crypto/tls"
	_ "embed"
)

//go:embed cert.pem
var Cret []byte

//go:embed csr.pem
var Csr []byte

//go:embed key.pem
var Key []byte

func Config() (*tls.Config, error) {
	cert, err := tls.X509KeyPair(Cret, Key)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		NextProtos:         []string{"h3-23"}, // 支持的HTTP/3版本
	}

	return tlsConfig, nil
}
