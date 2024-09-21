package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// Function to create HMAC signature and return API headers
func createHeaders(token, secret string) (map[string]string, error) {
	// Nonce and timestamp
	nonce := uuid.New().String()
	t := time.Now().UnixNano() / int64(time.Millisecond)

	// String to sign
	stringToSign := fmt.Sprintf("%s%d%s", token, t, nonce)

	// HMAC SHA256 hash
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := h.Sum(nil)

	// Base64 encoding
	sign := base64.StdEncoding.EncodeToString(signature)

	// Build API headers
	apiHeader := make(map[string]string)
	apiHeader["Authorization"] = token
	apiHeader["Content-Type"] = "application/json"
	apiHeader["charset"] = "utf-8"
	apiHeader["t"] = fmt.Sprintf("%d", t)
	apiHeader["sign"] = sign
	apiHeader["nonce"] = nonce

	return apiHeader, nil
}

// Function to make the API request and return the response body
func callSwitchBotAPI(url string, headers map[string]string) ([]byte, error) {
	// Create a GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers to the request
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Create HTTP client and make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: API request failed with status code %d", resp.StatusCode)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	return body, nil
}

func main() {
	// Token and secret from environment variables
	token := os.Getenv("SWITCHBOT_TOKEN")
	secret := os.Getenv("SWITCHBOT_API_KEY")

	if token == "" || secret == "" {
		fmt.Println("Error: SWITCHBOT_TOKEN or SWITCHBOT_API_KEY environment variable is not set")
		return
	}

	// Create headers
	apiHeader, err := createHeaders(token, secret)
	if err != nil {
		fmt.Printf("Error creating headers: %v\n", err)
		return
	}

	// Call the first API to get devices
	apiURLDevices := "https://api.switch-bot.com/v1.1/devices"
	bodyDevices, err := callSwitchBotAPI(apiURLDevices, apiHeader)
	if err != nil {
		fmt.Printf("Error calling /devices API: %v\n", err)
		return
	}

	// Print the response from the /devices API
	fmt.Printf("Response from /devices: %s\n", string(bodyDevices))

	// Parse the devices response to get a device ID
	var devicesResponse map[string]interface{}
	err = json.Unmarshal(bodyDevices, &devicesResponse)
	if err != nil {
		fmt.Printf("Error unmarshalling /devices response: %v\n", err)
		return
	}

	// Assume the first device in the response
	devices := devicesResponse["body"].(map[string]interface{})["deviceList"].([]interface{})
	if len(devices) == 0 {
		fmt.Println("No devices found.")
		return
	}

	// Extract the first device's ID
	firstDevice := devices[0].(map[string]interface{})
	deviceID := firstDevice["deviceId"].(string)
	fmt.Printf("Using deviceId: %s\n", deviceID)

	// Call the second API to get the status of the specific device
	apiURLDeviceStatus := fmt.Sprintf("https://api.switch-bot.com/v1.1/devices/%s/status", deviceID)
	bodyDeviceStatus, err := callSwitchBotAPI(apiURLDeviceStatus, apiHeader)
	if err != nil {
		fmt.Printf("Error calling /devices/{deviceId}/status API: %v\n", err)
		return
	}

	// Print the response from the /devices/{deviceId}/status API
	fmt.Printf("Response from /devices/{deviceId}/status: %s\n", string(bodyDeviceStatus))

	// Parse the status response
	var statusResponse map[string]interface{}
	err = json.Unmarshal(bodyDeviceStatus, &statusResponse)
	if err != nil {
		fmt.Printf("Error unmarshalling /devices/{deviceId}/status response: %v\n", err)
		return
	}

	// Extract the specific fields from the response safely
	body, ok := statusResponse["body"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: response body is missing or not in expected format.")
		return
	}

	// Extract and check if fields exist before accessing them
	deviceIDStatus, ok := body["deviceId"].(string)
	if !ok {
		deviceIDStatus = "N/A"
	}
	deviceType, ok := body["deviceType"].(string)
	if !ok {
		deviceType = "N/A"
	}
	hubDeviceId, ok := body["hubDeviceId"].(string)
	if !ok {
		hubDeviceId = "N/A"
	}
	humidity, ok := body["humidity"].(float64)
	if !ok {
		humidity = 0.0
	}
	temperature, ok := body["temperature"].(float64)
	if !ok {
		temperature = 0.0
	}

	// Print the extracted fields
	fmt.Printf("Device ID: %s\n", deviceIDStatus)
	fmt.Printf("Device Type: %s\n", deviceType)
	fmt.Printf("Hub Device ID: %s\n", hubDeviceId)
	fmt.Printf("Humidity: %.2f\n", humidity)
	fmt.Printf("Temperature: %.2f°C\n", temperature)
}
