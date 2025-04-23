package terabee

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type GetPayload struct {
	In  int `json:"in_counts"`
	Out int `json:"out_counts"`
}

func (p *GetPayload) GetNetOcc() int {
	return p.In - p.Out
}

func GetNetOccupancy(serial string) (int, error) {
	url := fmt.Sprintf("https://terabee-%s.local/wizard/api/get_counts?format=json", serial)
	username := "people_counting_admin"
	password := fmt.Sprintf("%s--admin", serial)

	// Create a custom Transport that skips TLS verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Create an HTTP client with the custom Transport
	client := &http.Client{Transport: tr}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0, nil
	}

	// Set the basic authentication header
	req.SetBasicAuth(username, password)

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error performing request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	// Process the response
	fmt.Println("Status Code:", resp.StatusCode)
	// You can read the response body here as well
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		os.Exit(1)
	}

	fmt.Println("Response Body:", string(body))

	var payload GetPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshall payload: %s", body)
		fmt.Println(errMsg)
		fmt.Println(err.Error())
		return 0, err
	}
	return payload.GetNetOcc(), nil
}
