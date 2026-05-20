package config

import "github.com/spf13/viper"

func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	return viper.ReadInConfig()
}
