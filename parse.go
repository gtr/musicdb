package main

import (
	"fmt"
	"os"
	"strings"
)

func ParseListenFlag(args []string, i int) string {
	if len(args) <= i+1 {
		fmt.Println("incorrect usage")
		os.Exit(1)
	}
	port := ":" + args[i+1]
	return port
}

func ParseEndpoint(endpoint string) string {
	hostname := "localhost"
	port := ""
	if strings.Contains(endpoint, ":") {
		arr := strings.Split(endpoint, ":")
		if len(arr) > 2 {
			fmt.Println("incorrect usage")
			os.Exit(1)
		}
		if arr[0] != "" {
			hostname = arr[0]
		}
		port = arr[1]
	} else {
		fmt.Println("incorrect usage")
		os.Exit(1)
	}

	return hostname + ":" + port
}

func ParseBackendEndpointsFlag(args []string, i int) []string {
	endpoints := []string{}

	// If the flag doesn't contain an input, exit.
	if len(args) == i+1 {
		fmt.Println("incorrect usage")
		os.Exit(1)
	}
	input := args[i+1]

	// Check if there is a seperated list of backend endpoints.
	if strings.Contains(input, ",") {
		arr := strings.Split(input, ",")

		for _, endpoint := range arr {
			parsedEndpoint := ParseEndpoint(endpoint)
			endpoints = append(endpoints, parsedEndpoint)
		}

	} else {
		parsedEndpoint := ParseEndpoint(input)
		endpoints = append(endpoints, parsedEndpoint)
	}

	return endpoints
}
