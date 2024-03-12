package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	http "net/http"
	"os"
)

type requestData struct {
	URL      string `json:"URL"`
	NAME     string `json:"name"`
	COMMENTS string `json:"comments"`
}

const (
	token = "fio-u-8cp3REGKZNw6MgK8eHvRUQXEuRpfT_tRWn2K7OOyeu0hYM39CWqTQTa_j6aHdIa-"
)

func parseRequest(r *http.Request) (requestData, error) {

	if r.Method != http.MethodPost { // Check for POST method
		//return nil, fmt.Errorf("invalid method %s", r.Method)
		return requestData{}, fmt.Errorf("invalid method %s", r.Method)
	}
	if r.Header.Get("Content-Type") != "application/json" {
		//return nil, fmt.Errorf("invalid content type %s", r.Header.Get("Content-Type"))
		return requestData{}, fmt.Errorf("invalid content type %s", r.Header.Get("Content-Type"))
	}
	defer r.Body.Close()

	// Decode into requestData struct
	var data requestData
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		//return nil, fmt.Errorf("invalid request body: %v", err)
		return requestData{}, fmt.Errorf("invalid request body: %v", err)
	}
	fmt.Println("Data:", data)
	return data, nil
}

func downloadFile(data requestData, token string) error {
	fmt.Println("Downloading file from", data.URL, "With name", data.NAME)

	if data.NAME == "" {
		data.NAME = "downloaded_file"
	}

	// Create a new HTTP client with bearer token
	client := &http.Client{}
	client.Transport = &http.Transport{
		// Optional: Configure transport properties if needed
	}
	req, err := http.NewRequest(http.MethodGet, data.URL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Download the file
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return err
	}
	defer resp.Body.Close()

	// Create the file for writing
	file, err := os.Create(data.NAME)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	// Handle potential content type for filename
	contentType := resp.Header.Get("Content-Type")
	filename := data.NAME // Default filename
	if contentType != "" {
		// Extract extension from content type (logic depends on specific format)
		// You can use libraries like "github.com/mattn/go-mimetype" for parsing
		// filename = filename + GetExtensionFromContentType(contentType)
	}

	// Download with buffered reader for potentially large files
	reader := bufio.NewReader(resp.Body)
	_, err = io.Copy(file, reader)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return err
	}

	fmt.Println("File downloaded and saved to disk:", filename)
	return nil
}

func main() {
	http.HandleFunc("/_/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		data, err := parseRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if data.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		downloadFile(data, token)

	})
	fmt.Println("Server listening on port %s")
	http.ListenAndServe(":8080", nil)
}
