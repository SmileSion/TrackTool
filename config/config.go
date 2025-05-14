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

type Config struct {
	Server ServerConfig `toml:"server"`
	Redis  RedisConfig  `toml:"redis"`
	Mysql  MysqlConfig  `toml:"mysql"`
	Log    LogConfig    `toml:"log"`
}

var Conf Config

func InitConfig() {
	if _, err := toml.DecodeFile("config/config.toml", &Conf); err != nil {
		panic(err)
	}
}
