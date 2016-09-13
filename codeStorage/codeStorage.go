package codeStorage

import (
	"io/ioutil"
	"os"

	"archive/zip"

	"path/filepath"

	"io"

	"github.com/satori/go.uuid"
)

const (
	StoragePath = "/tmp/executor"
	TempPath    = "/tmp"
)

func baseDirForKey(key string) string {
	return filepath.Join(StoragePath, key)
}

func unzipFile(f *zip.File, key string) error {
	dir := filepath.Join(baseDirForKey(key), f.Name)
	createDirErr := os.MkdirAll(filepath.Dir(dir), os.ModeDir|os.ModePerm)
	if createDirErr != nil {
		return createDirErr
	}

	rc, fErr := f.Open()
	defer rc.Close()
	if fErr != nil {
		return fErr
	}

	file, createFileErr := os.Create(dir)
	defer file.Close()
	if createFileErr != nil {
		return createFileErr
	}

	_, copyErr := io.Copy(file, rc)
	if copyErr != nil {
		return copyErr
	}

	return nil
}

func StoreCode(key string, codeZipped []byte) error {
	tmpFilename := filepath.Join(TempPath, uuid.NewV4().String())
	writeFileError := ioutil.WriteFile(tmpFilename, codeZipped, os.ModePerm)
	if writeFileError != nil {
		return writeFileError
	}

	reader, zipReaderError := zip.OpenReader(tmpFilename)
	if zipReaderError != nil {
		return zipReaderError
	}

	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		unzipFileErr := unzipFile(f, key)
		if unzipFileErr != nil {
			return unzipFileErr
		}
	}

	return nil
}

func GetStoredFiles(key string) ([]os.FileInfo, error) {
	files := make([]os.FileInfo, 0)

	walkErr := filepath.Walk(baseDirForKey(key), func(_ string, info os.FileInfo, walkError error) error {
		if walkError != nil {
			return walkError
		}

		files = append(files, info)
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return files, nil
}
