package torrentfile

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

type BencodePeerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

// GET request to the announce URL supplied in the .torrent file, with a few query parameters:
func (t *TorrentData) BuildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func ParseResponse(r io.Reader) (*BencodePeerResponse, error) {
	benResponse := BencodePeerResponse{}
	err := bencode.Unmarshal(r, &benResponse)
	if err != nil {
		return nil, err
	}
	return &benResponse, nil
}

func Unmarshal(peersBin []byte) ([]Peer, error) {
	const peersize = 6 // 4 for ip, 2 for port number
	if len(peersBin)%peersize != 0 {
		return nil, fmt.Errorf("malformed peers")
	}

	numOfPeers := len(peersBin) / peersize
	peers := make([]Peer, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		offset := i * peersize
		peers[i].IP = net.IP(peersBin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16(peersBin[offset+4 : offset+6])
	}

	return peers, nil
}
