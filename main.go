package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type Config struct {
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

var db *sql.DB

func main() {
	logFile, err := os.OpenFile("wachtwoord_generator.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Fout bij openen logbestand:", err)
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("Starten van wachtwoord generator...")

	config, err := laadConfig("config.json")
	if err != nil {
		log.Fatalf("Fout bij laden config: %v", err)
	}

	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=%s", config.User, config.DBName, config.SSLMode)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Fout bij databaseverbinding: %v", err)
	}
	defer db.Close()
	log.Println("Verbonden met de database")

	fmt.Println("Wachtwoord Generator")

	var length int
	fmt.Print("Hoeveel karakters moet het wachtwoord zijn?: ")
	fmt.Scanln(&length)
	log.Printf("Ingevoerde lengte: %d", length)
	if length <= 0 {
		log.Println("Lengte is ongeldig (<= 0)")
		fmt.Println("Lengte moet groter zijn dan 0.")
		return
	} else if length > 64 {
		log.Println("Lengte is ongeldig (> 64)")
		fmt.Println("Lengte moet kleiner zijn dan 64.")
		return
	}

	useUpper := vraagJaNee("Hoofdletters gebruiken? (ja/nee): ")
	useCijfers := vraagJaNee("Cijfers gebruiken? (ja/nee): ")
	useSymbolen := vraagJaNee("Symbolen gebruiken? (ja/nee): ")

	log.Printf("Instellingen - Lengte: %d, Hoofdletters: %t, Cijfers: %t, Symbolen: %t", length, useUpper, useCijfers, useSymbolen)

	var wachtwoord string
	for {
		wachtwoord = maakWachtwoord(length, useUpper, useCijfers, useSymbolen)
		bestaat, err := wachtwoordBestaat(wachtwoord)
		if err != nil {
			log.Fatalf("Fout bij controleren wachtwoord: %v", err)
		}
		if !bestaat {
			break
		}
		log.Println("Wachtwoord bestaat al, nieuw genereren...")
	}

	err = slaWachtwoordOp(wachtwoord)
	if err != nil {
		log.Fatalf("Fout bij opslaan wachtwoord: %v", err)
	}

	log.Println("Nieuw wachtwoord succesvol opgeslagen")
	fmt.Println("Je nieuwe wachtwoord is:", wachtwoord)
}

func laadConfig(pad string) (Config, error) {
	file, err := os.Open(pad)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	log.Println("Configuratie geladen uit config.json")
	return config, nil
}

func vraagJaNee(vraag string) bool {
	var antwoord string
	for {
		fmt.Print(vraag)
		fmt.Scanln(&antwoord)
		antwoord = strings.ToLower(strings.TrimSpace(antwoord))
		log.Printf("Antwoord op '%s': %s", vraag, antwoord)
		if antwoord == "ja" || antwoord == "j" {
			return true
		} else if antwoord == "nee" || antwoord == "n" {
			return false
		} else {
			fmt.Println("Antwoord 'ja' of 'nee'.")
			log.Println("Ongeldig antwoord ingevoerd.")
		}
	}
}

func maakWachtwoord(lengte int, hoofdletters, cijfers, symbolen bool) string {
	kleineLetters := "abcdefghijklmnopqrstuvwxyz"
	hoofdLetters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	getallen := "0123456789"
	specials := "!@#$%&*?"

	tekens := kleineLetters
	if hoofdletters {
		tekens += hoofdLetters
	}
	if cijfers {
		tekens += getallen
	}
	if symbolen {
		tekens += specials
	}

	wachtwoord := ""
	for i := 0; i < lengte; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(tekens))))
		if err != nil {
			log.Printf("Fout bij willekeurig getal genereren: %v", err)
			continue
		}
		wachtwoord += string(tekens[index.Int64()])
	}
	log.Printf("Wachtwoord gegenereerd: %s", wachtwoord)
	return wachtwoord
}

func wachtwoordBestaat(wachtwoord string) (bool, error) {
	var id int
	err := db.QueryRow("SELECT id FROM passwords WHERE password = $1", wachtwoord).Scan(&id)
	if err == sql.ErrNoRows {
		log.Println("Wachtwoord bestaat nog niet.")
		return false, nil
	}
	if err != nil {
		log.Printf("Fout bij controleren op bestaan wachtwoord: %v", err)
		return false, err
	}
	log.Println("Wachtwoord bestaat al in database.")
	return true, nil
}

func slaWachtwoordOp(wachtwoord string) error {
	_, err := db.Exec("INSERT INTO passwords (password) VALUES ($1)", wachtwoord)
	if err != nil {
		log.Printf("Fout bij invoegen in database: %v", err)
	} else {
		log.Printf("Wachtwoord opgeslagen: %s", wachtwoord)
	}
	return err
}
