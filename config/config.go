package config

import (
	"github.com/BurntSushi/toml"
)

type ServerConfig struct {
	Port int `toml:"port"`
}

type RedisConfig struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type MysqlConfig struct {
	DSN string `toml:"dsn"`
}

type LogConfig struct {
	Filepath string `toml:"filepath"`
}

type AESConfig  struct {
	Key string `toml:"key"` // SM4密钥
}

type Config struct {
	Server ServerConfig `toml:"server"`
	Redis  RedisConfig  `toml:"redis"`
	Mysql  MysqlConfig  `toml:"mysql"`
	Log    LogConfig    `toml:"log"`
	AES    AESConfig    `toml:"aes"`
}

var Conf Config

func InitConfig() {
	if _, err := toml.DecodeFile("config/config.toml", &Conf); err != nil {
		panic(err)
	}
}
