package paelito_shared

import (
  "os"
  "path/filepath"
  "strings"
  "github.com/pkg/errors"
  "math/rand"
  "time"
)

const VersionFormat = "20060102T150405MST"
const Port = "45362"

type WordPosition struct {
  Word string
  ParagraphIndex int
  HtmlFilename string
}


func GetRootPath() (string, error) {
	hd, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "os error")
	}
	dd := os.Getenv("SNAP_USER_COMMON")
	if strings.HasPrefix(dd, filepath.Join(hd, "snap", "go")) || dd == "" {
		dd = filepath.Join(hd, "paelito_data")
    os.MkdirAll(dd, 0777)
	}

	return dd, nil
}


func UntestedRandomString(length int) string {
  var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
  const charset = "abcdefghijklmnopqrstuvwxyz1234567890"

  b := make([]byte, length)
  for i := range b {
    b[i] = charset[seededRand.Intn(len(charset))]
  }
  return string(b)
}


func DoesPathExists(p string) bool {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}
