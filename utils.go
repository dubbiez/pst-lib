// utils.go

package pstlib

import (
	"fmt"
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
func copyAndClose(dst io.Writer, src io.ReadCloser) error {
	_, err := io.Copy(dst, src)
	if err != nil {
		src.Close() // Close the source if an error occurred during copy.
		return fmt.Errorf("error while copying data: %w", err)
	}
	return src.Close()
}

// downloadFile downloads a file from the specified URL and saves it to the given path.
func downloadFile(filepath string, url string, client *http.Client) error {
	// Get the data using the provided http.Client
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-successful response status codes.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned non-2xx status: %d %s", resp.StatusCode, resp.Status)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}