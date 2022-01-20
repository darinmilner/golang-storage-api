package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

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
	RemoteFSName     string `yaml:"remotefsname"`
	MaxUploadSize    int64
	AllowedMimeTypes []string
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

	var maxUploadSize int64
	max, err := strconv.Atoi(os.Getenv("MAX_UPLOAD_SIZE"))
	if err != nil {
		maxUploadSize = 10 << 20
	} else {
		maxUploadSize = int64(max)
	}

	//file uploads
	exploded := strings.Split(os.Getenv("ALLOWED_FILE_TYPES"), ",")
	var mimeTypes []string
	for _, m := range exploded {
		mimeTypes = append(mimeTypes, m)
	}

	conf := &Config{
		MaxUploadSize:    int64(maxUploadSize),
		AllowedMimeTypes: mimeTypes,
	}

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
