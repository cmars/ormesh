// +build darwin

package config

func (a *config.Config) platformDefaults() {
	if a.TorBinaryPath == "" {
		a.TorBinaryPath = "/usr/bin/tor"
	}
}
