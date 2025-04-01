package main

import (
	"bufio"
	"custom-load-balancer/backend"
	"custom-load-balancer/balancer"
	"custom-load-balancer/jwt"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// checkServer verifies if a server is running by making a request
func checkServer(url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ Server %s is not reachable: %v\n", url, err)
		return false
	}
	defer resp.Body.Close()
	fmt.Printf("✅ Server %s responded with status %d\n", url, resp.StatusCode)
	return resp.StatusCode == http.StatusOK
}

// getUserRole prompts the user to select a valid role
func getUserRole() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter role (Admin, Client, User): ")
		role, _ := reader.ReadString('\n')
		role = strings.TrimSpace(role)

		switch role {
		case "Admin", "Client", "User":
			return role
		default:
			fmt.Println("❌ Invalid role. Please enter Admin, Client, or User.")
		}
	}
}

// sendRequest sends a request to the load balancer with the JWT token
func sendRequest(token string) {
	url := "http://localhost:8080/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("❌ Failed to create request:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ Error sending request:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("✅ Load Balancer Response: %s\n", resp.Status)
}

func main() {
	// Start backend servers
	backendPorts := []string{"8081", "8082", "8083"}
	for _, port := range backendPorts {
		go backend.StartServer(port)
	}
	time.Sleep(1 * time.Second) // Wait for backends to start

	// Start load balancer
	go func() {
		fmt.Println("🚀 Starting Load Balancer on port 8080...")
		http.HandleFunc("/", balancer.ForwardRequest)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Printf("❌ Error starting Load Balancer: %v\n", err)
		}
	}()

	// Ensure load balancer is running
	fmt.Println("⏳ Waiting for Load Balancer to start...")
	loadBalancerReady := false
	for i := 0; i < 10; i++ {
		if checkServer("http://localhost:8080/") {
			loadBalancerReady = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !loadBalancerReady {
		fmt.Println("❌ Load Balancer did not start in time.")
		return
	}

	// Get user role and generate JWT
	role := getUserRole()
	token, err := jwt.GenerateJWT(role)
	if err != nil {
		fmt.Println("❌ Error generating JWT:", err)
		return
	}

	fmt.Printf("🔑 Generated JWT for %s: %s\n", role, token)

	// Send request to Load Balancer
	sendRequest(token)

	// Keep program running
	select {}
}
