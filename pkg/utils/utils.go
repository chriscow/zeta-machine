package utils

import "os"

// CreateFolder creates the entire path (wrapping os.MkdirAll) checking if it
// exists first. If the path exists it does not return an error.
func CreateFolder(path string) error {
	exists, err := PathExists(path)
	if err != nil {
		return err
	}

	if !exists {
		err := os.MkdirAll(path, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// PathExists returns whether the given file or directory exists or not
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
