// utils.go

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
}