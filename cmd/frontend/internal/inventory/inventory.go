// Package inventory scans a directory tree to identify the
// programming languages, etc., in use.
package inventory

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/src-d/enry/v2"
	"github.com/src-d/enry/v2/data"
)

// Inventory summarizes a tree's contents (e.g., which programming
// languages are used).
type Inventory struct {
	// Languages are the programming languages used in the tree.
	Languages []Lang `json:"Languages,omitempty"`
}

// Lang represents a programming language used in a directory tree.
type Lang struct {
	// Name is the name of a programming language (e.g., "Go" or
	// "Java").
	Name string `json:"Name,omitempty"`
	// TotalBytes is the total number of bytes of code written in the
	// programming language.
	TotalBytes uint64 `json:"TotalBytes,omitempty"`
	// TotalLines is the total number of lines of code written in the
	// programming language.
	TotalLines uint64 `json:"TotalLines,omitempty"`
}

// minFileBytes is the minimum byte size prefix for each file to read when using file contents for
// language detection.
const minFileBytes = 16 * 1024

// detect performs an inventory of the file passed in. If readFile is provided, the language
// detection uses heuristics based on the file content for greater accuracy.
func detect(ctx context.Context, file os.FileInfo, readFile func(ctx context.Context, path string, minBytes int64) ([]byte, error)) (string, error) {
	if !file.Mode().IsRegular() || enry.IsVendor(file.Name()) {
		return "", nil
	}

	// In many cases, GetLanguageByFilename can detect the language conclusively just from the
	// filename. Only try to read the file (which is much slower) if needed.
	matchedLang, safe := GetLanguageByFilename(file.Name())
	if !safe && readFile != nil {
		data, err := readFile(ctx, file.Name(), minFileBytes)
		if err != nil {
			return "", err
		}
		matchedLang = enry.GetLanguage(file.Name(), data)
	}
	return matchedLang, nil
}

func getLang(ctx context.Context, file os.FileInfo, rc io.ReadCloser) (Lang, error) {
	defer func() {
		if rc != nil {
			rc.Close()
		}
	}()

	if !file.Mode().IsRegular() || enry.IsVendor(file.Name()) {
		return Lang{}, nil
	}

	lang := Lang{}
	// In many cases, GetLanguageByFilename can detect the language conclusively just from the
	// filename. If not, we pass a subset of the file contents for analysis.
	matchedLang, safe := GetLanguageByFilename(file.Name())
	if rc != nil {
		var data []byte
		var err error
		if !safe {
			// Detect language
			r := io.LimitReader(rc, minFileBytes)
			data, err = ioutil.ReadAll(r)
			if err != nil {
				return lang, err
			}
			matchedLang = enry.GetLanguage(file.Name(), data)
		}

		// Count lines
		var linecount int
		scanner := bufio.NewScanner(io.MultiReader(bytes.NewReader(data), rc))
		for scanner.Scan() {
			linecount++
		}
		if scanner.Err() != nil {
			return lang, errors.Wrap(scanner.Err(), "scanning file")
		}
		lang.TotalLines = uint64(linecount)

		lang.TotalBytes = uint64(file.Size())
	}
	lang.Name = matchedLang
	return lang, nil
}

// GetLanguageByFilename returns the guessed language for the named file (and safe == true if this
// is very likely to be correct).
func GetLanguageByFilename(name string) (language string, safe bool) {
	language, safe = enry.GetLanguageByExtension(name)
	if language == "GCC Machine Description" && filepath.Ext(name) == ".md" {
		language = "Markdown" // override detection for .md
	}
	return language, safe
}

func init() {
	// Treat .tsx and .jsx as TypeScript and JavaScript, respectively, instead of distinct languages
	// called "TSX" and "JSX". This is more consistent with user expectations.
	data.ExtensionsByLanguage["TypeScript"] = append(data.ExtensionsByLanguage["TypeScript"], ".tsx")
	data.LanguagesByExtension[".tsx"] = []string{"TypeScript"}
	data.ExtensionsByLanguage["JavaScript"] = append(data.ExtensionsByLanguage["JavaScript"], ".jsx")
	data.LanguagesByExtension[".jsx"] = []string{"JavaScript"}
}
