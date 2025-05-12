package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mhunterdev/webHost/src/pgconnect"
	"github.com/mhunterdev/webHost/src/requirements"
	"golang.org/x/crypto/ssh/terminal"
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
	rootDir, _ := exec.Command("pwd").Output()
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

	// Helper function to read secure input (hidden password)
	readSecureInput := func(prompt string) string {
		fmt.Println(prompt)
		for {
			bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				fmt.Println("Error reading password:", err)
				return ""
			}
			fmt.Println() // Print a newline after password input
			fmt.Println("Confirm password:")
			byteConfirmPassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				fmt.Println("Error reading password:", err)
				return ""
			}
			if string(byteConfirmPassword) != string(bytePassword) {
				fmt.Println("Passwords do not match. Please try again.")
				continue
			} else {
				fmt.Println("")
				return string(bytePassword)
			}
		}
	}

	// Collect inputs
	envVars := map[string]string{
		"OS":            OS,
		"ROOT_DIR":      string(rootDir),
		"DOCKER_IP":     readInput("Enter IP address for local docker container"),
		"POST_USER":     readInput("Enter a username for the postgresql install"),
		"POST_PASSWORD": readSecureInput("Enter a password for the postgresql install:"),
		"PG_PORT":       readInput("Enter the port postgrsql should listen on"),
		"DBNAME":        readInput("Enter a name for the new database"),
		"SUBNET":        readInput("Enter the subnet for the docker network (e.g.,192.168.100.0/24)"),
		"SSL_MODE":      "verify-full",
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
func BuildContainers() {

	godotenv.Load(".env")

	// Helper function to verify docker is installed
	docker := func() bool {
		// Use os/exec to check if docker is installed
		cmd := exec.Command("docker", "--version")
		err := cmd.Run()
		if err != nil {
			return false
		}
		return true
	}

	//Verify the installation of docker
	// If docker is not installed, install it
	for {
		if !docker() {
			fmt.Println("Docker is not installed. Installing Docker...")
			requirements.InstallDocker()
			break
		} else {
			fmt.Println("Docker is already installed. Setting Docker environment...")
			break
		}
	}

	//Create the docker configuration and mount directory
	os.Mkdir(".docker", 0755)
	os.Mkdir(".docker/compose", 0755)
	os.MkdirAll(".docker/compose/mounts", 0755)
	os.MkdirAll(".docker/compose/mounts/custom", 0755)

	// Set the docker environment configuration
	user := os.Getenv("POST_USER")
	password := os.Getenv("POST_PASSWORD")
	dbname := os.Getenv("DBNAME")
	dockerip := os.Getenv("DOCKER_IP")
	dbport := os.Getenv("PG_PORT")
	subnet := os.Getenv("SUBNET")
	conf := fmt.Sprintf("POST_USER=%s\nPOST_PASSWORD=%s\nDBNAME=%s\nPG_PORT=%s\nDOCKER_IP=%s\nSUBNET=%s\n", user, password, dbname, dbport, dockerip, subnet)

	file, err := os.Create(".docker/compose/.env")
	if err != nil {
		return
	}
	defer file.Close()

	//Write environment file
	_, err = file.Write([]byte(conf))
	if err != nil {
		fmt.Println("Error writing to .env file:", err)
		return
	}

	// Enable and start Docker service
	cmd := "sudo systemctl enable docker && sudo systemctl restart docker"
	err2 := exec.Command("bash", "-c", cmd).Run()
	if err2 != nil {
		fmt.Println("Docker is not running. Please start Docker and try again.")
		return
	}
	fmt.Println("Docker started successfully.")

	// Check if docker-compose is installed
	cmd = "docker-compose --version"
	err = exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		fmt.Println("Docker Compose is not installed. Installing Docker Compose...")
		requirements.InstallDockerCompose()
	} else {
		fmt.Println("Docker Compose is already installed.")
	}

	// Create the docker-compose.yml file
	compose := `# THIS FILE IS AUTO-GENERATED
#DO NOT EDIT THIS FILE
#TO MAKE CHANGES, EDIT THE .env FILE
services:
  postgres:
    image: postgres:16
    container_name: postgres_container
    environment:
      POSTGRES_USER: ${POST_USER}
      POSTGRES_PASSWORD: ${POST_PASSWORD}
      POSTGRES_DB: ${DBNAME}
    ports:
      - "${PG_PORT}:5432"
    networks:
      custom_network:
        ipv4_address: ${DOCKER_IP}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - .docker/compose/mounts:/var/lib/postgresql/custom
    command:
      - "postgres"
      - "-c"
      - "config_file=/var/lib/postgresql/data/postgresql.conf"
      - "-c"
      - "hba_file=/var/lib/postgresql/custom/pg_hba.conf"

networks:
  custom_network:
    driver: bridge
    ipam:
      config:
        - subnet: ${SUBNET}

volumes:
  postgres_data:
`

	// Create the directory if it doesn't exist
	yamlFile, err := os.Create(".docker/compose/docker-compose.yml")
	if err != nil {
		fmt.Println("Error creating docker-compose.yml file:", err)
		return
	}
	defer yamlFile.Close()

	_, err = yamlFile.Write([]byte(compose))
	if err != nil {
		fmt.Println("Error writing to docker-compose.yml file:", err)
		return
	}

	fmt.Println("Docker configuration file created successfully.")

	// Start the Docker container using docker-compose
	fmt.Println("Starting Postgres container...")
	cmd2 := "sudo docker-compose -f .docker/compose/docker-compose.yml up -d"
	err3 := exec.Command("bash", "-c", cmd2).Run()
	if err3 != nil {
		fmt.Println("Error starting Docker container:", err)
		return
	}
	// Check if the container is running
	cmd = "sudo docker ps"
	err = exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		fmt.Println("Error checking Docker container status:", err)
		return
	}

	fmt.Println("Database container started successfully.")
}

func main() {
	// Check if the .env file exists
	if !CheckEnvFile() {
		SetEnv()
		BuildContainers()
	} else {
		fmt.Println("Environment file already exists. Skipping environment setup.")
	}
	pgconnect.TestConnection()

}
