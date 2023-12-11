package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var secret []byte

func init() {
	secretKey := os.Getenv("WEBHOOK_SECRET")
	if secretKey == "" {
		log.Fatal("WEBHOOK_SECRET environment variable not set")
	}
	secret = []byte(secretKey)
}

func verifySignature(message, providedSignature []byte) bool {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	expectedSignature := fmt.Sprintf("sha256=%x", expectedMAC)
	return hmac.Equal([]byte(expectedSignature), providedSignature)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Invalid method")
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body")
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}
	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		log.Println("Error parsing body")
		http.Error(w, "Error parsing body", http.StatusBadRequest)
		return
	}

	if timestampStr, ok := jsonData["timestamp"].(string); ok {
		if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
			t := time.Unix(timestamp, 0)
			if time.Since(t) > 5*time.Minute {
				log.Printf("Signature expired: %s", t.Format(time.RFC3339))
				http.Error(w, "Signature expired", http.StatusUnauthorized)
				return
			}
		} else {
			log.Println("Invalid timestamp format")
			http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
			return
		}
	} else {
		log.Println("Timestamp missing")
		http.Error(w, "Timestamp missing", http.StatusBadRequest)
		return
	}

	if !verifySignature(body, []byte(signature)) {
		log.Println("Invalid signature")
		http.Error(w, "Invalid signature", http.StatusForbidden)
		return
	}

	fmt.Printf("Received webhook: %s\n", string(body))

	cmd := exec.Command("/bin/bash", "./deploy.sh")
	err = cmd.Run()
	if err != nil {
		log.Printf("Deployment script failed: %s", err)
		http.Error(w, "Deployment script failed", http.StatusInternalServerError)
		return
	}
	log.Println("Deployment script succeeded")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Deployment script succeeded"))
}

func main() {
	http.HandleFunc("/webhook", handleWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
