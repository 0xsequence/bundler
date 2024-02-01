package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/multiformats/go-multiaddr"
)

type Config struct {
	GitCommit string `toml:"-"`

	SeedKey   string   `toml:"seed_key"` // if empty will generate new
	P2PPort   int      `toml:"p2p_port"`
	RPCPort   int      `toml:"rpc_port"`
	BootNodes []string `toml:"boot_nodes"`

	Logging LoggingConfig `toml:"logging"`

	BootNodeAddrs []multiaddr.Multiaddr `toml:"-"`
}

type LoggingConfig struct {
	Level   string `toml:"level"`
	JSON    bool   `toml:"json"`
	Concise bool   `toml:"concise"`
}

func NewFromFile(file string, env string, cfg *Config) error {
	if file == "" {
		file = env
	}
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return fmt.Errorf("failed to load config file: %w", err)
	}
	if _, err := toml.DecodeFile(file, cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	return initConfig(cfg)
}

func initConfig(cfg *Config) error {
	bootNodeAddrs := make([]multiaddr.Multiaddr, 0, len(cfg.BootNodes))
	for _, s := range cfg.BootNodes {
		addr, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			return err
		}
		bootNodeAddrs = append(bootNodeAddrs, addr)
	}
	cfg.BootNodeAddrs = bootNodeAddrs

	return nil
}
