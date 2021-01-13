package utils

import "os"

// CreateFolder creates the entire path (wrapping os.MkdirAll) checking if it
// exists first. If the path exists it does not return an error.
func CreateFolder(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info == nil {
		err := os.MkdirAll(path, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}
