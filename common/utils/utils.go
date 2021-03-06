package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/lp2p/p2pvpn/log"
	mh "github.com/multiformats/go-multihash"
)

// StrToCid transform string to cid.Cid.
func StrToCid(ns string) cid.Cid {
	h, err := mh.Sum([]byte(ns), mh.SHA2_256, -1)
	if err != nil {
		panic(err)
	}

	return cid.NewCidV1(cid.Raw, h)
}

// GetPublicIP access ip api to get public IP
// If failed, will return 127.0.0.1
func GetPublicIP() string {
	response, err := http.Get("https://api-ipv4.ip.sb/ip")
	if err != nil {
		return "127.0.0.1"
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("%v", err)
		}
	}(response.Body)

	body, _ := io.ReadAll(response.Body)
	ip := string(body)
	ip = strings.TrimRight(ip, "\n")
	return ip
}

func Md5(secret string) string {
	hash := md5.New()
	hash.Write([]byte(secret))
	return hex.EncodeToString(hash.Sum(nil))
}
