package setup

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	OS, _ := getPackageManager()
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
		"DOCKER_IP":     "127.0.0.1",
		"POST_USER":     readInput("Enter a username for the postgresql install"),
		"POST_PASSWORD": readSecureInput("Enter a password for the postgresql install:"),
		"PG_PORT":       readInput("Enter the port postgrsql should listen on"),
		"DBNAME":        readInput("Enter a name for the new database"),
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

// Determines what package manager to use
func getPackageManager() (string, error) {

	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "", fmt.Errorf("error opening /etc/os-release: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			// Extract the value of ID
			id := strings.Trim(line[len("ID="):], `"`)
			switch id {
			case "ubuntu", "debian":
				fmt.Println("Debian based OS detected")
				return "apt", nil
			case "centos", "rhel", "fedora", "rocky", "almalinux":
				fmt.Println("RHEL based OS detected")
				return "dnf", nil
			case "arch":
				fmt.Println("Arch based OS detected")
				return "pacman", nil
			default:
				return "", fmt.Errorf("unsupported Linux distribution: %s", id)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading /etc/os-release: %v", err)
	}

	return "", fmt.Errorf("could not determine Linux distribution")
}

// LogError writes errors to the Install.log file
func LogError(message string, err error) {
	logFile, logErr := os.OpenFile("Install.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logErr != nil {
		fmt.Println("Error opening log file:", logErr)
		return
	}
	defer logFile.Close()

	logMessage := fmt.Sprintf("%s: %v\n", message, err)
	logFile.WriteString(logMessage)
	fmt.Println(message, err)
}

// LogInfo writes informational messages to the Install.log file
func LogInfo(message string) {
	logFile, logErr := os.OpenFile("Install.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logErr != nil {
		fmt.Println("Error opening log file:", logErr)
		return
	}
	defer logFile.Close()

	logFile.WriteString(message + "\n")
	fmt.Println(message)
}

// verifyDockerInstallation checks if Docker is installed
func verifyDockerInstallation() bool {
	cmd := exec.Command("docker", "--version")
	err := cmd.Run()
	if err != nil {
		LogError("Docker is not installed.", err)
		return false
	}
	LogInfo("Docker is already installed.")
	return true
}

// enableAndStartDocker enables and starts the Docker service
func enableAndStartDocker() error {
	cmd := "sudo systemctl enable docker && sudo systemctl restart docker"
	err := exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		LogError("Failed to enable and start Docker service.", err)
		return err
	}
	LogInfo("Docker service started successfully.")
	return nil
}

// verifyDockerComposeInstallation checks if Docker Compose is installed
func verifyDockerComposeInstallation() error {
	cmd := exec.Command("docker-compose", "--version")
	err := cmd.Run()
	if err != nil {
		LogError("Docker Compose is not installed.", err)
		installDockerCompose()
	} else {
		LogInfo("Docker Compose is already installed.")
	}
	return nil
}

// startDockerContainer starts the PostgreSQL container using docker-compose
func startDockerContainer() error {
	cmd := "sudo docker-compose -f docker/docker-compose.yml up -d"
	err := exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		LogError("Error starting Docker container.", err)
		return err
	}
	LogInfo("PostgreSQL container started successfully.")
	return nil
}

// UpdatePgHbaConf updates the pg_hba.conf file inside the container
func UpdatePgHbaConf(containerName string) error {
	updateCmd := `echo 'host all all 0.0.0.0/0 md5' >> /var/lib/postgresql/data/pg_hba.conf`
	err := exec.Command("sudo", "docker", "exec", containerName, "bash", "-c", updateCmd).Run()
	if err != nil {
		LogError("Error updating pg_hba.conf.", err)
		return err
	}
	LogInfo("pg_hba.conf updated successfully.")
	return nil
}

// reloadPostgresConfiguration reloads the PostgreSQL configuration
func reloadPostgresConfiguration(containerName string) error {
	reloadCmd := exec.Command("sudo", "docker", "exec", containerName, "pg_ctl", "reload")
	err := reloadCmd.Run()
	if err != nil {
		LogError("Error reloading PostgreSQL configuration.", err)
		return err
	}
	LogInfo("PostgreSQL configuration reloaded successfully.")
	return nil
}

