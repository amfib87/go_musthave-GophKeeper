package service

import (
	"fmt"
	"path"
	"strings"
)

// Извлекаем ID из пути URL
func GetIDfromPath(pathURL string) (string, error) {
	segments := strings.Split(path.Clean(pathURL), "/")

	var id string
	if len(segments) >= 2 {
		id = segments[len(segments)-1] // последний сегмент — ID
	} else {
		return "", fmt.Errorf("ID not found in path")
	}

	return id, nil
}
