// +build darwin

package config

func (c *Config) platformDefaults() {
	if c.Node.Agent.TorBinaryPath == "" {
		c.Node.Agent.TorBinaryPath = "/Applications/TorBrowser.app/Contents/MacOS/Tor/tor.real"
	}
}
