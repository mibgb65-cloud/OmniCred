package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type credential struct {
	ID       int64  `json:"id,omitempty"`
	Provider string `json:"provider"`
	Account  string `json:"account"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type credentialList struct {
	Items []credential `json:"items"`
}

func main() {
	base := flag.String("base", "http://127.0.0.1:8787", "running OmniCred base URL")
	flag.Parse()
	if err := check(strings.TrimRight(*base, "/")); err != nil {
		fmt.Fprintln(os.Stderr, "end-to-end check failed:", err)
		os.Exit(1)
	}
	fmt.Println("end-to-end check passed: desktop API and credential CRUD are working")
}

func check(base string) error {
	health, err := http.Get(base + "/healthz")
	if err != nil {
		return fmt.Errorf("check API health: %w", err)
	}
	defer health.Body.Close()
	if health.StatusCode != http.StatusOK {
		return fmt.Errorf("health response status %d", health.StatusCode)
	}

	preflight, err := http.NewRequest(http.MethodOptions, base+"/api/v1/credentials", nil)
	if err != nil {
		return err
	}
	preflight.Header.Set("Origin", "http://wails.localhost")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflightResponse, err := http.DefaultClient.Do(preflight)
	if err != nil {
		return err
	}
	defer preflightResponse.Body.Close()
	if preflightResponse.StatusCode != http.StatusNoContent || preflightResponse.Header.Get("Access-Control-Allow-Origin") != "http://wails.localhost" {
		return fmt.Errorf("desktop CORS preflight was rejected")
	}

	created := credential{}
	input := credential{Provider: "github", Account: "e2e@example.com", Username: "e2e-user", Password: "test-password-do-not-use"}
	if err := jsonRequest(http.MethodPost, base+"/api/v1/credentials", input, http.StatusCreated, &created); err != nil {
		return err
	}
	itemURL := fmt.Sprintf("%s/api/v1/credentials/%d", base, created.ID)
	cleanupNeeded := true
	defer func() {
		if cleanupNeeded {
			_ = jsonRequest(http.MethodDelete, itemURL, nil, http.StatusNoContent, nil)
		}
	}()
	if created.ID <= 0 || created.Provider != "github" {
		return fmt.Errorf("create returned unexpected record")
	}

	list := credentialList{}
	if err := jsonRequest(http.MethodGet, base+"/api/v1/credentials?provider=github&q=e2e-user", nil, http.StatusOK, &list); err != nil {
		return err
	}
	if len(list.Items) != 1 || list.Items[0].ID != created.ID {
		return fmt.Errorf("filtered list did not return created record")
	}

	updated := credential{}
	input = credential{Provider: "google", Account: "updated@example.com", Username: "updated-user", Password: "updated-test-password"}
	if err := jsonRequest(http.MethodPut, itemURL, input, http.StatusOK, &updated); err != nil {
		return err
	}
	if updated.Provider != "google" || updated.Account != "updated@example.com" {
		return fmt.Errorf("update returned unexpected record")
	}
	if err := jsonRequest(http.MethodDelete, itemURL, nil, http.StatusNoContent, nil); err != nil {
		return err
	}
	cleanupNeeded = false
	return jsonRequest(http.MethodGet, itemURL, nil, http.StatusNotFound, nil)
}

func jsonRequest(method, url string, input any, expected int, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != expected {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("%s %s returned %d: %s", method, url, response.StatusCode, message)
	}
	if output != nil {
		if err := json.NewDecoder(response.Body).Decode(output); err != nil {
			return err
		}
	}
	return nil
}
