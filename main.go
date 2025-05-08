package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/mhunterdev/webHost/src/certs"
	"github.com/mhunterdev/webHost/src/docker"
	"github.com/mhunterdev/webHost/src/requirements"
)

// Checks for the .env file
func CheckEnvFile() bool {
	// Check if the .env file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return false
	} else {
		fmt.Println("Beginning environment configuration...")
		return true
	}
}

// Sets basic .env vars
func SetEnv() {
	// Create a .env file
	_, err := os.Create(".env")
	if err != nil {
		fmt.Println("Error creating .env file:", err)
		return
	}

	OS, _ := requirements.GetPackageManager()

	// Start the prompt for environment variables
	fmt.Println("Gathering user environment inputs...")

	// Helper function to read input
	readInput := func(prompt string) string {
		fmt.Println(prompt)
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return scanner.Text()
		}
		fmt.Println("Error reading input:", scanner.Err())
		return ""
	}

	// Collect inputs
	envVars := map[string]string{
		"PACKAGE_MANAGER": OS,
		"CN":              readInput("Enter the Common Name (CN) for the certificate:"),
		"O":               readInput("Enter the Organization (O) for the certificate:"),
		"OU":              readInput("Enter the Organizational Unit (OU) for the certificate:"),
		"C":               readInput("Enter the Country (C) for the certificate:"),
		"ST":              readInput("Enter the State/Province (ST) for the certificate:"),
		"L":               readInput("Enter the Locality (L) for the certificate:"),
		"EMAIL":           readInput("Enter the Email for the certificate:"),
		"KEY_SIZE":        readInput("Enter the Key Size for the certificate (2048, 4096):"),
		"DNS":             readInput("DNS Alternative Names for the certificate (comma separated):"),
		"DOCKER_IP":       readInput("Enter IP address for local docker container"),
		"POST_USER":       readInput("Enter a username for the postgresql install"),
		"PG_PORT":         readInput("Enter the port postgrsql should listen on"),
		"DBNAME":          readInput("Enter a name for the new database"),
		"SUBNET":          readInput("Enter the subnet for the docker network (e.g.,192.168.100/24)"),
		"SSL_MODE":        "verify-full",
		"DB_CERT":         "certs/.postgres/certs/server.crt",
		"DB_KEY":          "certs/.postgres/certs/private/server.key",
		"DB_ROOT":         "certs/.postgres/certs/private/root.crt",
	}

	// Write all inputs to the .env file
	err = godotenv.Write(envVars, ".env")
	if err != nil {
		fmt.Println("Error writing to .env file:", err)
		return
	}

	fmt.Println("Environment variables successfully written to .env file.")
}

// BuildRequirements combines minor functions into a single action to build the requirements
func BuildRequirements() {
	requirements.CreateDirs()
	requirements.InstallOpenSSL()
	requirements.InstallDocker()
	docker.StartDocker()
	requirements.InstallDockerCompose()
	docker.ConfigureDocker()
	docker.ConfigureDockerEnv()
	certs.BuildCACerts()
	docker.StartDB()
}

func main() {
	// Check if the .env file exists
	if !CheckEnvFile() {
		BuildRequirements()
	} else {
		return
	}
}
