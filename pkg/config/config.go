package config

import (
    "encoding/base64"
    "gopkg.in/yaml.v3"
    "io/ioutil"
    "os"
)

const (
    CONFIG_PATH = "config.yml"
)

type Config struct {
    Salt []byte
    Secret []byte
    LoginToKey map[string][]byte
    LoggerLevel int
    Port int
    JaegerAddr string
    TracerName string
    SentryDsn string
}

type yamlConfig struct {
    LoggerLevel int `yaml:"logger_level"`
    Port int `yaml:"port"`
    Users map[string]string `yaml:"users"`
    JaegerAddr string `yaml:"jaeger_name"`
    TracerName string `yaml:"tracer_name"`
    SentryDsn string `yaml:"sentry_dsn"`
}

func decodeKeys(base64EncodedKeys map[string]string) (map[string][]byte, error) {
    result := make(map[string][]byte)

    for login, base64EncodedKey := range base64EncodedKeys {
        key, err := base64.StdEncoding.DecodeString(base64EncodedKey)
        if err != nil {
            return result, err
        }
        result[login] = key
    }

    return result, nil
}

func Load() (*Config, error) {
    config_bytes, err := ioutil.ReadFile(CONFIG_PATH)
    if err != nil {
        return nil, err
    }

    var yamlCfg yamlConfig
    err = yaml.Unmarshal(config_bytes, &yamlCfg)
    if err != nil {
        return nil, err
    }

    loginToKey, err := decodeKeys(yamlCfg.Users)
    if err != nil {
        return nil, err
    }

    cfg := &Config {
        Salt: []byte(os.Getenv("SALT")),
        Secret: []byte(os.Getenv("SECRET")),
        LoginToKey: loginToKey,
        LoggerLevel: yamlCfg.LoggerLevel,
        Port: yamlCfg.Port,
        JaegerAddr: yamlCfg.JaegerAddr,
        TracerName: yamlCfg.TracerName,
        SentryDsn: yamlCfg.SentryDsn,
    }
    return cfg, nil
}
