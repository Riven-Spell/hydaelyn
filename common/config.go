package common

import (
	"errors"
	"os"
	"path"

	"github.com/spf13/viper"
)

// ========== LCM =========

var LCMServiceConfig = LCMService{
	Name: LCMServiceNameConfig,
	Startup: func() error {
		_, err := ReadConfig()
		return err
	},
	GetSvc: func() interface{} {
		cfg, _ := ReadConfig() // should work by now.
		return cfg
	},
}

// ========== CORE ==========

var ConfigTarget string
var singleViperCfg *viper.Viper
var singleCfg *Config

type Config struct {
	Log     ConfigLog
	DB      ConfigDB
	Discord ConfigDiscord
}

func defaultConfig(homeDir string) Config {
	return Config{
		Log: ConfigLog{
			LogDir:        path.Join(homeDir, ".hydaelyn", "logs"),
			LogStdOut:     true,
			RetainLogDays: 7,
		},
		DB: ConfigDB{
			DBConnectionString: "mysql://hydaelyn:password@tcp(127.0.0.1)/",
			DBName:             "hydaelyn",
		},
	}
}

type ConfigDB struct {
	DBConnectionString string
	DBName             string
}

type ConfigDiscord struct {
	BotApplicationID string
	BotToken         string
	BotOwner         string
}

type ConfigLog struct {
	LogStdOut     bool // if false, stdout logger is disabled.
	LogDir        string
	RetainLogDays uint // if 0, file-logger is disabled
}

// todo: reload config?

func getViperConfig() (*viper.Viper, error) {
	if singleViperCfg != nil {
		return singleViperCfg, nil
	}

	conf := viper.New()

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	_ = os.MkdirAll(path.Join(home, ".hydaelyn"), 0700)

	conf.AddConfigPath(path.Join(home, ".hydaelyn"))
	conf.SetConfigName("hydaelyn")
	conf.SetConfigType("yaml")
	err = conf.ReadInConfig() // try to read the config

	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Create the config
		conf.Set("config", defaultConfig(home))
		err = conf.SafeWriteConfig() // try to write the config
		if err == nil {
			err = errors.New("configuration required in " + path.Join(home, ".hydaelyn", "hydaelyn.yaml"))
		}

		return conf, err
	}

	if err != nil {
		singleViperCfg = conf
	}

	return conf, err
}

func ReadConfig() (*Config, error) {
	if singleCfg != nil {
		return singleCfg, nil
	}

	conf, err := getViperConfig()

	if err != nil {
		return nil, err
	}

	out := &Config{}

	err = conf.UnmarshalKey("config", out)
	if err != nil {
		return nil, err
	}

	singleCfg = out

	return out, nil
}
