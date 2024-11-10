package dmctl

import "fmt"

commands := map[string]string{
	"start": "Start a new application",
	"stop": "Stop an existing application",
	"deploy": "Deploy a new application",
	"rollout": "Rollout an existing application",
	"delete": "Delete an existing application",
	"list": "List all applications",
	"portforward": "Portforward a local port to a remote port",
}

func main() {
	fmt.Println("Hello World")
}
