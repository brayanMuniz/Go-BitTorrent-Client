package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
	"net/http"
	"os"
)

const Port uint16 = 6881

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

func (tf *TorrentData) DownloadToFile() error {

	// Set up params
	randomBytes := make([]byte, 20)
	rand.Read(randomBytes)
	getRequestURL, err := tf.buildTrackerURL([20]byte(randomBytes), Port)
	if err != nil {
		return err
	}

	// Get request
	resp, err := http.Get(getRequestURL)
	if err != nil {
		fmt.Println("Error making request:", err)
		return err
	}
	defer resp.Body.Close()

	// raw ben -> ben struct
	benResponse := bencodePeerResponse{}
	err = bencode.Unmarshal(resp.Body, &benResponse)
	if err != nil {
		return err
	}

	peers, err := unmarshalPeers([]byte(benResponse.Peers))
	if err != nil {
		fmt.Print(err)
		return err
	}

	fmt.Println(peers[0].IP, peers[0].Port)

	return nil
}

func Open(filePath string) (TorrentData, error) {
	// open file
	tFile, err := os.Open(filePath)
	if err != nil {
		return TorrentData{}, err
	}
	defer tFile.Close()

	// raw ben -> ben struct
	benToTorrent := bencodeTorrent{}
	err = bencode.Unmarshal(tFile, &benToTorrent)
	if err != nil {
		return TorrentData{}, err
	}

	// ben struct -> go struct
	return benToTorrent.toTorrentData()
}

func (bt *bencodeTorrent) toTorrentData() (TorrentData, error) {
	// InfoHash
	var buffer bytes.Buffer
	err := bencode.Marshal(&buffer, bt.Info)
	if err != nil {
		return TorrentData{}, fmt.Errorf("Not able to marshal bencodeInfo")
	}
	infoHash := sha1.Sum(buffer.Bytes())

	// Pieces
	if len(bt.Info.Pieces)%20 != 0 {
		return TorrentData{}, fmt.Errorf("Pieces are not 20 length long")
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
