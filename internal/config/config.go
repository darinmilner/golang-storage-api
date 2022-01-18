package config

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AppName string `yaml:"appname"`
	HTTP    struct {
		Port string
	} `yaml:"http"`
	Logger struct {
		Level string `yaml:"level"`
	}
	Minio struct {
		Endpoint string
		Key      string
		Secret   string
		UseSSL   bool
		Region   string
		Bucket   string
	} `yaml:"minio"`
	AWS struct {
		Endpoint string
		Key      string
		Secret   string
		Region   string
		Bucket   string
	} `yaml:"aws"`
	RemoteFSName string `yaml:"remotefsname"`
}

func NewConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	confPath := os.Getenv("STORAGE_CONFIG")
	if confPath == "" {

		return nil, errors.New("STORAGE_CONFIG env variable is not set")
	}

	conf := &Config{}
	file, err := os.Open(confPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err = yaml.NewDecoder(file).Decode(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
