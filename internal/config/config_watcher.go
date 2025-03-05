package config

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/log"
)

// WatchConfig sets up a configuration file watcher that monitors for changes
// and reloads the configuration when necessary.
func (c *Config) WatchConfig(ctx context.Context) {
	configPath := c.V.ConfigFileUsed()

	if fileErr := validateConfigFile(configPath); fileErr != nil {
		log.Error(fileErr)
		return
	}

	if initErr := c.initializeConfigState(); initErr != nil {
		log.Error("Failed to initialize config state", "err", initErr)
		return
	}

	go c.watchConfigChanges(ctx)
}

// watchConfigChanges continuously monitors the configuration file for changes
// and triggers a reload when necessary.
func (c *Config) watchConfigChanges(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	hashCheckCount := 0
	const hashCheckInterval = 12

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
		return fmt.Errorf("getting initial file mTime: %w", err)
	}
	c.configLastModTime = info.ModTime()

	configHash, hashErr := c.configFileHash(c.V.ConfigFileUsed())
	if hashErr != nil {
		return fmt.Errorf("getting initial file hash: %w", hashErr)
	}
	c.configHash = configHash

	return nil
}

// reloadConfig reloads the configuration when a change is detected.
func (c *Config) reloadConfig(reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Infof("Config file %s, reloading config", reason)

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
	if _, copyErr := io.Copy(hasher, file); copyErr != nil {
		return "", copyErr
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// hasConfigHashChanged calculates and compares the current hash of the config file
// with the stored hash to detect content changes. Returns true if the hash has
// changed or if there was an error computing the new hash.
func (c *Config) hasConfigHashChanged() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	configHash, err := c.configFileHash(c.V.ConfigFileUsed())
	if err != nil {
		log.Error("configFileHash", "err", err)
		return true
	}
	return c.configHash != configHash
}

// hasConfigMtimeChanged checks if the configuration file has been modified since the last check.
func (c *Config) hasConfigMtimeChanged() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info, err := os.Stat(c.V.ConfigFileUsed())
	if err != nil {
		log.Error("Checking config file", "err", err)
		return false
	}

	return info.ModTime().After(c.configLastModTime)
}
