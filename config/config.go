package config

import "gopkg.in/ini.v1"

type Config struct {
	Bot struct {
		Token    string `ini:"token_first"`
		TokenTwo string `ini:"token_second"`
	} `ini:"bot"`
	Redis struct {
		Addr     string `ini:"addr"`
		Username string `ini:"username"`
		Password string `ini:"password"`
	} `ini:"redis"`
}

func ReadFromFile[config any](fileName string) (*config, error) {
	var conf config
	err := ini.MapTo(&conf, fileName)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
