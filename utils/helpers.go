package utils

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"net"
	"net/netip"

	"github.com/sqlc-dev/pqtype"
)

func GetIpAddr(clientIP string) pqtype.Inet {
	if clientIP == "::1" {
		clientIP = "127.0.0.1"
	}

	ip := net.ParseIP(clientIP)

	if ip == nil {
		// Handle the case where ctx.ClientIP() doesn't return a valid IP address
		return pqtype.Inet{}
	}

	inet := pqtype.Inet{
		IPNet: net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(32, 32), // If you're dealing with IPv4 addresses
		},
		Valid: true,
	}
	return inet
}

func GetIpAddrForPgx(clientIP string) (*netip.Addr, error) {
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

func GenerateUniqueToken(len int) (string, error) {
	// Generate a cryptographically secure random value
	randomBytes := make([]byte, len)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(randomBytes)

	return token, nil
}
