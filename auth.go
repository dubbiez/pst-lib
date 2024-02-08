// auth.go

package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	//"regexp"  // Unused
	//"strconv"  // Unused
	"strings"
)

// Constants related to authentication.
const (
	UnknownAuthType int = iota
	BasicAuthType
	DigestAuthType

	proxyAuthorizationHeader = "Proxy-Authorization"
	proxyAuthenticateHeader  = "Proxy-Authenticate"
)

// proxyUserAuthData holds the username and password for proxy authentication.
type proxyUserAuthData struct {
	username string
	password string
}

// digestAuthData holds the details required for digest authentication.
type digestAuthData struct {
	realm     string
	qop       string
	nonce     string
	opaque    string
	algorithm string
	cnonce    string
	nc        uint64
}

// addBasicAuthHeader adds a basic authorization header to the request.
func addBasicAuthHeader(req *http.Request, userData *proxyUserAuthData) {
	auth := userData.username + ":" + userData.password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add(proxyAuthorizationHeader, basicAuth)
}

// addDigestAuthHeader adds a digest authorization header to the request.
func addDigestAuthHeader(req *http.Request, userData *proxyUserAuthData, digestData *digestAuthData, method string, uri string) {
	ha1 := getMD5Hash(userData.username + ":" + digestData.realm + ":" + userData.password)
	ha2 := getMD5Hash(method + ":" + uri)
	nonceCount := fmt.Sprintf("%08x", digestData.nc)
	digestData.nc++
	response := getMD5Hash(ha1 + ":" + digestData.nonce + ":" + nonceCount + ":" + digestData.cnonce + ":" + digestData.qop + ":" + ha2)

	authHeader := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", qop=%s, nc=%s, cnonce="%s", response="%s", opaque="%s", algorithm=%s`,
		userData.username, digestData.realm, digestData.nonce, uri, digestData.qop, nonceCount, digestData.cnonce, response, digestData.opaque, digestData.algorithm)

	req.Header.Add(proxyAuthorizationHeader, authHeader)
}

// getDigestAuthData parses the 'Proxy-Authenticate' header to extract digest auth details.
func getDigestAuthData(h string) *digestAuthData {
	result := &digestAuthData{}
	for _, token := range strings.Split(h, ",") {
		token = strings.Trim(token, " ")
		if strings.HasPrefix(token, "nonce=") {
			result.nonce = strings.Trim(token[6:], `"`)
		} else if strings.HasPrefix(token, "realm=") {
			result.realm = strings.Trim(token[6:], `"`)
		} else if strings.HasPrefix(token, "qop=") {
			result.qop = strings.Trim(token[4:], `"`)
		} else if strings.HasPrefix(token, "opaque=") {
			result.opaque = strings.Trim(token[7:], `"`)
		} else if strings.HasPrefix(token, "algorithm=") {
			result.algorithm = strings.Trim(token[10:], `"`)
		}
	}
	result.cnonce = makeRandomString(8)
	return result
}

// getMD5Hash returns the MD5 hash of the input string.
func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// makeRandomString generates a random string of specified length.
func makeRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := 0; i < length; i++ {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}