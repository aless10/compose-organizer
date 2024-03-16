package main

import (
	"bufio"
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
			command := createCommand(args, configPath)
			fmt.Println(command)
		},
	}

	var upCmd = &cobra.Command{
		Use:   "up",
		Short: "Create tmux windows and execute Docker Compose commands",
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("config")
			command := createCommand(args, configPath)
			fmt.Println("You are about to run this command")
			fmt.Println(command)
			fmt.Println("Do you want to continue?[Y/n]")
			reader := bufio.NewReader(os.Stdin)
			// ReadString will block until the delimiter is entered
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("An error occured while reading input. Please try again", err)
				return
			}
			input = strings.TrimSuffix(input, "\n")
			if input != "Y" && input != "y" {
				return
			}
			formatCommand := formatCommand(command)
			err = execCommand(formatCommand)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
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
	return cmd, nil
}

func createCommand(args []string, configPath string) string {
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
	return tmuxCmd
}

func formatCommand(command string) string {
	return strings.ReplaceAll(command, "\n", ";")
}

func execCommand(command string) error {
	// Create the tmux window with the specified name
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error while running command: %v", err)
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
