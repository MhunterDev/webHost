package requirements

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Creates directories for the web certificates, docker-compose, and the database
func CreateDirs() {
	// Create directories for certificates
	fmt.Println("Creating required directories...")

	err := os.MkdirAll(".certs/private", 0755)
	if err != nil {
		fmt.Println("Error creating base directories")
		return
	}
	err2 := os.MkdirAll(".certs/.postgres/certs/private", 0755)
	if err2 != nil {
		fmt.Println("Error creating base directories")
		return
	}
	err3 := os.MkdirAll(".docker/compose/", 0755)
	if err3 != nil {
		fmt.Println("Error creating base directories")
		return
	}
}

// Determine what package manager to use
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
				return "yum", nil
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
	fmt.Println("Installing docker requirement...")

	os, _ := GetPackageManager()

	//Switch to determine how we are installing postgres
	var cmd *exec.Cmd
	switch os {
	case "apt":
		cmd = exec.Command("sudo", "apt", "update", "&&", "sudo", "apt", "install", "-y", "docker")
	case "yum":
		cmd = exec.Command("sudo", "yum", "install", "-y", "docker")
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
	fmt.Println("Installing openssl requirement...")
	mgr, err := GetPackageManager()
	if err != nil {
		fmt.Println("Error determining OS")
	}

	// Start the database install
	fmt.Println("Installing openssl requirement")

	// Placeholder variable for the cmg
	var cmd *exec.Cmd

	switch mgr {
	case "apt":
		cmd = exec.Command("sudo", "apt", "update", "&&", "sudo", "apt", "install", "-y", "openssl")
	case "yum":
		cmd = exec.Command("sudo", "yum", "install", "-y", "openssl")
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "openssl")
	default:
		fmt.Println("Unsupported package manager")
		return
	}

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error installing openssl...")
	}

}

// Installs the docker requirement
func InstallDockerCompose() {
	fmt.Println("Installing docker-compose requirement...")

	os, _ := GetPackageManager()

	//Switch to determine how we are installing postgres
	var cmd *exec.Cmd
	switch os {
	case "apt":
		cmd = exec.Command("sudo", "apt", "update", "&&", "sudo", "apt", "install", "-y", "docker-compose")
	case "yum":
		cmd = exec.Command("sudo", "yum", "install", "-y", "docker-compose")
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "docker-compose")
	default:
		fmt.Println("Unsupported package manager")
		return
	}
	cmd.Run()
}
