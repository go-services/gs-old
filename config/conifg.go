package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"gs/fs"

	"github.com/asaskevich/govalidator"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var log = logrus.WithFields(logrus.Fields{"package": "config"})

type AddressConfig struct {
	Url  string `json:"url"`
	Port int    `json:"port"`
}

type ServiceConfig struct {
	Name string        `json:"name" valid:"alphanum,required"`
	Http AddressConfig `json:"http"`
	Grpc AddressConfig `json:"grpc"`
}

type GSConfig struct {
	Module   string          `json:"module"`
	Services []ServiceConfig `json:"services"`
}

var cfg *GSConfig

func Read() (*GSConfig, error) {
	if cfg != nil {
		return cfg, nil
	}

	configFile := viper.GetString(GSConfigFileName)
	log.Debugf("Reading config `%s`", configFile)

	jsonData, err := fs.ReadFile(configFile)
	if err != nil {
		return nil, errors.New("not in a GS project, you need to be in a GS project to run this command")
	}

	config, err := decode(jsonData)

	if err != nil {
		return nil, errors.New("GS config malformed: " + err.Error())
	}

	if err := validate(config); err != nil {
		return nil, errors.New("GS config not valid: " + err.Error())
	}
	cfg = &config
	return &config, nil
}
func ReRead() (*GSConfig, error) {
	cfg = nil
	return Read()
}

func Update(config GSConfig) error {
	encoded, err := encode(config)
	if err != nil {
		return err
	}
	return fs.WriteFile(viper.GetString(GSConfigFileName), encoded)
}

func decode(jsonData string) (config GSConfig, err error) {
	err = json.NewDecoder(bytes.NewBufferString(jsonData)).Decode(&config)
	return
}
func encode(config GSConfig) (string, error) {
	encoded, err := json.MarshalIndent(config, "", "\t")
	return string(encoded), err
}

func validate(config GSConfig) error {
	_, err := govalidator.ValidateStruct(config)
	return err
}
