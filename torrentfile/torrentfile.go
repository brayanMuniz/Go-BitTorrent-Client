package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io"
)

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

func Open(r io.Reader) (*bencodeTorrent, error) {
	benToTorrent := bencodeTorrent{}
	err := bencode.Unmarshal(r, &benToTorrent)
	if err != nil {
		return nil, err
	}
	return &benToTorrent, nil
}

func (bt *bencodeTorrent) ToTorrentData() (TorrentData, error) {
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
