// websockets.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	_ "github.com/bmizerany/pq"

	"github.com/kabukky/httpscerts"
)

// Players stores a list of all players in the file
type Players struct {
	Players []Player `json:"players"`
}

// Player stores information about the player to parse onto the webpage
type Player struct {
	State string `json:"state"`
	Rank  int    `json:"rank"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

func main() {

	// Check if the cert files are available.
	err := httpscerts.Check("cert.pem", "key.pem")
	// If they are not available, generate new ones.
	if err != nil {
		err = httpscerts.Generate("cert.pem", "key.pem", "127.0.0.1:8080")
		if err != nil {
			log.Fatal("Error: Couldn't create https certs.")
		}
	}
	// Handler points to available directories
	http.Handle("/home/", http.StripPrefix("/home/", http.FileServer(http.Dir("home"))))
	http.Handle("/home/about/", http.StripPrefix("/home/about/", http.FileServer(http.Dir("home/about"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("scripts"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path[1:] == "" {
			http.ServeFile(w, r, "home/index.html")
		} else if r.URL.Path[1:7] == "states" {
			fmt.Fprintf(w, retrievePlayers(r.URL.Path[8:]))
		} else {
			http.ServeFile(w, r, "home/"+r.URL.Path[1:]+".html")
		}
	})

	// Serve /submitPlayer with a text response.
	http.HandleFunc("/submitPlayer", func(w http.ResponseWriter, r *http.Request) {
		// Parses Form
		err := r.ParseForm()
		if err != nil {
			http.Error(w, fmt.Sprintf("error parsing url %v", err), 500)
		}
		// Extracts information passed from AJAX statement on examplesubmitPlayer.html
		id := r.FormValue("ID")
		idInt, _ := strconv.Atoi(id)
		writeToPlayer(getUserInfo(idInt))
		fmt.Fprintf(w, "Success!")
	})

	//Serves local webpage for testing
	if true {
		errhttp := http.ListenAndServe(":8080", nil)
		if errhttp != nil {
			log.Fatal("Web server (HTTP): ", errhttp)
		}
	} else {
		//Serves the webpage
		errhttps := http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil)
		if errhttps != nil {
			log.Fatal("Web server (HTTPS): ", errhttps)
		}
	}

}

func getUserInfo(id int) Player {
	//var finalPlayer Player
	//TODO: Parse user info from osu website
	// Make HTTP GET request
	response, err := http.Get("https://osu.ppy.sh/user/17785281")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Copy data from the response to standard output
	rawHTML, err := io.Copy(os.Stdout, response.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Number of bytes copied to STDOUT:", rawHTML)

	//return finalPlayer
	return Player{ID: id, Rank: 1234567, Name: "Test", State: "ohio"}
}

func writeToPlayer(newPlayer Player) {
	// Open our jsonFile
	jsonFile, err := os.Open("data/players.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var currentList Players
	json.Unmarshal(byteValue, &currentList)

	// add further value into it
	currentList.Players = append(currentList.Players, Player{ID: newPlayer.ID, Rank: newPlayer.Rank, Name: newPlayer.Name, State: newPlayer.State})

	// now Marshal it
	finalList, _ := json.Marshal(currentList)

	err = ioutil.WriteFile("data/players.json", finalList, 0644)

}

func retrievePlayers(state string) string {
	var finalString = buildHTMLHeader()
	// Open our jsonFile
	jsonFile, err := os.Open("data/players.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Players array
	var players Players

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'players' which we defined above
	json.Unmarshal(byteValue, &players)

	players = sortPlayers(players)

	finalString += `<body>
    <div class="navbar">
        <a href="/">Home</a>
    </div>

	<div class="main">
	<ol>`

	for i := 0; i < len(players.Players); i++ {
		//TODO: sort players by rank
		if players.Players[i].State == state {
			finalString += ("<li><p>Name: " + players.Players[i].Name + "</p>")
			finalString += ("<p>State: " + players.Players[i].State + "</p>")
			finalString += ("<p>Rank: " + strconv.Itoa(players.Players[i].Rank) + "</p></li>")
		}
	}

	finalString += `</ol></div></body>`

	finalString += buildHTMLFooter()

	return finalString
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func buildHTMLHeader() string {
	var finalHeader string = `<!DOCTYPE html>
	<html>
	<title>Default Homepage</title>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link rel="stylesheet" type="text/css" href="../css/main.css" />
	<link rel="stylesheet" type="text/css" href="../css/flexbox.css" />
	<link rel="stylesheet" type="text/css" href="../css/normalize.css" />`
	return finalHeader
}

func buildHTMLFooter() string {
	var finalFooter string = `</html>`
	return finalFooter
}

func sortPlayers(list Players) Players {
	sort.SliceStable(list.Players, func(i, j int) bool {
		return list.Players[i].Rank < list.Players[j].Rank
	})

	return list
}
