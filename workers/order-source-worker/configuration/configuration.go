/*
 * MIT License
 *
 * Copyright (c) 2020 Tom Greasley
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

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
	c.v.SetDefault(varFreq, 2.5)
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

func (c *Configuration) GetFrequency() float64 {
	return c.v.GetFloat64(varFreq)
}

// GetPathToConfig returns the path to the config file
func (c *Configuration) GetPathToConfig() string {
	return c.v.GetString(varPathToConfig)
}
