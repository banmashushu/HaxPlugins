package config

import (
	"github.com/spf13/viper"
)

// AppConfig 应用配置
type AppConfig struct {
	Hotkey      string `mapstructure:"hotkey"`
	DataSource  string `mapstructure:"data_source"`
	AutoUpdate  bool   `mapstructure:"auto_update"`
	UpdateInterval int `mapstructure:"update_interval"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// Config 总配置
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.haxplugins")

	// 默认值
	viper.SetDefault("app.hotkey", "f1")
	viper.SetDefault("app.data_source", "opgg")
	viper.SetDefault("app.auto_update", true)
	viper.SetDefault("app.update_interval", 360)
	viper.SetDefault("database.path", "./data/haxplugins.db")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
