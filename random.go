package userd

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// EncryptEachAndArchive does stuff
func EncryptEachAndArchive() error {
	fileData1 := [][]string{{"key1", "key2"}, {"value1", "value2"}}
	fileData2 := [][]string{{"key1", "key2"}, {"value1", "value2"}}
	var b1 bytes.Buffer
	csvWriter1 := csv.NewWriter(&b1)
	for _, r := range fileData1 {
		csvWriter1.Write(r)
	}
	csvWriter1.Flush()

	var b2 bytes.Buffer
	csvWriter2 := csv.NewWriter(&b2)
	for _, r := range fileData2 {
		csvWriter2.Write(r)
	}
	csvWriter2.Flush()

	fzip, _ := os.Create("archive.zip")
	defer fzip.Close()

	fzipWriter := zip.NewWriter(fzip)
	defer fzipWriter.Close()

	var files = []struct {
		Name string
		Body bytes.Buffer
	}{
		{"fileData1", b1},
		{"fileData2", b2},
	}

	for _, file := range files {
		if f, err := fzipWriter.Create(file.Name); err == nil {
			if _, err := f.Write(file.Body.Bytes()); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// ReadArchive does stuff
func ReadArchive(w io.Writer) error {
	// Open a zip archive for reading.
	r, err := zip.OpenReader("archive.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		fmt.Printf("Contents of %s:\n", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer rc.Close()
		content, err := ioutil.ReadAll(rc)
		w.Write(content)
		fmt.Println()
	}
	return nil
}
