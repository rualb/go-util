package utilzip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

// UnZip extracts the contents of a zip archive and returns a map
// where the keys are filenames and the values are the file contents as []byte.
func UnZip(data []byte, maxFileSize int) (map[string][]byte, error) {

	// DoS vulnerability via decompression bomb

	if maxFileSize == 0 {
		maxFileSize = 25 * 1000 * 1000 // 25mb
	}

	if len(data) > maxFileSize {
		return nil, fmt.Errorf("error data too big %v max size %v", len(data), maxFileSize)
	}

	// Create a reader from the zip data
	reader := bytes.NewReader(data)

	// Open the zip archive
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip reader: %v", err)
	}

	// Initialize the result map
	result := make(map[string][]byte)

	// Iterate through each file in the zip archive
	for _, file := range zipReader.File {
		// Open the file
		f, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %v", file.Name, err)
		}
		defer f.Close()

		// Read the file contents
		buf := new(bytes.Buffer)

		// G110: Potential DoS vulnerability via decompression bomb

		limitedReader := io.LimitReader(f, int64(maxFileSize))
		if _, err := io.Copy(buf, limitedReader); err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file.Name, err)
		}

		// Store the file contents in the result map
		result[file.Name] = buf.Bytes()
	}

	return result, nil
}

// Zip converts a map[string][]byte to a zip archive as []byte.
func Zip(files map[string][]byte) ([]byte, error) {
	// Create a buffer to hold the zip data
	buf := new(bytes.Buffer)

	// Create a zip writer
	zipWriter := zip.NewWriter(buf)

	// Iterate over the map and add each file to the zip archive
	for name, data := range files {
		// Create a new file in the zip archive
		zipFile, err := zipWriter.Create(name)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s in zip archive: %v", name, err)
		}

		// Write the file data to the zip entry
		_, err = zipFile.Write(data)
		if err != nil {
			return nil, fmt.Errorf("failed to write data to file %s: %v", name, err)
		}
	}

	// Close the zip writer to flush any remaining data
	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %v", err)
	}

	// Return the zip data
	return buf.Bytes(), nil
}
