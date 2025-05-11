package certs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

// Helper function to format DNS names for OpenSSL configuration
func formatAltNames(dns string) string {
	altNames := strings.Split(dns, ",")
	var formatted string
	for i, name := range altNames {
		formatted += fmt.Sprintf("DNS.%d = %s\n", i+1, strings.TrimSpace(name))
	}
	return formatted
}

func PromptForCerts() {

	readInput := func(prompt string) string {
		fmt.Println(prompt)
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return scanner.Text()
		}
		fmt.Println("Error reading input:", scanner.Err())
		return ""
	}

	fmt.Println("Generating SSL certificates...")
	fmt.Println("Please enter the following details:")
	envVars := map[string]string{
		"CN":       readInput("Enter the Common Name (CN) for the certificate:"),
		"O":        readInput("Enter the Organization (O) for the certificate:"),
		"OU":       readInput("Enter the Organizational Unit (OU) for the certificate:"),
		"C":        readInput("Enter the Country (C) for the certificate:"),
		"ST":       readInput("Enter the State/Province (ST) for the certificate:"),
		"L":        readInput("Enter the Locality (L) for the certificate:"),
		"EMAIL":    readInput("Enter the Email for the certificate:"),
		"KEY_SIZE": readInput("Enter the Key Size for the certificate (2048, 4096):"),
		"DNS":      readInput("DNS Alternative Names for the certificate (comma separated):"),
	}
	err := godotenv.Write(envVars, ".env")
	if err != nil {
		fmt.Println("Error writing to .env file:", err)
		return
	}
}

// Builds your web servers CA certificate
func BuildCACerts() {

	fmt.Println("Building CA certificates...")
	// Pull in the cert details from .env
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	// Retrieve certificate details from environment variables
	cn := os.Getenv("CN")
	o := os.Getenv("O")
	ou := os.Getenv("OU")
	c := os.Getenv("C")
	st := os.Getenv("ST")
	l := os.Getenv("L")
	email := os.Getenv("EMAIL")
	keySize := os.Getenv("KEY_SIZE")
	dns := os.Getenv("DNS")

	// Prepare OpenSSL command arguments

	keyFile := ".certs/private/server.key"
	certFile := ".certs/server.crt"
	csrFile := ".certs/private/server.csr"
	opensslConfig := ".certs/openssl.cnf"

	// Generate OpenSSL configuration file for SANs
	err = os.WriteFile(opensslConfig, []byte(fmt.Sprintf(`
  [ req ]
  distinguished_name = req_distinguished_name
  req_extensions = v3_req
  prompt = no
  
  [ req_distinguished_name ]
  CN = %s
  O = %s
  OU = %s
  C = %s
  ST = %s
  L = %s
  emailAddress = %s
  
  [ v3_req ]
  subjectAltName = @alt_names
  
  [ alt_names ]
  %s
  `, cn, o, ou, c, st, l, email, formatAltNames(dns))), 0644)
	if err != nil {
		fmt.Println("Error writing OpenSSL configuration file:", err)
		return
	}

	// Generate private key with specified key size
	cmd := exec.Command("openssl", "genrsa", "-out", keyFile, keySize)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error generating private key:", err)
		return
	}

	// Generate certificate signing request (CSR) with SANs
	cmd = exec.Command("openssl", "req", "-new", "-key", keyFile, "-out", csrFile, "-config", opensslConfig)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error generating CSR:", err)
		return
	}

	// Generate self-signed certificate with SANs
	cmd = exec.Command("openssl", "x509", "-req", "-days", "365", "-in", csrFile, "-signkey", keyFile, "-out", certFile, "-extensions", "v3_req", "-extfile", opensslConfig)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error generating self-signed certificate:", err)
		return
	}

	fmt.Println("SSL certificate successfully created:")
	fmt.Println("Private Key:", keyFile)
	fmt.Println("Certificate:", certFile)
}
