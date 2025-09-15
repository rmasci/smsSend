package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Helper to compute SHA256 hash of body and then base64 encode
func computeSHA256Base64(body []byte) string {
	h := sha256.New()
	h.Write(body)
	sum := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(sum)
}

// Helper to compute the authorization header
func computeAuthorizationHeader(accessKey, method, requestPath, query, date, contentHash string) (string, error) {
	// According to docs, the string to sign is something like:
	//   METHOD + "\n" + requestPath + "\n" + query + "\n" + date + "\n" + contentHash
	// (May also include other signed headers, but minimally these).
	// Then compute HMAC-SHA256 of this using the access key, base64 encode.
	stringToSign := strings.Join([]string{
		method,
		requestPath,
		query,
		date,
		contentHash,
	}, "\n")

	keyBytes, err := base64.StdEncoding.DecodeString(accessKey)
	if err != nil {
		return "", fmt.Errorf("invalid access key encoding: %w", err)
	}

	mac := hmac.New(sha256.New, keyBytes)
	_, err = mac.Write([]byte(stringToSign))
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// The Authorization header format per docs:
	// "HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature=<signature>"
	// (or similar; make sure signed headers list matches what you've included)
	authHeader := fmt.Sprintf(
		"HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature=%s",
		signature,
	)
	return authHeader, nil
}

func main() {
	// === Replace these ===
	endpoint := "https://<RESOURCE_NAME>.communication.azure.com"
	accessKey := "<YOUR_BASE64_ENCODED_ACCESS_KEY>" // From Azure portal
	from := "+1YOUR_ACS_NUMBER"
	to := "+1DESTINATION_NUMBER"
	message := "Hello from Go via ACS REST"
	// =====================

	apiVersion := "2021-03-07"
	path := "/sms"
	query := fmt.Sprintf("api-version=%s", apiVersion)
	url := fmt.Sprintf("%s%s?%s", endpoint, path, query)

	// Build the request body
	payload := map[string]interface{}{
		"from":    from,
		"message": message,
		"smsRecipients": []map[string]string{
			{"to": to},
		},
		"smsSendOptions": map[string]bool{
			"enableDeliveryReport": true,
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	// Compute required headers
	date := time.Now().UTC().Format(http.TimeFormat) // RFC1123 format
	contentHash := computeSHA256Base64(bodyBytes)

	// Parse out host for signed headers if needed
	// For host header, e.g. "<RESOURCE_NAME>.communication.azure.com"
	host := strings.TrimPrefix(endpoint, "https://")
	host = strings.TrimPrefix(host, "http://")
	// (If endpoint has path, strip that, but usually it’s just the host)

	// Compute auth header
	authHeader, err := computeAuthorizationHeader(accessKey, "POST", path, query, date, contentHash)
	if err != nil {
		panic(err)
	}

	// Create the request
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		panic(err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ms-date", date)
	req.Header.Set("x-ms-content-sha256", contentHash)
	req.Header.Set("Host", host)
	req.Header.Set("Authorization", authHeader)

	// Send it
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, string(respBody))
}
