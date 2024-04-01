package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// loads configs in this order:
// default.yml file
// then any config_#{env}.yml file
// then environment variables
func LoadConfig(path string, config interface{}) error {

	//where to read from
	viper.AddConfigPath(path)

	// 1st load from default.yml
	viper.SetConfigName("defaults")

	//search in defined path we added earlier and parse configs
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	// load from config_{env}.yml
	env := viper.GetString("Env")
	if env != "" {
		filename := "config_" + env
		viper.SetConfigName(filename)
	}

	// merge into the defaults we parsed earlier
	if err := viper.MergeInConfig(); err != nil {
		fmt.Println(err)
	}

	//unmarshal into config
	err := viper.Unmarshal(config)

	if err != nil {
		return err
	}
	viper.AutomaticEnv()

	return nil
}
