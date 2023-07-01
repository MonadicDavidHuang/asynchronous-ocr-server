package config

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	PROFILE = "PROFILE"

	PROFILE_LOCAL = "local"
)

var (
	config Config
	once   sync.Once
)

type Config struct {
	Host     string `mapstructure:"db_host"`
	Port     string `mapstructure:"db_port"`
	Database string `mapstructure:"db_database"`
	User     string `mapstructure:"db_user"`
	Pass     string `mapstructure:"db_pass"`
}

func Init(isForTest bool) Config {
	once.Do(func() {
		config = constructConfig(isForTest)
	})

	return config
}

func constructConfig(isForTest bool) Config {
	profile := mustGetProperProfile()

	if isForTest && profile != PROFILE_LOCAL {
		msg := fmt.Sprintf(
			"Testing is only allowed for [%s] profile, "+
				"but the profile is [%s].",
			PROFILE_LOCAL,
			profile,
		)

		panic(msg)
	}

	configFileName := fmt.Sprintf("config_%s.yml", profile)
	configFilePath := constructConfigFilePath(isForTest, configFileName)

	viper.SetConfigFile(configFilePath)
	viper.ReadInConfig()
	viper.AutomaticEnv()

	config := Config{}

	switch profile {
	case PROFILE_LOCAL:
		viper.Unmarshal(&config)
	default:
		msg := "Somehow the profile is not proper"
		log.Info(msg)
	}

	return config
}

func mustGetProperProfile() string {
	viper.AutomaticEnv() // to get PROFILE from environment variable

	profile := viper.GetString(PROFILE)
	if !isGoodProfile(profile) {
		msg := fmt.Sprintf(
			"Retrieved profile is not good, "+
				"the profile must be one of [%s] (or not defined, i.e. empty string), "+
				"but it was %s.",
			PROFILE_LOCAL,
			profile,
		)

		panic(msg)
	}

	if profile == "" {
		profile = PROFILE_LOCAL // fallback to local profile
	}

	return profile
}

func isGoodProfile(profile string) bool {
	switch profile {
	case PROFILE_LOCAL, "":
		return true
	default:
		return false
	}
}

func constructConfigFilePath(isForTest bool, configFileName string) string {
	configFilePath := ""

	if isForTest {
		configFilePath = fmt.Sprintf("../config/configs/%s", configFileName)
	} else {
		configFilePath = fmt.Sprintf("./config/configs/%s", configFileName)
	}

	return configFilePath
}
