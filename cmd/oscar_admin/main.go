// oscar-admin manages Open OSCAR Server through its management API.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultAPIURL = "http://127.0.0.1:8080"
const usage = `usage:
  oscar-admin set-password [-url URL] SCREEN_NAME [PASSWORD]
  oscar-admin user list [-url URL]
  oscar-admin user show [-url URL] SCREEN_NAME
  oscar-admin user create [-url URL] SCREEN_NAME [PASSWORD]
  oscar-admin user delete [-url URL] SCREEN_NAME
  oscar-admin session list [-url URL]
  oscar-admin session show [-url URL] SCREEN_NAME
  oscar-admin session disconnect [-url URL] SCREEN_NAME
  oscar-admin message send [-url URL] FROM TO TEXT
  oscar-admin version [-url URL]`

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

	switch args[0] {
	case "set-password":
		return setPassword(args[1:], stdin, stdout, client)
	case "user":
		return userCommand(args[1:], stdin, stdout, client)
	case "session":
		return sessionCommand(args[1:], stdout, client)
	case "message":
		return messageCommand(args[1:], stdout, client)
	case "version":
		return getCommand(args[1:], "/version", stdout, client)
	default:
		return fmt.Errorf("unknown command %q\n%s", args[0], usage)
	}
}

func userCommand(args []string, stdin io.Reader, stdout io.Writer, client *http.Client) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	switch args[0] {
	case "list":
		return getCommand(args[1:], "/user", stdout, client)
	case "show":
		apiURL, operands, err := parseArgs(args[1:], 1, 1)
		if err != nil {
			return err
		}
		response, err := request(client, apiURL, http.MethodGet, "/user/"+url.PathEscape(operands[0])+"/account", nil, http.StatusOK)
		return printResponse(stdout, response, err)
	case "create":
		apiURL, operands, err := parseArgs(args[1:], 1, 2)
		if err != nil {
			return err
		}
		password, err := passwordFromArgsOrStdin(operands[1:], stdin)
		if err != nil {
			return err
		}
		_, err = request(client, apiURL, http.MethodPost, "/user", map[string]string{"screen_name": operands[0], "password": password}, http.StatusCreated)
		if err == nil {
			_, err = fmt.Fprintf(stdout, "created user %s\n", operands[0])
		}
		return err
	case "delete":
		apiURL, operands, err := parseArgs(args[1:], 1, 1)
		if err != nil {
			return err
		}
		_, err = request(client, apiURL, http.MethodDelete, "/user", map[string]string{"screen_name": operands[0]}, http.StatusNoContent)
		if err == nil {
			_, err = fmt.Fprintf(stdout, "deleted user %s\n", operands[0])
		}
		return err
	default:
		return fmt.Errorf("unknown user command %q", args[0])
	}
}

func sessionCommand(args []string, stdout io.Writer, client *http.Client) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	switch args[0] {
	case "list":
		return getCommand(args[1:], "/session", stdout, client)
	case "show", "disconnect":
		apiURL, operands, err := parseArgs(args[1:], 1, 1)
		if err != nil {
			return err
		}
		path := "/session/" + url.PathEscape(operands[0])
		if args[0] == "show" {
			response, err := request(client, apiURL, http.MethodGet, path, nil, http.StatusOK)
			return printResponse(stdout, response, err)
		}
		_, err = request(client, apiURL, http.MethodDelete, path, nil, http.StatusNoContent)
		if err == nil {
			_, err = fmt.Fprintf(stdout, "disconnected %s\n", operands[0])
		}
		return err
	default:
		return fmt.Errorf("unknown session command %q", args[0])
	}
}

func messageCommand(args []string, stdout io.Writer, client *http.Client) error {
	if len(args) == 0 || args[0] != "send" {
		return errors.New("usage: oscar-admin message send [-url URL] FROM TO TEXT")
	}
	apiURL, operands, err := parseArgs(args[1:], 3, -1)
	if err != nil {
		return err
	}
	_, err = request(client, apiURL, http.MethodPost, "/instant-message", map[string]string{
		"from": operands[0],
		"to":   operands[1],
		"text": strings.Join(operands[2:], " "),
	}, http.StatusOK)
	if err == nil {
		_, err = fmt.Fprintln(stdout, "message sent")
	}
	return err
}

func setPassword(args []string, stdin io.Reader, stdout io.Writer, client *http.Client) error {
	apiURL, operands, err := parseArgs(args, 1, 2)
	if err != nil {
		return err
	}
	password, err := passwordFromArgsOrStdin(operands[1:], stdin)
	if err != nil {
		return err
	}
	_, err = request(client, apiURL, http.MethodPut, "/user/password", map[string]string{
		"screen_name": operands[0],
		"password":    password,
	}, http.StatusNoContent)
	if err == nil {
		_, err = fmt.Fprintf(stdout, "password updated for %s\n", operands[0])
	}
	return err
}

func getCommand(args []string, path string, stdout io.Writer, client *http.Client) error {
	apiURL, _, err := parseArgs(args, 0, 0)
	if err != nil {
		return err
	}
	response, err := request(client, apiURL, http.MethodGet, path, nil, http.StatusOK)
	return printResponse(stdout, response, err)
}

func parseArgs(args []string, minArgs, maxArgs int) (string, []string, error) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	apiURL := flags.String("url", envOrDefault("OSCAR_API_URL", defaultAPIURL), "management API URL")
	if err := flags.Parse(args); err != nil {
		return "", nil, err
	}
	if flags.NArg() < minArgs || maxArgs >= 0 && flags.NArg() > maxArgs {
		return "", nil, errors.New(usage)
	}
	return *apiURL, flags.Args(), nil
}

func passwordFromArgsOrStdin(args []string, stdin io.Reader) (string, error) {
	if len(args) == 1 && args[0] != "" {
		return args[0], nil
	}
	passwordBytes, err := io.ReadAll(io.LimitReader(stdin, 1025))
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	if len(passwordBytes) > 1024 {
		return "", errors.New("password input is too long")
	}
	password := strings.TrimSuffix(strings.TrimSuffix(string(passwordBytes), "\n"), "\r")
	if password == "" {
		return "", errors.New("password is required")
	}
	return password, nil
}

func request(client *http.Client, apiURL, method, path string, payload any, wantStatus int) ([]byte, error) {
	var body io.Reader = http.NoBody
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, strings.TrimRight(apiURL, "/")+path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	response, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != wantStatus {
		return nil, fmt.Errorf("request failed: %s: %s", resp.Status, strings.TrimSpace(string(response)))
	}
	return response, nil
}

func printResponse(stdout io.Writer, response []byte, err error) error {
	if err != nil {
		return err
	}
	if _, err := stdout.Write(response); err != nil {
		return err
	}
	if len(response) > 0 && response[len(response)-1] != '\n' {
		_, err = fmt.Fprintln(stdout)
	}
	return err
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
