package config

import (
	"github.com/spf13/pflag"

	"github.com/vesoft-inc/nebula-br/pkg/storage"
)

const (
	FlagStorage  = "storage"
	FlagMetaAddr = "meta"
	FlagSpaces   = "spaces"

	FlagLogPath  = "log"
	FlagLogDebug = "debug"

	FlagMetaSSL    = "enable_meta_ssl"
	FlagMetaCAPath = "ca_path"
	FlagMetaCert   = "cert_path"
	FlagMetaKey    = "key_path"

	flagBackupName = "name"
)

func AddCommonFlags(flags *pflag.FlagSet) {
	flags.String(FlagLogPath, "br.log", "Specify br detail log path")
	flags.Bool(FlagLogDebug, false, "Output log in debug level or not")

	flags.Bool(FlagMetaSSL, false, "Enable SSL connection to meta server")
	flags.String(FlagMetaCAPath, "", "Path to CA certificate for meta SSL")
	flags.String(FlagMetaCert, "", "Path to client certificate for meta SSL")
	flags.String(FlagMetaKey, "", "Path to client private key for meta SSL")

	storage.AddFlags(flags)
}

type MetaSSLConfig struct {
	Enable   bool
	CAPath   string
	CertPath string
	KeyPath  string
}

func (c *MetaSSLConfig) ParseFlags(flags *pflag.FlagSet) error {
	var err error

	c.Enable, err = flags.GetBool(FlagMetaSSL)
	if err != nil {
		return err
	}

	c.CAPath, err = flags.GetString(FlagMetaCAPath)
	if err != nil {
		return err
	}

	c.CertPath, err = flags.GetString(FlagMetaCert)
	if err != nil {
		return err
	}

	c.KeyPath, err = flags.GetString(FlagMetaKey)
	if err != nil {
		return err
	}

	return nil
}

type NodeInfo struct {
	Addrs   string
	RootDir string
	DataDir []string
}
