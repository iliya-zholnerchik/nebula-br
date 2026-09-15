package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	log "github.com/sirupsen/logrus"

	"github.com/vesoft-inc/nebula-br/pkg/utils"
	"github.com/vesoft-inc/nebula-go/v3/nebula"
	"github.com/vesoft-inc/nebula-go/v3/nebula/meta"
)

const (
	defaultTimeout = 120 * time.Second
)

func newMetaTLSConfig(cfg MetaSSLConfig) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(cfg.CAPath)
	if err != nil {
		return nil, fmt.Errorf("read CA certificate failed: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("append CA certificate failed")
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load client certificate/key failed: %w", err)
	}

	return &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func connect(metaAddr *nebula.HostAddr, sslConfig MetaSSLConfig) (*meta.MetaServiceClient, error) {
	log.WithField("meta address", utils.StringifyAddr(metaAddr)).
		Info("Try to connect meta service.")

	var transport thrift.Transport
	var err error

	if sslConfig.Enable {
		tlsConfig, err := newMetaTLSConfig(sslConfig)
		if err != nil {
			return nil, fmt.Errorf("create TLS config failed: %w", err)
		}

		sock, err := thrift.NewSSLSocketTimeout(
			utils.StringifyAddr(metaAddr),
			tlsConfig,
			defaultTimeout,
		)
		if err != nil {
			return nil, fmt.Errorf("open SSL socket failed: %w", err)
		}

		bufferedTranFactory := thrift.NewBufferedTransportFactory(128 << 10)
		transport = thrift.NewFramedTransport(
			bufferedTranFactory.GetTransport(sock),
		)
	} else {
		timeoutOption := thrift.SocketTimeout(defaultTimeout)
		addressOption := thrift.SocketAddr(utils.StringifyAddr(metaAddr))

		sock, err := thrift.NewSocket(timeoutOption, addressOption)
		if err != nil {
			return nil, fmt.Errorf("open socket failed: %w", err)
		}

		bufferedTranFactory := thrift.NewBufferedTransportFactory(128 << 10)
		transport = thrift.NewFramedTransport(
			bufferedTranFactory.GetTransport(sock),
		)
	}

	pf := thrift.NewBinaryProtocolFactoryDefault()
	client := meta.NewMetaServiceClientFactory(transport, pf)

	if err := client.CC.Open(); err != nil {
		return nil, fmt.Errorf("open meta failed %w", err)
	}

	req := newVerifyClientVersionReq()
	resp, err := client.VerifyClientVersion(req)
	if err != nil || resp.Code != nebula.ErrorCode_SUCCEEDED {
		log.WithError(err).
			WithField("addr", metaAddr).
			Error("Incompatible version between client and server.")
		client.Close()
		return nil, err
	}

	log.WithField("meta address", utils.StringifyAddr(metaAddr)).
		Info("Connect meta server successfully.")

	return client, nil
}

func newVerifyClientVersionReq() *meta.VerifyClientVersionReq {
	return &meta.VerifyClientVersionReq{
		ClientVersion: []byte(nebula.Version),
		Host:          nebula.NewHostAddr(),
	}
}
