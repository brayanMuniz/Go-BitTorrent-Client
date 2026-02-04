package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

// NOTE: sha1 hashes are 20 bytes long

const debianInfoHash = "86F635034839F1EBE81AB96BEE4AC59F61DB9DDE"

type TorrentData struct {
	Announce    string
	InfoHash    [20]byte   // SHA-1 hash derived from the "info" section of the torrent file
	PieceHashes [][20]byte // can be n many pieces all of length 20
	PieceLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"` // sha-1 hashes all put together
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments. ")
		fmt.Println("Run it as: torrent-client torrent-file output-file")
		return
	}

	torrentFile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Could not open torrent file")
		return
	}
	defer torrentFile.Close()

	bencodeInfo, err := rawFileToBen(torrentFile)
	if err != nil {
		return
	}

	torrentData, err := bencodeInfo.toTorrentData()
	if err != nil {
		return
	}

	encoded := hex.EncodeToString(torrentData.InfoHash[:])
	if strings.ToUpper(encoded) == debianInfoHash {
		fmt.Println("Info hashes match")
	} else {
		fmt.Println("The info hash does not match")
	}

}

func rawFileToBen(r io.Reader) (*bencodeTorrent, error) {
	benToTorrent := bencodeTorrent{}
	err := bencode.Unmarshal(r, &benToTorrent)
	if err != nil {
		return nil, err
	}
	return &benToTorrent, nil
}

func (bt *bencodeTorrent) toTorrentData() (TorrentData, error) {
	// InfoHash
	var buffer bytes.Buffer
	err := bencode.Marshal(&buffer, bt.Info)
	if err != nil {
		return TorrentData{}, fmt.Errorf("Not able to marshal bencodeInfo")
	}
	infoHash := sha1.Sum(buffer.Bytes())

	// Get Pieces
	if len(bt.Info.Pieces)%20 != 0 {
		return TorrentData{}, fmt.Errorf("Pieces are not 20 lenght long")
	}

	piecesList := [][20]byte{}
	for i := 0; i < len(bt.Info.Pieces); i += 20 {
		end := i + 20
		hashByte := []byte(bt.Info.Pieces[i:end])
		piecesList = append(piecesList, [20]byte(hashByte))
	}

	return TorrentData{
		Announce:    bt.Announce,
		InfoHash:    infoHash,
		PieceHashes: piecesList,
		PieceLength: bt.Info.PieceLength,
		Name:        bt.Info.Name,
	}, nil
}

// GET request to the announce URL supplied in the .torrent file, with a few query parameters:
