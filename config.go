package main

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Email    string
	Password string
}

func GetConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Panicf("Error reading config file: %s", err)
	}

	return Config{Email: viper.Get("email").(string), Password: viper.Get("password").(string)}
}
