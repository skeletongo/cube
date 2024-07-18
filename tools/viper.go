package tools

import (
	"fmt"
	"github.com/spf13/viper"
)

var paths = []string{
	"/etc/cube",
	"$HOME/.cube",
	".",
}

func GetViper(name string) *viper.Viper {
	vp := viper.New()
	vp.SetConfigName(name)
	vp.SetConfigType("json")
	for _, v := range paths {
		vp.AddConfigPath(v)
	}

	err := vp.ReadInConfig()
	if err != nil {
		vp = viper.New()
		vp.SetConfigName(name)
		vp.SetConfigType("yaml")
		for _, v := range paths {
			vp.AddConfigPath(v)
		}
		if err = vp.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
	return vp
}
