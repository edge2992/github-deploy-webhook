package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
)

var secret = []byte("adfgadfe123FDSa")

func varifySignature(message, providedSignature []byte) bool {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	expectedSignature := fmt.Sprintf("sha256=%x", expectedMAC)
	return hmac.Equal([]byte(expectedSignature), providedSignature)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}

	if !varifySignature(body, []byte(signature)) {
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
	// 応答を返す
}

func main() {
	http.HandleFunc("/webhook", handleWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
