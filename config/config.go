package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/multiformats/go-multiaddr"
)

type Config struct {
	GitCommit string `toml:"-"`

	Mnemonic string `toml:"mnemonic"`

	P2PPort   int      `toml:"p2p_port"`
	RPCPort   int      `toml:"rpc_port"`
	BootNodes []string `toml:"boot_nodes"`

	Logging LoggingConfig `toml:"logging"`

	NetworkConfig   NetworkConfig   `toml:"network"`
	MempoolConfig   MempoolConfig   `toml:"mempool"`
	SendersConfig   SendersConfig   `toml:"senders"`
	CollectorConfig CollectorConfig `toml:"collector"`

	BootNodeAddrs []multiaddr.Multiaddr `toml:"-"`
}

type LoggingConfig struct {
	ServiceName     string `toml:"service"`
	Level           string `toml:"level"`
	JSON            bool   `toml:"json"`
	Concise         bool   `toml:"concise"`
	RequestHeaders  bool   `toml:"req_headers"`
	ResponseHeaders bool   `toml:"resp_headers"`
	Source          string `toml:"source"`
}

type NetworkConfig struct {
	RpcUrl  string `toml:"rpc_url"`
	IpfsUrl string `toml:"ipfs_url"`

	ValidatorContract string `toml:"validator_contract"`
}

type MempoolConfig struct {
	Size        uint `toml:"max_size"`
	IngressSize uint `toml:"max_ingress_size"`

	MaxEndorserGasLimit uint `toml:"max_endorser_gas_limit"`
}

type PrunerConfig struct {
	GracePeriodSeconds int `toml:"grace_period"`
	RunWaitMillis      int `toml:"run_wait_millis"`
}

type SendersConfig struct {
	NumSenders uint `toml:"num_senders"`
}

type CollectorConfig struct {
	PriorityFee    int64   `toml:"min_priority_fee"`
	PriorityFeeMul float64 `toml:"priority_fee_mul"`

	References []PriceReference `toml:"references"`
}

type PriceReference struct {
	Token string `toml:"token"`

	UniswapV2 *UniswapV2Reference `toml:"uniswap_v2"`
}

type UniswapV2Reference struct {
	Pool      string `toml:"pool"`
	BaseToken string `toml:"base_token"`
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
