package config

import (
	"github.com/BurntSushi/toml"
	"log"
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
	MaxSize    int    `toml:"max_size"`
	MaxBackups int    `toml:"max_backups"`
	MaxAge     int    `toml:"max_age"`
	Compress   bool   `toml:"compress"`
}

type AESConfig  struct {
	Key string `toml:"key"` // AES密钥
}

type Capability  struct {
	Chan int `toml:"chan"`
	Worker int `toml:"worker"`
	Batch int `toml:"batch"`
}

type Config struct {
	Server ServerConfig `toml:"server"`
	Redis  RedisConfig  `toml:"redis"`
	Mysql  MysqlConfig  `toml:"mysql"`
	Log    LogConfig    `toml:"log"`
	AES    AESConfig    `toml:"aes"`
	Cap    Capability    `toml:"cap"`
}

var Conf Config

func InitConfig() {
	if _, err := toml.DecodeFile("config/config.toml", &Conf); err != nil {
		panic(err)
	}
	// ✅ 打印初始化日志
	log.Println("配置文件加载成功：")
	log.Printf("  Server Port: %d\n", Conf.Server.Port)
	log.Printf("  Redis Addr: %s, DB: %d\n", Conf.Redis.Addr, Conf.Redis.DB)
	log.Printf("  MySQL DSN: %s\n", Conf.Mysql.DSN)
	log.Printf("  Log File: %s\n", Conf.Log.Filepath)
	log.Printf("  AES Key: %s\n", Conf.AES.Key)
}
