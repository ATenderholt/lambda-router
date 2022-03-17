package zip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func createDirs(dirPath string) error {
	logger.Debugf("Creating directory if necessary %s ...", dirPath)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		e := fmt.Errorf("unable to create directory %s: %v", dirPath, err)
		logger.Error(e)
		return e
	}

	return nil
}

func saveFile(filePath string, file zip.File) error {
	destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		msg := fmt.Sprintf("unable to open file %s: %v", filePath, err)
		logger.Error(msg)
		return errors.New(msg)
	}
	defer destFile.Close()

	fileInArchive, err := file.Open()
	if err != nil {
		msg := fmt.Sprintf("unable to decompress file %s: %v", file.Name, err)
		logger.Error(msg)
		return errors.New(msg)
	}
	defer fileInArchive.Close()

	_, err = io.Copy(destFile, fileInArchive)
	if err != nil {
		msg := fmt.Sprintf("problem decompressing file %s: %v", file.Name, err)
		logger.Error(msg)
		return errors.New(msg)
	}

	return nil
}

func UncompressZipFile(file string, destPath string) error {
	reader, err := zip.OpenReader(file)
	if err != nil {
		msg := fmt.Sprintf("unable to uncompress zip from file %s: %v", file, err)
		logger.Error(msg)
		return errors.New(msg)
	}
	defer reader.Close()

	return decompressZipFile(&reader.Reader, destPath)
}

func UncompressZipFileBytes(bytes []byte, destPath string) error {
	content := ZipContent{Content: bytes, Length: int64(len(bytes))}
	reader, err := zip.NewReader(content, content.Length)
	if err != nil {
		msg := fmt.Sprintf("Unable to uncompress zip from bytes: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	return decompressZipFile(reader, destPath)
}

func decompressZipFile(reader *zip.Reader, destPath string) error {
	for _, f := range reader.File {
		filePath := filepath.Join(destPath, f.Name)

		var err error
		if f.FileInfo().IsDir() {
			err = createDirs(filePath)
			continue
		} else {
			err = createDirs(filepath.Dir(filePath))
		}

		if err != nil {
			msg := fmt.Sprintf("unable to create directory for file %s in %s: %v", f.Name, destPath, err)
			logger.Error(msg)
			return errors.New(msg)
		}

		logger.Info("Saving %s ...", filePath)
		err = saveFile(filePath, *f)
		if err != nil {
			msg := fmt.Sprintf("unable to save file %s in %s: %v", f.Name, destPath, err)
			logger.Error(msg)
			return errors.New(msg)
		}

	}

	return nil
}
