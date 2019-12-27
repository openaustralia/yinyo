package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func walk(content io.Reader, dir string,
	directoryCallback func(path string, mode os.FileMode) error,
	fileCallback func(path string, mode os.FileMode, content io.Reader) error,
	symlinkCallback func(linkPath string, path string) error,
) error {
	gzipReader, err := gzip.NewReader(content)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzipReader)
	for {
		file, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		if filepath.IsAbs(file.Name) {
			return errors.New("file paths should all be relative")
		}
		if filepath.IsAbs(file.Linkname) {
			return errors.New("links should all be relative")

		}
		nameAbsolute := filepath.Join(dir, file.Name)
		linkNameAbsolute := filepath.Join(filepath.Dir(nameAbsolute), file.Linkname)

		switch file.Typeflag {
		case tar.TypeDir:
			// Only try to create the directory if this is a new one
			if nameAbsolute != dir {
				err := directoryCallback(nameAbsolute, 0755)
				if err != nil {
					return err
				}
			}
		case tar.TypeReg:
			mode := file.FileInfo().Mode()
			err := fileCallback(nameAbsolute, mode, tarReader)
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			err = symlinkCallback(linkNameAbsolute, nameAbsolute)
			if err != nil {
				return err
			}
		default:
			return errors.New("Unexpected type in tar")
		}
	}
	return nil
}

// ExtractToDirectory takes a tar, gzipped archive and extracts it to a directory on the filesystem
func ExtractToDirectory(content io.Reader, dir string) error {
	createDirectory := func(path string, mode os.FileMode) error {
		return os.Mkdir(path, mode)
	}

	createFile := func(path string, mode os.FileMode, content io.Reader) error {
		f, err := os.OpenFile(
			path,
			os.O_RDWR|os.O_CREATE|os.O_TRUNC,
			mode,
		)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, content)
		return err
	}

	createSymlink := func(linkPath string, path string) error {
		return os.Symlink(linkPath, path)
	}

	return walk(content, dir, createDirectory, createFile, createSymlink)
}

// CreateFromDirectory creates an archive from a directory on the filesystem
// ignorePaths is a list of paths (relative to dir) that should be ignored and not archived
func CreateFromDirectory(dir string, ignorePaths []string) (io.Reader, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		for _, ignorePath := range ignorePaths {
			if path == filepath.Join(dir, ignorePath) {
				return nil
			}
		}
		relativePath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		var link string
		if info.Mode()&os.ModeSymlink != 0 {
			link, err = os.Readlink(path)
			if err != nil {
				return err
			}
			if filepath.IsAbs(link) {
				// Convert the absolute link to a relative link
				absPath, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				d := filepath.Dir(absPath)
				link, err = filepath.Rel(d, link)
				if err != nil {
					return err
				}
			}
		}
		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		header.Name = relativePath
		tarWriter.WriteHeader(header)

		// If it's a regular file then write the contents
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(tarWriter, f)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	// TODO: This should always get called
	tarWriter.Close()
	gzipWriter.Close()
	return &buffer, nil
}