// Verifys the package manager installs the docker requirement
func installDocker() { // Corrected function name
	// Load the environment variables
	godotenv.Load(".env")
	fmt.Println("Installing docker requirement...")

	os := os.Getenv("OS")
	if os == "" {
		fmt.Println("Error: OS environment variable is not set.")
		return
	}

	// Switch to determine how we are installing Docker
	var cmd *exec.Cmd
	switch os {
	case "apt":
		cmd = exec.Command("bash", "-c", "sudo apt update && sudo apt install -y docker")
	case "dnf":
		cmd = exec.Command("sudo", "dnf", "install", "-y", "docker")
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "docker")
	default:
		fmt.Println("Unsupported package manager")
		return
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Error installing Docker:", err)
		return
	}
	fmt.Println("Docker installed successfully.")
}

// Installs the openssl requirement
func installOpenSSL() {
	godotenv.Load(".env")
	fmt.Println("Installing openssl requirement...")

	os := os.Getenv("OS")
	if os == "" {
		fmt.Println("Error: OS environment variable is not set.")
		return
	}

	// Placeholder variable for the cmd
	var cmd *exec.Cmd

	switch os {
	case "apt":
		cmd = exec.Command("bash", "-c", "sudo apt update && sudo apt install -y openssl")
	case "dnf":
		cmd = exec.Command("sudo", "dnf", "install", "-y", "openssl")
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "openssl")
	default:
		fmt.Println("Unsupported package manager")
		return
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Error installing OpenSSL:", err)
		return
	}
	fmt.Println("OpenSSL installed successfully.")
}

// Verifys the os and installs the docker-compose requirement
func installDockerCompose() {
	fmt.Println("Installing docker-compose requirement...")

	godotenv.Load(".env")

	//Load the environment variables
	os := os.Getenv("OS")
	if os == "" {
		fmt.Println("Error: OS environment variable is not set.")
		return
	}

	//Switch to determine how we are installing postgres
	var cmd *exec.Cmd
	switch os {
	case "apt":
		cmd = exec.Command("sudo", "apt", "update", "&&", "sudo", "apt", "install", "-y", "docker-compose")
	case "dnf":
		cmd = exec.Command("sudo", "dnf", "install", "-y", "docker-compose")
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "docker-compose")
	default:
		fmt.Println("Unsupported package manager")
		return
	}
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error installing docker-compose...")
		return
	}
	fmt.Println("Docker-compose installed successfully.")
}

// BuildContainers verifies docker intallation and
// configuration yaml is in place
// If not, it builds them and starts the postgresql container
func BuildContainers() {
	godotenv.Load(".env")

	if !verifyDockerInstallation() {
		LogInfo("Installing Docker...")
		installDocker() // Corrected function call
	}

	if err := enableAndStartDocker(); err != nil {
		return
	}

	if err := verifyDockerComposeInstallation(); err != nil {
		return
	}

	LogInfo("Starting PostgreSQL container...")
	if err := startDockerContainer(); err != nil {
		return
	}

	time.Sleep(3 * time.Second) // Wait for the container to start, may need to be adjusted

	// Check if the container is running
	cmd := exec.Command("sudo", "docker", "ps", "-q", "--filter", "name=pgsql_db")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		fmt.Println("PostgreSQL container is running.")
		LogInfo("PostgreSQL container is running.")
	} else if err == nil {
		fmt.Println("Waiting for postgres container...")
		time.Sleep(5 * time.Second)
		output, err = cmd.Output()
		if err == nil && len(output) > 0 {
			fmt.Println("PostgreSQL container is running.")
			LogInfo("PostgreSQL container is running.")
		} else {
			fmt.Println("PostgreSQL container is not running.")
			LogError("PostgreSQL container is not running.", err)
			return
		}
	} else {
		fmt.Println("Error checking PostgreSQL container status.")
		LogError("Error checking PostgreSQL container status.", err)
		return
	}
	containerName := "pgsql_db"
	LogInfo("Updating pg_hba.conf inside the container...")
	if err2 := UpdatePgHbaConf(containerName); err2 != nil {
		return
	}

	LogInfo("Reloading PostgreSQL configuration...")
	if err := reloadPostgresConfiguration(containerName); err != nil {
		return
	}

	LogInfo("Database container setup completed successfully.")
}

func Setup() {

	if !CheckEnvFile() {
		SetEnv()
		installOpenSSL()
		BuildContainers()
	} else {
		fmt.Println("Exising installation detected. Skipping initial configuration setup.")

	}

}
