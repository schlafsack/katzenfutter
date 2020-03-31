package configuration

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

const (
	varBrokerAddr   = "worker.zeebe.brokerEndpoint"
	varFreq         = "worker.frequency"
	varPathToConfig = "config.file"
)

type Configuration struct {
	v *viper.Viper
}

func New() *Configuration {
	c := Configuration{
		v: viper.New(),
	}

	c.v.SetDefault(varPathToConfig, "config.yml")
	c.v.SetDefault(varBrokerAddr, "0.0.0.0:26500")
	c.v.SetDefault(varFreq, 10)
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetTypeByDefaultValue(true)
	c.v.SetConfigFile(c.GetPathToConfig())
	err := c.v.ReadInConfig() // Find and read the config file
	log.Println("loading config", "path", c.GetPathToConfig())
	// just use the default value(s) if the config file was not found
	if _, ok := err.(*os.PathError); ok {
		log.Printf("no config file '%s' not found. Using default values", c.GetPathToConfig())
	} else if err != nil { // Handle other errors that occurred while reading the config file
		panic(fmt.Errorf("fatal error while reading the config file: %s", err))
	}
	// monitor the changes in the config file
	c.v.WatchConfig()
	c.v.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed", "file", e.Name)
	})
	return &c
}

func (c *Configuration) GetBrokerEndpoint() string {
	return c.v.GetString(varBrokerAddr)
}

func (c *Configuration) GetFrequency() int {
	return c.v.GetInt(varFreq)
}

// GetPathToConfig returns the path to the config file
func (c *Configuration) GetPathToConfig() string {
	return c.v.GetString(varPathToConfig)
}
