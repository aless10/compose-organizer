package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Windows map[string]WindowConfig `yaml:"windows"`
}

type WindowConfig struct {
	Services       []string `yaml:"services"`
	Command        string   `yaml:"command"`
	CommandOptions []string `yaml:"command_options"`
}

func main() {
	var rootCmd = &cobra.Command{Use: "gocli"}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Print the configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			printConfig(configPath)
		},
	}

	var printCmd = &cobra.Command{
		Use:   "dryUp",
		Short: "Print command to create tmux windows and execute Docker Compose commands",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			var sessionName string
			if len(args) == 1 {
				sessionName = args[0]
			} else {
				sessionName = "defaultSessionName"
			}
			config, err := loadConfig(configPath)
			if err != nil {
				log.Fatalf("Error loading configuration: %v", err)
			}
			tmuxCmd := "tmux new-session -d -s " + sessionName + " -n Window \n"
			// 			session="testing2"
			// tmux new-session -d -s $session -n Window1
			// tmux split-window -h -p 33 -t $session
			// tmux send-keys -t $session:Window1.1 "echo 'sono nel pane 1'" C-m
			// tmux send-keys -t $session:Window1.2 "echo 'sono nel pane 2'" C-m
			// Iterate through each window in the configuration
			i := 1
			for windowName, windowConfig := range config.Windows {
				// Create a new tmux window with the specified services and command options
				if i > 1 {
					tmuxCmd += "tmux split-window -h -p 50 -t " + sessionName + "\n"
				}
				result, err := printTmuxCommand(sessionName, i, windowConfig.Command, windowConfig.Services, windowConfig.CommandOptions)
				if err != nil {
					log.Printf("Error creating tmux window %s: %v", windowName, err)
				}
				i += 1
				tmuxCmd += result
			}
			fmt.Println(tmuxCmd)
		},
	}

	var upCmd = &cobra.Command{
		Use:   "up",
		Short: "Create tmux windows and execute Docker Compose commands",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			config, err := loadConfig(configPath)
			if err != nil {
				log.Fatalf("Error loading configuration: %v", err)
			}

			// Iterate through each window in the configuration
			for windowName, windowConfig := range config.Windows {
				// Create a new tmux window with the specified services and command options
				err := createTmuxWindow(windowName, windowConfig.Services, windowConfig.CommandOptions)
				if err != nil {
					log.Printf("Error creating tmux window %s: %v", windowName, err)
				}
			}
		},
	}

	rootCmd.PersistentFlags().StringP("config", "c", "config.yaml", "Path to the configuration file")

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(printCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func loadConfig(configPath string) (*Config, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return &config, nil
}

func printTmuxCommand(sessionName string, pane int, command string, services []string, commandOptions []string) (string, error) {
	// Build the Docker Compose command
	dockerComposeCmd := "docker compose "
	dockerComposeCmd += command + " "
	dockerComposeCmd += strings.Join(services, " ") + " "
	dockerComposeCmd += strings.Join(commandOptions, " ") + " "

	// Run the Docker Compose command in the tmux window
	cmd := "tmux send-keys -t " + sessionName + ":Window." + strconv.Itoa(pane) + " " + "\"" + dockerComposeCmd + "\" C-m\n"
	// Create the tmux window with the specified name

	// fmt.Println(cmd)
	return cmd, nil
}

func createTmuxWindow(windowName string, services []string, commandOptions []string) error {
	// Create the tmux window with the specified name
	cmd := exec.Command("tmux", "new-window", "-n", windowName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error creating tmux window: %v", err)
	}

	// Build the Docker Compose command
	dockerComposeCmd := []string{"docker-compose"}
	dockerComposeCmd = append(dockerComposeCmd, "up")
	dockerComposeCmd = append(dockerComposeCmd, commandOptions...)
	dockerComposeCmd = append(dockerComposeCmd, services...)

	// Run the Docker Compose command in the tmux window
	cmd = exec.Command("tmux", "send-keys", "-t", windowName+":0", strings.Join(dockerComposeCmd, " ")+" C-m")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error sending Docker Compose command to tmux window: %v", err)
	}

	return nil
}

func printConfig(configPath string) {
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("Error marshalling config to YAML: %v", err)
	}

	fmt.Println(string(yamlData))
}
