// logger.go

package pst

import (
	"log"
	"os"
)

// Logger wraps the standard log.Logger and provides logging levels.
type Logger struct {
	*log.Logger
}

// NewLogger creates a new Logger instance.
func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info logs informational messages.
func (l *Logger) Info(v ...interface{}) {
	l.SetPrefix("INFO: ")
	l.Println(v...)
}

// Error logs error messages.
func (l *Logger) Error(v ...interface{}) {
	l.SetPrefix("ERROR: ")
	l.Println(v...)
}

// Debug logs debug messages.
func (l *Logger) Debug(v ...interface{}) {
	l.SetPrefix("DEBUG: ")
	l.Println(v...)
}"

"// utils.go

package pst

import (
	"io"
	"net/http"
	"os"
)

// copyHeader copies headers from source to destination.
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// copyAndClose copies from src to dst until either EOF is reached
// on src or an error occurs. It then closes the src.
func copyAndClose(dst io.Writer, src io.ReadCloser) {
	defer src.Close()
	io.Copy(dst, src)
}

// downloadFile downloads a file from the specified URL and saves it to the given path.
func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}"

"// httpclient.go

package pst

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Constants related to the HTTP client.
const (
	tcpKeepAliveInterval = 30 * time.Second
	maxRedirectsCount    = 10
	userAgentHeader      = "User-Agent"
	defaultUserAgent     = "Mozilla/5.0 (compatible; MSIE 7.0; Windows NT 6.0)"
)

// Config holds the configuration parameters for the HTTP client.
type Config struct {
	ProxyAddr    string
	UserAgent    string
	AuthType     int
	Username     string
	Password     string
	IgnoreCert   bool
	Timeout      int
	MaxRedirects int
}

// makeHTTPClient initializes and returns a new HTTP client configured with a proxy and custom dialer.
func makeHTTPClient(cfg *Config) *http.Client {
	proxyURL, err := url.Parse(cfg.ProxyAddr)
	if err != nil {
		log.Fatalf("Error parsing proxy URL: %v", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			KeepAlive: tcpKeepAliveInterval,
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.IgnoreCert},
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	}

	if cfg.MaxRedirects <= 0 {
		transport.DisableKeepAlives = true
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= cfg.MaxRedirects {
				return http.ErrUseLastResponse
			}
			// Copy the headers from the initial request to the next request.
			if len(via) > 0 {
				for attr, val := range via[0].Header {
					req.Header[attr] = val
				}
			}
			return nil
		},
	}

	// Add a default User-Agent if one is not provided.
	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent
	}

	return client
}"

"// auth.go

package pst

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Constants related to authentication.
const (
	unknownAuthType int = iota
	basicAuthType
	digestAuthType

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