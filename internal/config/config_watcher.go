package config

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/log"
)

// WatchConfig sets up a configuration file watcher that monitors for changes
// and reloads the configuration when necessary.
func (c *Config) WatchConfig() {
	configPath := c.V.ConfigFileUsed()

	if err := validateConfigFile(configPath); err != nil {
		log.Error(err)
		return
	}

	if err := c.initializeConfigState(); err != nil {
		log.Error("Failed to initialize config state:", err)
		return
	}

	go c.watchConfigChanges()
}

// watchConfigChanges continuously monitors the configuration file for changes
// and triggers a reload when necessary.
func (c *Config) watchConfigChanges() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	hashCheckCount := 0
	const hashCheckInterval = 12

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if c.hasConfigMtimeChanged() {
				c.reloadConfig("mTime changed")
				hashCheckCount = 0
				continue
			}

			if hashCheckCount >= hashCheckInterval {
				if c.hasConfigHashChanged() {
					c.reloadConfig("hash changed")
				}
				hashCheckCount = 0
			}

			hashCheckCount++
		}
	}
}

// initializeConfigState sets up the initial state of the configuration,
// including the last modification time and hash of the config file.
func (c *Config) initializeConfigState() error {
	info, err := os.Stat(c.V.ConfigFileUsed())
	if err != nil {
		return fmt.Errorf("getting initial file mTime: %v", err)
	}
	c.configLastModTime = info.ModTime()

	configHash, err := c.configFileHash(c.V.ConfigFileUsed())
	if err != nil {
		return fmt.Errorf("getting initial file hash: %v", err)
	}
	c.configHash = configHash

	return nil
}

// hasConfigHashChanged calculates and compares the current hash of the config file
// with the stored hash to detect content changes. Returns true if the hash has
// changed or if there was an error computing the new hash.
func (c *Config) hasConfigHashChanged() bool {
	configHash, err := c.configFileHash(c.V.ConfigFileUsed())
	if err != nil {
		log.Error("configFileHash", "err", err)
		return true
	}
	return c.configHash != configHash
}

// reloadConfig reloads the configuration when a change is detected.
func (c *Config) reloadConfig(reason string) {
	log.Infof("Config file %s, reloading config", reason)
	c.mu.Lock()
	defer c.mu.Unlock()

	newConfig := New()

	if err := newConfig.Load(); err != nil {
		log.Error("Reloading config:", err)
		return
	}

	*c = *newConfig

	c.updateConfigState()
}

// updateConfigState updates the configuration state after a reload.
func (c *Config) updateConfigState() {
	configHash, _ := c.configFileHash(c.V.ConfigFileUsed())
	c.configHash = configHash
	c.ReloadTimeStamp = time.Now().Format(time.RFC3339)
	info, _ := os.Stat(c.V.ConfigFileUsed())
	c.configLastModTime = info.ModTime()
}

// Function to calculate the SHA-256 hash of a file
func (c *Config) configFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// hasConfigMtimeChanged checks if the configuration file has been modified since the last check.
func (c *Config) hasConfigMtimeChanged() bool {
	info, err := os.Stat(c.V.ConfigFileUsed())
	if err != nil {
		log.Errorf("Checking config file: %v", err)
		return false
	}

	return info.ModTime().After(c.configLastModTime)
}
