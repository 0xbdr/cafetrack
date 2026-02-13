package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

type Entry struct {
	ID      uint32 `json:"id"`
	Item    string `json:"item"`
	Price   int64  `json:"price"` // stored in dirhams?
	Created int64  `json:"created"`
}

type DB struct {
	Balance int64   `json:"balance"`
	Entries []Entry `json:"entries"`
}

type Config struct {
	PasswordHash string `json:"password_hash"`
}

func dataDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".cafetrack")
	os.MkdirAll(dir, 0755)
	return dir
}

func dbPath() string     { return filepath.Join(dataDir(), "db.json") }
func logPath() string    { return filepath.Join(dataDir(), "log.txt") }
func configPath() string { return filepath.Join(dataDir(), "config.json") }

func loadDB() DB {
	var db DB
	data, err := os.ReadFile(dbPath())
	if err != nil {
		return DB{}
	}
	json.Unmarshal(data, &db)
	return db
}

func saveDB(db DB) {
	data, _ := json.MarshalIndent(db, "", "  ")
	os.WriteFile(dbPath(), data, 0644)
}

func logAction(s string) {
	f, _ := os.OpenFile(logPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(time.Now().Format(time.RFC3339) + " " + s + "\n")
}

func nextID(entries []Entry) uint32 {
	var max uint32
	for _, e := range entries {
		if e.ID > max {
			max = e.ID
		}
	}
	return max + 1
}

func centsFromString(s string) int64 {
	f, _ := strconv.ParseFloat(s, 64)
	return int64(f * 100)
}

func formatCents(c int64) string {
	return fmt.Sprintf("%.2f", float64(c)/100)
}

func add(item, priceStr string) {
	db := loadDB()
	price := centsFromString(priceStr)

	id := nextID(db.Entries)

	db.Entries = append(db.Entries, Entry{
		ID:      id,
		Item:    item,
		Price:   price,
		Created: time.Now().Unix(),
	})

	saveDB(db)
	logAction(fmt.Sprintf("ADD id=%d price=%s", id, priceStr))
	fmt.Println("ID:", id)
}

func listUnpaid() {
	db := loadDB()
	for _, e := range db.Entries {
		fmt.Printf("%d | %-20s | %s\n", e.ID, e.Item, formatCents(e.Price))
	}
}

func showBalance() {
	db := loadDB()
	fmt.Println("Balance:", formatCents(db.Balance))
}

func payID(idStr string) {
	db := loadDB()
	id64, _ := strconv.ParseUint(idStr, 10, 32)
	id := uint32(id64)

	var newEntries []Entry
	for _, e := range db.Entries {
		if e.ID != id {
			newEntries = append(newEntries, e)
		} else {
			logAction(fmt.Sprintf("PAY_ID id=%d", id))
		}
	}

	db.Entries = newEntries
	saveDB(db)
}

func payPartial(priceStr string) {
	db := loadDB()
	payment := centsFromString(priceStr)
	db.Balance += payment

	var paid []uint32
	var remaining []Entry

	for _, e := range db.Entries {
		if db.Balance >= e.Price {
			db.Balance -= e.Price
			paid = append(paid, e.ID)
		} else {
			remaining = append(remaining, e)
		}
	}

	db.Entries = remaining
	saveDB(db)

	logAction(fmt.Sprintf("PAY_PARTIAL amount=%s", priceStr))

	fmt.Println("Paid IDs:", paid)
	fmt.Println("Remaining Balance:", formatCents(db.Balance))
}

func passwdSet() {
	fmt.Print("New password: ")
	var p string
	fmt.Scanln(&p)

	hash := sha256.Sum256([]byte(p))
	cfg := Config{PasswordHash: hex.EncodeToString(hash[:])}

	data, _ := json.Marshal(cfg)
	os.WriteFile(configPath(), data, 0600)
	fmt.Println("Password set.")
}

func wipe() {
	cfgData, err := os.ReadFile(configPath())
	if err != nil {
		fmt.Println("No password set.")
		return
	}

	var cfg Config
	json.Unmarshal(cfgData, &cfg)

	fmt.Print("Password: ")
	var p string
	fmt.Scanln(&p)

	hash := sha256.Sum256([]byte(p))
	if hex.EncodeToString(hash[:]) != cfg.PasswordHash {
		fmt.Println("Wrong password.")
		return
	}

	os.RemoveAll(dataDir())
	fmt.Println("Wiped.")
}

func showLog() {
	cmd := exec.Command("less", logPath())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func help() {
	fmt.Println(`
cafetrack add "item" 25.00
cafetrack listunpaid
cafetrack balance
cafetrack pay <id>
cafetrack pay -p <price>
cafetrack log
cafetrack wipe
cafetrack passwd
`)
}

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}

	switch os.Args[1] {

	case "add":
		if len(os.Args) < 4 {
			help()
			return
		}
		add(os.Args[2], os.Args[3])

	case "listunpaid":
		listUnpaid()

	case "balance":
		showBalance()

	case "pay":
		if len(os.Args) < 3 {
			return
		}
		if os.Args[2] == "-p" && len(os.Args) == 4 {
			payPartial(os.Args[3])
		} else {
			payID(os.Args[2])
		}

	case "log":
		showLog()

	case "wipe":
		wipe()

	case "passwd":
		passwdSet()

	case "help":
		help()
	}
}
