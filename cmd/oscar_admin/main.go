// oscar-admin manages Open OSCAR Server through its management API.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultAPIURL = "http://127.0.0.1:8080"
const usage = "usage: oscar-admin set-password [-url URL] SCREEN_NAME [PASSWORD]"

func main() {
	client := &http.Client{Timeout: 10 * time.Second}
	if err := run(os.Args[1:], os.Stdin, os.Stdout, client); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer, client *http.Client) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		_, err := fmt.Fprintln(stdout, usage)
		return err
	}
	if args[0] != "set-password" {
		return fmt.Errorf("unknown command %q", args[0])
	}

	flags := flag.NewFlagSet("set-password", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	apiURL := flags.String("url", envOrDefault("OSCAR_API_URL", defaultAPIURL), "management API URL")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() < 1 || flags.NArg() > 2 {
		return errors.New(usage)
	}

	password := flags.Arg(1)
	if password == "" {
		passwordBytes, err := io.ReadAll(io.LimitReader(stdin, 1025))
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		if len(passwordBytes) > 1024 {
			return errors.New("password input is too long")
		}
		password = strings.TrimSuffix(strings.TrimSuffix(string(passwordBytes), "\n"), "\r")
	}
	if password == "" {
		return errors.New("password is required")
	}

	body, err := json.Marshal(map[string]string{
		"screen_name": flags.Arg(0),
		"password":    password,
	})
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPut, strings.TrimRight(*apiURL, "/")+"/user/password", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("update password: %s: %s", resp.Status, strings.TrimSpace(string(message)))
	}

	_, err = fmt.Fprintf(stdout, "password updated for %s\n", flags.Arg(0))
	return err
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
