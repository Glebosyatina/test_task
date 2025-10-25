package config

import(
	"os"
	"gopkg.in/yaml.v2"
)

const confPath = "config/config.yaml"

type Config struct{
	ApiVersion string `yaml:"api"`
	Server struct{
		Host string	`yaml:"host"`
		Addr string	`yaml:"addr"`
	} `yaml:"server"`
	DbConf struct{
		Host string	`yaml:"host"`
		DbUser string	`yaml:"dbuser"`
		DbName string	`yaml:"dbname"`
		DbPassword string	`yaml:"dbpassword"`
	} `yaml:"db"`
}

func ReadConfig() (*Config){
	var AppConfig Config
	
	f, err := os.Open(confPath)
	if err != nil{
		panic(err)	
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	err = dec.Decode(&AppConfig)
	if err != nil {
		panic(err)
	}

	return &AppConfig
}


