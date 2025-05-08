package docker

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

func ConfigureDocker() {

	conf := `
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
	cnf, err := os.Create(".docker/compose/docker-compose.yml")
	if err != nil {
		fmt.Println("Error creating docker-compose.yml file:", err)
		return
	}
	defer cnf.Close()

	_, err = cnf.Write([]byte(conf))
	if err != nil {
		fmt.Println("Error writing to docker-compose.yml file:", err)
		return
	}

	fmt.Println("Docker configuration file created successfully.")
}

func ConfigureDockerEnv() {
	godotenv.Load(".env")
	user := os.Getenv("POST_USER")
	password := os.Getenv("POST_PASSWORD")
	dbname := os.Getenv("DBNAME")
	dockerip := os.Getenv("DOCKER_IP")
	dbport := os.Getenv("PG_PORT")
	subnet := os.Getenv("SUBNET")
	conf := fmt.Sprintf("POST_USER=%s\nPOST_PASSWORD=%s\nDBNAME=%s\nPG_PORT=%s\nDOCKER_IP=%s\nSUBNET=%s\n", user, password, dbname, dbport, dockerip, subnet)

	cnf, err := os.Create(".docker/compose/.env")
	if err != nil {
		return
	}
	defer cnf.Close()
	_, err = cnf.Write([]byte(conf))
	if err != nil {
		fmt.Println("Error writing to .env file:", err)
		return

	}
	fmt.Println("docker env set successfully.")
}

func StartDocker() {

	fmt.Println("Starting Docker...")

	// Enable and start Docker service
	// This command may require sudo privileges
	cmd := "sudo systemctl enable docker && sudo systemctl start docker"
	err := exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		fmt.Println("Docker is not running. Please start Docker and try again.")
		return
	}

	// Check if Docker is running
	cmd = "docker info"
	err = exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		fmt.Println("Docker is not running. Please start Docker and try again.")
		return
	}
	fmt.Println("Docker installed successfully.")
}

func StartDB() {
	fmt.Println("Starting Postgres container...")

	// Start the Docker container using docker-compose
	cmd := "sudo docker-compose -f .docker/compose/docker-compose.yml up -d"
	err := exec.Command("bash", "-c", cmd).Run()
	if err != nil {
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
