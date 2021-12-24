package common

import (
	"errors"
	"os"
	"path"
	
	"github.com/spf13/viper"
)

// ========== LCM =========

var LCMServiceConfig = LCMService{
	Name: "config",
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

const (
	mariaDBConnString string = "DBConnectionString"
	botToken                 = "BotToken"
	defaultPrefix            = "DefaultPrefix"
	logDir                   = "LogDir"
	retainLogDays            = "RetainLogDays"
	logStdOut                = "LogStdOut"
	botOwner                 = "BotOwner"
)

type Config struct {
	DBConnectionString string
	BotToken           string
	BotOwner           string
	DefaultPrefix      string
	LogStdOut          bool // if false, stdout logger is disabled.
	LogDir             string
	RetainLogDays      uint // if 0, file-logger is disabled
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
		conf.Set(mariaDBConnString, "")
		conf.Set(botToken, "")
		conf.Set(defaultPrefix, "!")
		conf.Set(logDir, path.Join(home, ".hydaelyn", "logs"))
		conf.Set(retainLogDays, 2)
		conf.Set(logStdOut, true)
		conf.Set(botOwner, "@Virepri#2512")
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
	
	out.DBConnectionString = conf.GetString(mariaDBConnString)
	out.BotToken = conf.GetString(botToken)
	out.DefaultPrefix = conf.GetString(defaultPrefix)
	out.LogDir = conf.GetString(logDir)
	out.LogStdOut = conf.GetBool(logStdOut)
	out.RetainLogDays = conf.GetUint(retainLogDays)
	out.BotOwner = conf.GetString(botOwner)
	
	singleCfg = out
	
	return out, nil
}
