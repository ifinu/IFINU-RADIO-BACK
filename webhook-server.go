package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const (
	webhookPort   = "9000"
	deployScript  = "/var/www/ifinu-radio/ifinu-radio-back/webhook-deploy.sh"
	webhookSecret = "" // Will be set via WEBHOOK_SECRET env var
)

func main() {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("WEBHOOK_SECRET environment variable must be set")
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Verify signature
		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" {
			log.Println("Missing X-Hub-Signature-256 header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !verifySignature(body, signature, secret) {
			log.Println("Invalid signature")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Println("✅ Valid webhook received, triggering deploy...")

		// Execute deploy script in background
		go func() {
			cmd := exec.Command("/bin/bash", deployScript)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("❌ Deploy failed: %v\n%s", err, output)
			} else {
				log.Printf("✅ Deploy completed:\n%s", output)
			}
		}()

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Deploy triggered")
	})

	log.Printf("🚀 Webhook server listening on port %s", webhookPort)
	log.Fatal(http.ListenAndServe(":"+webhookPort, nil))
}

func verifySignature(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
