package config

import "github.com/spf13/viper"

func init() {
	//this is a must have otherwise overriding via env vars will not work
	_ = viper.BindEnv("Env", "ENV")
}
