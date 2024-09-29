package main

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	DbName string
}

func GetConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Panicf("Error reading config file: %s", err)
	}

	return Config{
		DbName: viper.Get("dbName").(string),
	}
}
