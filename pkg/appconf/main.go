package appconf

import co "github.com/msolimans/wikimovie/pkg/config"

func LoadConfig(path string, config interface{}) error {
	return co.LoadConfig(path, config)
}
