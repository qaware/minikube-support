package plugins

import (
	"fmt"
	"github.com/chr-fritz/minikube-support/pkg/apis"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

// CombinedStartStopPlugin is a simple plugin that combines several plugins together using a combine function.
type CombinedStartStopPlugin struct {
	pluginName     string
	plugins        []apis.StartStopPlugin
	singleRunnable bool
}

// CombineFunc combines several plugins and returns them as array.
type CombineFunc func() ([]apis.StartStopPlugin, error)

// NewCombinedPlugin creates a new plugin that combines some more plugins to one.
func NewCombinedPlugin(pluginName string, plugins []apis.StartStopPlugin, singleRunnable bool) (*CombinedStartStopPlugin, error) {
	if len(plugins) < 2 {
		return nil, fmt.Errorf("at least two plugins are required to combine them. Only %v given", len(plugins))
	}

	return &CombinedStartStopPlugin{
		pluginName:     pluginName,
		plugins:        plugins,
		singleRunnable: singleRunnable,
	}, nil
}

// String returns the plugin name.
func (c *CombinedStartStopPlugin) String() string {
	return c.pluginName
}

func (c *CombinedStartStopPlugin) IsSingleRunnable() bool {
	return c.singleRunnable
}

// Start really combines the plugins together and starts them all.
func (c *CombinedStartStopPlugin) Start(messageChannel chan *apis.MonitoringMessage) (string, error) {
	var errors *multierror.Error
	for _, plugin := range c.plugins {
		_, err := plugin.Start(messageChannel)
		if err != nil {
			errors = multierror.Append(errors, err)
		}
	}
	return c.pluginName, errors.ErrorOrNil()
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
