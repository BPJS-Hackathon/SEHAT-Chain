package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bpjs-hackathon/sehat-chain/internal/core"
	"github.com/bpjs-hackathon/sehat-chain/types"
)

// Configuration structure
type Config struct {
	NodeID     string                  `json:"node_id"`
	Secret     string                  `json:"secret"`
	Port       string                  `json:"port"`
	APIPort    string                  `json:"api_port"`
	Validators []types.ValidatorConfig `json:"validators"`
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	nodeID := flag.String("id", "", "Node ID (overrides config file)")
	secret := flag.String("secret", "", "Node secret (overrides config file)")
	port := flag.String("port", "", "P2P Port (overrides config file)")
	flag.Parse()

	// Load configuration from file
	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		fmt.Println("Creating default config.json...")
		if err := createDefaultConfig(*configPath); err != nil {
			fmt.Printf("Failed to create default config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Default config created. Please edit config.json and restart.")
		os.Exit(0)
	}

	// Override config with command line flags if provided
	if *nodeID != "" {
		config.NodeID = *nodeID
	}
	if *secret != "" {
		config.Secret = *secret
	}
	if *port != "" {
		config.Port = *port
	}

	// Validate configuration
	if config.NodeID == "" || config.Secret == "" || config.Port == "" || config.APIPort == "" {
		fmt.Println("❌ Error: node_id, secret, and port must be specified")
		os.Exit(1)
	}

	// Determine if this node is a validator
	isValidator := false
	for _, v := range config.Validators {
		if v.ID == config.NodeID {
			isValidator = true
			break
		}
	}

	nodeType := "Light Node"
	if isValidator {
		nodeType = "Validator Node"
	}

	fmt.Println("========================================")
	fmt.Printf("Starting SEHAT-Chain %s\n", nodeType)
	fmt.Println("========================================")
	fmt.Printf("Node ID: %s\n", config.NodeID)
	fmt.Printf("P2P Port: %s\n", config.Port)
	fmt.Printf("Secret: %s\n", maskSecret(config.Secret))
	fmt.Printf("Known Validators: %d\n", len(config.Validators))
	fmt.Println("========================================")

	// Create and start node
	node := core.CreateNode(config.NodeID, config.Secret, config.Port, config.APIPort, config.Validators)

	fmt.Printf("Node %s created\n", config.NodeID)
	fmt.Println("Genesis block initialized")

	// Start the node (opens P2P and connects to network)
	if !isValidator {
		time.Sleep(time.Second * 3)
	}
	node.Start()

	fmt.Printf("Node %s is running on port %s\n", config.NodeID, config.Port)
	fmt.Printf("Connecting finished with final peer count: %d\n", len(node.P2P.Peers))

	if isValidator {
		fmt.Println("Validator mode: Ready to propose blocks")
	} else {
		fmt.Println("Light node mode: Listening for blocks")
	}

	// Tx Bodong
	if !isValidator {
		time.Sleep(time.Second * 10)
		fmt.Printf("timer to create fake tx")
		node.TestTx()
	}

	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")

	// Keep the program running
	select {}
}

// loadConfig loads configuration from JSON file
func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(path string) error {
	defaultConfig := Config{
		NodeID: "validator-1",
		Secret: "secret-validator-1",
		Port:   "9001",
		Validators: []types.ValidatorConfig{
			{
				ID:      "validator-1",
				Secret:  "secret-validator-1",
				Address: ":9001",
			},
			{
				ID:      "validator-2",
				Secret:  "secret-validator-2",
				Address: ":9002",
			},
			{
				ID:      "validator-3",
				Secret:  "secret-validator-3",
				Address: ":9003",
			},
		},
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(defaultConfig)
}

// maskSecret masks the secret for display purposes
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}
