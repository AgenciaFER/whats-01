package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	var apiURL string
	flag.StringVar(&apiURL, "api", "http://localhost:8080", "API base URL")
	flag.Parse()

	url := fmt.Sprintf("%s/sessions", apiURL)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching sessions:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("API error: %s - %s\n", resp.Status, body)
		os.Exit(1)
	}

	var sessions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		fmt.Println("Error decoding response:", err)
		os.Exit(1)
	}

	for _, s := range sessions {
		fmt.Printf("ID: %s\n", s["ID"])
		if name, ok := s["Name"]; ok {
			fmt.Printf("Name: %s\n", name)
		}
		fmt.Printf("Status: %s\n", s["Status"])
		fmt.Printf("ConnectedAt: %s\n", s["ConnectedAt"])
		fmt.Printf("LastActive: %s\n", s["LastActive"])
		if stats, ok := s["Stats"].(map[string]interface{}); ok {
			fmt.Printf("Contacts: %v, Groups: %v, Conversations: %v\n",
				stats["Contacts"], stats["Groups"], stats["Conversations"])
		}
		fmt.Println("-----")
	}
}
