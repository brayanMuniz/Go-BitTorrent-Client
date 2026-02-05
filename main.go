package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/brayanMuniz/Go-BitTorrent-Client/torrentfile"
)

const debianInfoHash = "86F635034839F1EBE81AB96BEE4AC59F61DB9DDE"
const Port uint16 = 6881

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments. ")
		fmt.Println("Run it as: torrent-client torrent-file output-file")
		return
	}

	tFile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Could not open torrent file")
		return
	}
	defer tFile.Close()

	bencodeInfo, err := torrentfile.Open(tFile)
	if err != nil {
		return
	}

	torrentData, err := bencodeInfo.ToTorrentData()
	if err != nil {
		return
	}

	hexEncoded := hex.EncodeToString(torrentData.InfoHash[:])
	if strings.ToUpper(hexEncoded) == debianInfoHash {
		fmt.Println("Info hashes match")
	} else {
		fmt.Println("The info hash does not match")
	}

	randomBytes := make([]byte, 20)
	rand.Read(randomBytes)
	getRequestURL, err := torrentData.BuildTrackerURL([20]byte(randomBytes), Port)
	if err != nil {
		return
	}
	fmt.Println(getRequestURL)

	resp, err := http.Get(getRequestURL)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	benParsed, err := torrentfile.ParseResponse(resp.Body)
	if err != nil {
		fmt.Println("Not able to parse the response")
		return
	}

	peers, err := torrentfile.Unmarshal([]byte(benParsed.Peers))
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Println(peers[0].IP, peers[0].Port)
}
