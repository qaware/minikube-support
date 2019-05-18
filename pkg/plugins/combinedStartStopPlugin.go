package plugins

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/sirupsen/logrus"
)

const CombinedStartStopPluginName = "combined"

// CombinedStartStopPlugin is a simple plugin that combines several plugins together using a combine function.
type CombinedStartStopPlugin struct {
	combineFunc CombineFunc
	plugins     []apis.StartStopPlugin
}

// CombineFunc combines several plugins and returns them as array.
type CombineFunc func() ([]apis.StartStopPlugin, error)

// NewCombinedPlugin creates a new plugin that combines some more plugins to one.
func NewCombinedPlugin(combineFunc CombineFunc) apis.StartStopPlugin {
	return &CombinedStartStopPlugin{
		combineFunc: combineFunc,
	}
}

// String returns the plugin name.
func (*CombinedStartStopPlugin) String() string {
	return CombinedStartStopPluginName
}

// Start really combines the plugins together and starts them all.
func (c *CombinedStartStopPlugin) Start(messageChannel chan *apis.MonitoringMessage) (string, error) {
	if c.combineFunc == nil {
		return "", fmt.Errorf("can not start the combined plugin: combine function is nil")
	}

	plugins, e := c.combineFunc()
	if e != nil {
		return "", fmt.Errorf("can not start all combined plugins: %s", e)
	}

	for _, plugin := range plugins {
		_, err := plugin.Start(messageChannel)
		if err != nil {
			logrus.Errorf("Unable to start plugin %s: %s", plugin, err)
		} else {
			c.plugins = append(c.plugins, plugin)
		}
	}
	return CombinedStartStopPluginName, nil
}

// Stop stops all plugins.
func (c *CombinedStartStopPlugin) Stop() error {
	for _, plugin := range c.plugins {
		go func() {
			logrus.Debugf("Terminating plugin: %s", plugin)
			e := plugin.Stop()
			if e != nil {
				logrus.Warnf("Unable to terminate plugin %s: %s", plugin, e)
			}
		}()
	}
	return nil
}
