package main

import (
	"fmt"

	"github.com/onehumancorp/mono/srcs/agents"
)

func main() {
	// Initialize the Agent Registry
	registry := agents.DefaultRegistry()

	// Use the Built-in provider
	provider, ok := registry.Get(agents.ProviderTypeBuiltin)
	if !ok {
		fmt.Println("Builtin provider not found")
		return
	}

	// This is a minimal Hello World example to show how to interface with the registry
	fmt.Printf("Hello World! Successfully loaded provider: %s\n", provider.Description())
	fmt.Printf("Is Authenticated: %v\n", provider.IsAuthenticated())

	fmt.Println("\nSupported Roles for this agent:")
	for _, role := range provider.SupportedRoles() {
		fmt.Printf("- %s\n", role)
	}
}
