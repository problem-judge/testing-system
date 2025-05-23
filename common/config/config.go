package config

import (
	"os"
	"testing_system/lib/logger"

	"github.com/xorcare/pointer"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port int     `yaml:"Port"`
	Host *string `yaml:"Host,omitempty"` // leave empty for localhost

	Logger *logger.Config `yaml:"Logger,omitempty"`

	Invoker *InvokerConfig `yaml:"Invoker,omitempty"`
	Master  *MasterConfig  `yaml:"Master,omitempty"`
	Storage *StorageConfig `yaml:"Storage,omitempty"`
	// TODO: Add instances here

	DB DBConfig `yaml:"DB"`
	// if instance is set up on server, leave connection empty
	MasterConnection  *Connection `yaml:"MasterConnection,omitempty"`
	StorageConnection *Connection `yaml:"StorageConnection,omitempty"`
}

func ReadConfig(configPath string) *Config {
	content, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	config := new(Config)
	err = yaml.Unmarshal(content, config)
	if err != nil {
		panic(err)
	}

	fillInConfig(config)

	return config
}

func fillInConfig(config *Config) {
	if config.Host == nil {
		config.Host = pointer.String("localhost")
	}

	fillInConnections(config)
	fillInMasterConfig(config.Master)
	fillInStorageConfig(config.Storage)
	FillInInvokerConfig(config.Invoker)
}
