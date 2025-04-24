package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	connStr := "user=marwan dbname=wachtwoorden sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Fout bij databaseverbinding:", err)
		return
	}
	defer db.Close()

	fmt.Println("Wachtwoord Generator")

	var length int
	fmt.Print("Hoeveel karakters moet het wachtwoord zijn?: ")
	fmt.Scanln(&length)
	if length <= 0 {
		fmt.Println("Lengte moet groter zijn dan 0.")
		return
	} else if length > 64 {
		fmt.Println("Lengte moet kleiner zijn dan 64.")
		return
	}

	useUpper := vraagJaNee("Hoofdletters gebruiken? (ja/nee): ")
	useCijfers := vraagJaNee("Cijfers gebruiken? (ja/nee): ")
	useSymbolen := vraagJaNee("Symbolen gebruiken? (ja/nee): ")

	var wachtwoord string
	for {
		wachtwoord = maakWachtwoord(length, useUpper, useCijfers, useSymbolen)
		bestaat, _ := wachtwoordBestaat(wachtwoord)
		if !bestaat {
			break
		}
	}

	err = slaWachtwoordOp(wachtwoord)
	if err != nil {
		fmt.Println("Fout bij opslaan:", err)
		return
	}

	fmt.Println("Je nieuwe wachtwoord is:", wachtwoord)
}

func vraagJaNee(vraag string) bool {
	var antwoord string
	for {
		fmt.Print(vraag)
		fmt.Scanln(&antwoord)
		antwoord = strings.ToLower(strings.TrimSpace(antwoord))
		if antwoord == "ja" || antwoord == "j" {
			return true
		} else if antwoord == "nee" || antwoord == "n" {
			return false
		} else {
			fmt.Println("Antwoord 'ja' of 'nee'.")
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
		index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(tekens))))
		wachtwoord += string(tekens[index.Int64()])
	}
	return wachtwoord
}

func wachtwoordBestaat(wachtwoord string) (bool, error) {
	var id int
	err := db.QueryRow("SELECT id FROM passwords WHERE password = $1", wachtwoord).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func slaWachtwoordOp(wachtwoord string) error {
	_, err := db.Exec("INSERT INTO passwords (password) VALUES ($1)", wachtwoord)
	return err
}
