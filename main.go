package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/brayanMuniz/Go-BitTorrent-Client/torrentfile"
)

const debianInfoHash = "86F635034839F1EBE81AB96BEE4AC59F61DB9DDE"

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments. ")
		fmt.Println("Run it as: torrent-client torrent-file output-file")
		return
	}

	torrentData, err := torrentfile.Open(os.Args[1])
	if err != nil {
		return
	}

	// Quick unit test
	hexEncoded := hex.EncodeToString(torrentData.InfoHash[:])
	if strings.ToUpper(hexEncoded) == debianInfoHash {
		fmt.Println("Info hashes match")
	} else {
		fmt.Println("The info hash does not match")
	}

	err = torrentData.DownloadToFile()
	if err != nil {
		return
	}

}
