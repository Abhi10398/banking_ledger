package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type createAccountReq struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

type depositReq struct {
	Amount int64 `json:"amount"`
}

type account struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Balance int64  `json:"balance"`
}

// seeds defines the 5 sample accounts and their initial balances (in paise).
var seeds = []struct {
	name    string
	balance int64
}{
	{"Alice", 500_000},
	{"Bob", 300_000},
	{"Carol", 750_000},
	{"Dave", 200_000},
	{"Eve", 1_000_000},
}

func main() {
	base := os.Getenv("BASE_URL")
	if base == "" {
		base = "http://localhost:8080/api"
	}

	fmt.Printf("Seeding against %s\n\n", base)

	for _, s := range seeds {
		acc, err := createAccount(base, s.name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", s.name, err)
			os.Exit(1)
		}
		if err := deposit(base, acc.ID, s.balance); err != nil {
			fmt.Fprintf(os.Stderr, "deposit %s: %v\n", s.name, err)
			os.Exit(1)
		}
		fmt.Printf("  %-8s  id=%-40s  balance=%d\n", s.name, acc.ID, s.balance)
	}

	fmt.Println("\nSeed complete.")
	fmt.Println("Pass any two account IDs above to concurrency_test.py:")
	fmt.Println("  python3 scripts/concurrency_test.py <account_a_id> <account_b_id>")
}

func createAccount(base, name string) (*account, error) {
	body, _ := json.Marshal(createAccountReq{Name: name, Currency: "INR"})
	resp, err := http.Post(base+"/accounts", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var acc account
	return &acc, json.NewDecoder(resp.Body).Decode(&acc)
}

func deposit(base, id string, amount int64) error {
	body, _ := json.Marshal(depositReq{Amount: amount})
	resp, err := http.Post(base+"/accounts/"+id+"/deposit", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}
