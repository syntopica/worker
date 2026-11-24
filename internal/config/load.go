package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
)

type rawConfig struct {
	Listen                *string               `json:"listen"`
	Models                map[string]rawModel   `json:"models"`
	Queues                map[string]rawQueue   `json:"queues"`
	Nodes                 map[string]rawNode    `json:"nodes"`
	Producers             map[string][]string   `json:"producers"`
	Privacy               map[string]rawClass   `json:"privacy"`
	MaxPayloadBytes       *int                  `json:"max_payload_bytes"`
	MaxSplit              *int                  `json:"max_split"`
	MaxParkedRuns         *int                  `json:"max_parked_runs"`
	SplitAfterPreemptions *int                  `json:"split_after_preemptions"`
	Profiles              map[string]rawProfile `json:"profiles"`
	Runners               map[string]rawRunner  `json:"runners"`
}

type rawModel struct {
	NumCtx    int     `json:"num_ctx"`
	KeepAlive string  `json:"keep_alive"`
	ColdGB    float64 `json:"cold_gb"`
	WarmGB    float64 `json:"warm_gb"`
}

type rawClass struct {
	Executors []string `json:"executors"`
	Trust     []string `json:"trust"`
}

const (
	defaultPayloadBytes = 1_048_576
	defaultSplit        = 8
	defaultParkedRuns   = 3
	defaultPreemptions  = 3
)

var (
	errRequired = errors.New("config: models, queues, nodes and producers are required")
	errLoopback = errors.New("config: listen is not a loopback address")
)

// Load parses config.json, applying the spec's defaults, and refuses an
// invalid value.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	var raw rawConfig

	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	if raw.Models == nil || raw.Queues == nil || raw.Nodes == nil || raw.Producers == nil {
		return nil, errRequired
	}

	config := &Config{
		Models:                make(map[string]ModelPin, len(raw.Models)),
		Nodes:                 make(map[string]NodePolicy, len(raw.Nodes)),
		Queues:                make(map[string]QueuePolicy, len(raw.Queues)),
		Profiles:              make(map[string]TaskProfile, len(raw.Profiles)),
		Producers:             raw.Producers,
		MaxPayloadBytes:       orDefault(raw.MaxPayloadBytes, defaultPayloadBytes),
		MaxSplit:              orDefault(raw.MaxSplit, defaultSplit),
		MaxParkedRuns:         orDefault(raw.MaxParkedRuns, defaultParkedRuns),
		SplitAfterPreemptions: orDefault(raw.SplitAfterPreemptions, defaultPreemptions),
	}

	err = fill(config, &raw, filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	return config, checkRoutes(config)
}

func fill(config *Config, raw *rawConfig, configDir string) error {
	err := listen(config, orDefault(raw.Listen, "127.0.0.1:8765"))
	if err != nil {
		return err
	}

	config.Privacy, config.Trust = defaultClasses()
	for class, override := range raw.Privacy {
		config.Privacy[class], config.Trust[class] = override.Executors, override.Trust
	}

	for name, model := range raw.Models {
		config.Models[name] = ModelPin{
			Name: name, NumCtx: model.NumCtx, KeepAlive: model.KeepAlive, ColdGB: model.ColdGB, WarmGB: model.WarmGB,
		}
	}

	for name := range raw.Nodes {
		node := raw.Nodes[name]
		config.Nodes[name] = parseNode(name, &node)
	}

	for name := range raw.Queues {
		queue := raw.Queues[name]

		config.Queues[name], err = parseQueue(name, &queue)
		if err != nil {
			return err
		}
	}

	for name := range raw.Profiles {
		profile := raw.Profiles[name]

		config.Profiles[name], err = parseProfile(name, &profile, configDir)
		if err != nil {
			return err
		}
	}

	return fillRunners(config, raw.Runners)
}

// listen reads the bind address. Tokens travel in plaintext; only a
// loopback bind keeps them on the host.
func listen(config *Config, address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("config: listen: %w", err)
	}

	ip, err := netip.ParseAddr(host)
	if host != "localhost" && (err != nil || !ip.IsLoopback()) {
		return errLoopback
	}

	config.ListenHost = host

	config.ListenPort, err = strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("config: listen: %w", err)
	}

	return nil
}
