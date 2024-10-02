package main

import (
    "fmt"
    "golang.org/x/crypto/ssh"
    "log"
)

func main() {
    // SSH server configuration
    serverIP := "your.server.ip"    // Replace with your server IP
    username := "your_username"     // Replace with your username
    password := "your_password"     // Replace with your password

    // New user details to be created
    newUser := "new_username" // Replace with the new username

    // Configure the SSH client
    config := &ssh.ClientConfig{
        User: username,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    // Connect to the SSH server
    client, err := ssh.Dial("tcp", serverIP+":22", config)
    if err != nil {
        log.Fatalf("Failed to connect: %s", err)
    }
    defer client.Close()

    // Create a session
    session, err := client.NewSession()
    if err != nil {
        log.Fatalf("Failed to create session: %s", err)
    }
    defer session.Close()

    // Command to create a new user (may require sudo)
    createUserCmd := fmt.Sprintf("sudo useradd %s", newUser)

    // Execute the command
    output, err := session.CombinedOutput(createUserCmd)
    if err != nil {
        log.Fatalf("Failed to create user: %s", err)
    }

    // Print the output of the command
    fmt.Println(string(output))
}
