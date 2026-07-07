package scaffold

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func IsInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func PromptInfraOptions(projectName string, preset *InfraOptions) InfraOptions {
	if preset != nil {
		opts := *preset
		if opts.SetupDatabase && opts.Database == "" {
			opts.Database = DatabasePostgres
		}
		return opts
	}

	if !IsInteractive() {
		return InfraOptions{}
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	fmt.Println("Infrastructure setup")

	opts := InfraOptions{}
	opts.SetupDatabase = askYesNo(reader, "Setup database with Docker Compose?", false)
	if opts.SetupDatabase {
		opts.Database = askDatabase(reader)
	}
	opts.SetupRedis = askYesNo(reader, "Setup Redis for caching with Docker Compose?", false)
	fmt.Println()

	_ = projectName
	return opts
}

func askYesNo(reader *bufio.Reader, prompt string, defaultYes bool) bool {
	suffix := "[y/N]"
	if defaultYes {
		suffix = "[Y/n]"
	}
	for {
		fmt.Printf("%s %s: ", prompt, suffix)
		line, err := reader.ReadString('\n')
		if err != nil {
			return defaultYes
		}
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" {
			return defaultYes
		}
		switch line {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please answer y or n.")
		}
	}
}

func askDatabase(reader *bufio.Reader) string {
	fmt.Println("Choose database engine:")
	fmt.Println("  1) PostgreSQL (recommended)")
	fmt.Println("  2) MySQL")
	for {
		fmt.Print("Select [1]: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return DatabasePostgres
		}
		line = strings.TrimSpace(line)
		if line == "" || line == "1" {
			return DatabasePostgres
		}
		if line == "2" || strings.EqualFold(line, "mysql") {
			return DatabaseMySQL
		}
		fmt.Println("Invalid choice. Enter 1 or 2.")
	}
}
