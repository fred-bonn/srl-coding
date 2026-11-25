package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultAddr = "http://localhost:8080"

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type startRequest struct {
	ChallengeID string `json:"challenge_id,omitempty"`
	TopicID     string `json:"topic_id,omitempty"`
}

type saveCodeRequest struct {
	Code string `json:"code"`
}

type completeRequest struct {
	Correct         bool  `json:"correct"`
	DurationSeconds int32 `json:"duration_seconds"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "register":
		err = register(os.Args[2:])
	case "login":
		err = login(os.Args[2:])
	case "today":
		err = call(mustToken(), "GET", "/today", nil)
	case "start":
		err = start(os.Args[2:])
	case "save":
		err = save(os.Args[2:])
	case "complete":
		err = complete(os.Args[2:])
	case "lesson":
		err = call(mustToken(), "POST", "/lessons/complete", nil)
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: cli <command> [args]

commands:
  register <email> <password>
  login <email> <password>
  today
  start -topic <id> | -challenge <id>
  save <attempt_id> <file>
  complete <attempt_id> [true|false] [duration_seconds]
  lesson

environment:
  SRL_ADDR    server address (default `+defaultAddr+`)
  SRL_SESSION session token file (default ~/.config/srlcode/session)
`)
}

func register(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: cli register <email> <password>")
	}
	body, err := json.Marshal(loginRequest{Email: args[0], Password: args[1]})
	if err != nil {
		return err
	}
	return call("", "POST", "/auth/register", body)
}

func login(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: cli login <email> <password>")
	}
	body, err := json.Marshal(loginRequest{Email: args[0], Password: args[1]})
	if err != nil {
		return err
	}
	raw, err := exchange("", "POST", "/auth/login", body)
	if err != nil {
		return err
	}
	var res struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return fmt.Errorf("parsing login response: %w", err)
	}
	if err := writeToken(res.Token); err != nil {
		return err
	}
	fmt.Printf("logged in, session saved to %s\n", sessionPath())
	return nil
}

func start(args []string) error {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	topic := fs.String("topic", "", "topic ID to review")
	challenge := fs.String("challenge", "", "challenge ID to attempt")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if (*topic == "") == (*challenge == "") {
		return fmt.Errorf("exactly one of -topic or -challenge is required")
	}
	body, err := json.Marshal(startRequest{ChallengeID: *challenge, TopicID: *topic})
	if err != nil {
		return err
	}
	return call(mustToken(), "POST", "/attempts", body)
}

func save(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: cli save <attempt_id> <file>")
	}
	code, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	body, err := json.Marshal(saveCodeRequest{Code: string(code)})
	if err != nil {
		return err
	}
	return call(mustToken(), "PUT", "/attempts/"+args[0]+"/code", body)
}

func complete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: cli complete <attempt_id> [true|false] [duration_seconds]")
	}
	correct := true
	if len(args) >= 2 {
		var err error
		correct, err = strconv.ParseBool(args[1])
		if err != nil {
			return fmt.Errorf("invalid correct value %q", args[1])
		}
	}
	duration := int32(0)
	if len(args) >= 3 {
		parsed, err := strconv.ParseInt(args[2], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid duration %q", args[2])
		}
		duration = int32(parsed)
	}
	body, err := json.Marshal(completeRequest{Correct: correct, DurationSeconds: duration})
	if err != nil {
		return err
	}
	return call(mustToken(), "POST", "/attempts/"+args[0]+"/complete", body)
}

func call(token, method, path string, body []byte) error {
	raw, err := exchange(token, method, path, body)
	if err != nil {
		return err
	}
	if len(raw) > 0 {
		var indented bytes.Buffer
		if json.Indent(&indented, raw, "", "  ") == nil {
			fmt.Println(indented.String())
		} else {
			fmt.Println(string(raw))
		}
	}
	return nil
}

func exchange(token, method, path string, body []byte) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, addr()+path, reader)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("%s: %s", res.Status, strings.TrimSpace(string(raw)))
	}
	return raw, nil
}

func addr() string {
	if a := os.Getenv("SRL_ADDR"); a != "" {
		return a
	}
	return defaultAddr
}

func sessionPath() string {
	if p := os.Getenv("SRL_SESSION"); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "srl-session"
	}
	return filepath.Join(dir, "srlcode", "session")
}

func writeToken(token string) error {
	if err := os.MkdirAll(filepath.Dir(sessionPath()), 0o700); err != nil {
		return err
	}
	return os.WriteFile(sessionPath(), []byte(token), 0o600)
}

func mustToken() string {
	raw, err := os.ReadFile(sessionPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: not logged in; run: cli login <email> <password>")
		os.Exit(1)
	}
	return strings.TrimSpace(string(raw))
}
