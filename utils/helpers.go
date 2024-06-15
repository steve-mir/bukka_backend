package utils

import (
	"crypto/rand"
	"math/big"
	"net/netip"
)

func GetIpAddr(clientIP string) (*netip.Addr, error) {
	if clientIP == "::1" {
		clientIP = "127.0.0.1"
	}

	addr, err := netip.ParseAddr(clientIP)
	if err != nil {
		// Handle the error if the IP address string is not valid.
		return nil, err
	}

	return &addr, nil
}

func GetKeyForToken(config Config, isRefresh bool) string {
	var key string
	if isRefresh {
		key = config.RefreshTokenSymmetricKey
	} else {
		key = config.AccessTokenSymmetricKey
	}

	return key
}

func GenerateSecureRandomNumber(max int64) (int64, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return nBig.Int64(), nil
}