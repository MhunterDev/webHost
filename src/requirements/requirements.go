package requirements

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

// Determines what package manager to use
func GetPackageManager() (string, error) {

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

// Installs the docker requirement
func InstallDocker() {
	// Load the environment variables
	godotenv.Load(".env")
	fmt.Println("Installing docker requirement...")

	os := os.Getenv("OS")
	if os == "" {
		fmt.Println("Error: OS environment variable is not set.")
		return
	}

	//Switch to determine how we are installing postgres
	var cmd *exec.Cmd
	switch os {
	case "apt":
		cmd = exec.Command("sudo", "apt", "update", "&&", "sudo", "apt", "install", "-y", "docker")
	case "dnf":
		cmd = exec.Command("sudo", "dnf", "install", "docker", "-y")
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "docker")
	default:
		fmt.Println("Unsupported package manager")
		return
	}
	cmd.Run()
}

// Installs the openssl requirement
func InstallOpenSSL() {

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
		cmd = exec.Command("sudo", "apt", "update", "&&", "sudo", "apt", "install", "-y", "openssl")
	case "yum":
		cmd = exec.Command("sudo", "dnf", "install", "-y", "openssl")
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "openssl")
	default:
		fmt.Println("Unsupported package manager")
		return
	}

	cmd.Run()
	if err := cmd.Run(); err != nil {
		fmt.Println("Error installing openssl...")
		return
	}
	fmt.Println("OpenSSL installed successfully.")
}

// Installs the docker requirement
func InstallDockerCompose() {
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
