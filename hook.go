package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	http "net/http"
	"os"
)

type downloadResponseSimple struct {
	Title       string `json:"title"`
	Description string `json:"description"` // Optional field for errors
}

type downloadResponseForm struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Fields      CustomFields `json:"fields"`
}

type CustomFields []struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Label string `json:"label"`
}
type requestData struct {
	Resource Resource `json:"resource"`
	Data     Data     `json:"data"`
	URL      string   `json:"original"`
	NAME     string   `json:"name"`
}

type Resource struct {
	ID   string `json:"id"`
	TYPE string `json:"type"`
}

type Data struct {
	downloadName string `json:"dw_name"`
}

type fileData struct {
	URL  string `json:"original"`
	NAME string `json:"name"`
}

var token string

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

func downloadFile(data fileData, token string) error {
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

func queryAsset(fields requestData, token string) (fileData, error) {
	asset_id := fields.Resource.ID

	fmt.Println("Querying asset with ID", asset_id, " and token:", token)
	// Create a new HTTP client with bearer token
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "https://api.frame.io/v2/assets/"+asset_id, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return fileData{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	//get asset details
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error getting asset details:", err)
		return fileData{}, err
	}
	defer resp.Body.Close()

	// Decode into fileData struct
	var data fileData
	err = json.NewDecoder(resp.Body).Decode(&data)

	// Read the response body
	body, err := io.ReadAll(resp.Body)

	// Print the response body as a string
	fmt.Println("Response Body:", string(body))

	return data, nil

}

func main() {
	token := os.Getenv("TOKEN") // Get the token from the environment

	if token == "" {
		token = "fio-u-8cp3REGKZNw6MgK8eHvRUQXEuRpfT_tRWn2K7OOyeu0hYM39CWqTQTa_j6aHdIa-"
	}
	http.HandleFunc("/_/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		fields, err := parseRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if fields.Data.downloadName == "" {
			w.Header().Set("Content-Type", "application/json")
			response := downloadResponseForm{
				Title:       "Download File",
				Description: "Please provide the name of the file to download",
				Fields: CustomFields{
					{Type: "text", Name: "dw_name", Value: "", Label: "File Name"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
		assetData, err := queryAsset(fields, token)

		if assetData.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		err = downloadFile(assetData, token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			successResponse := downloadResponseSimple{Title: "Yey!", Description: "File downloaded successfully"}
			json.NewEncoder(w).Encode(successResponse)
		}

	})
	fmt.Println("Server listening on port %s")
	http.ListenAndServe(":8080", nil)
}
