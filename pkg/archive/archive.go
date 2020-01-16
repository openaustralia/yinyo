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

// Validate checks whether the archive is correctly formatted. error is nil if it validates.
func Validate(content io.Reader) error {
	createDirectory := func(relativePath string, mode os.FileMode) error {
		return nil
	}

	createFile := func(relativePath string, mode os.FileMode, content io.Reader) error {
		return nil
	}

	createSymlink := func(relativeLinkPath string, relativePath string) error {
		return nil
	}

	return walk(content, createDirectory, createFile, createSymlink)
}

func walk(content io.Reader,
	directoryCallback func(relativePath string, mode os.FileMode) error,
	fileCallback func(relativePath string, mode os.FileMode, content io.Reader) error,
	symlinkCallback func(relativeLinkPath string, path string) error,
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
		switch file.Typeflag {
		case tar.TypeDir:
			err := directoryCallback(file.Name, 0755)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			mode := file.FileInfo().Mode()
			err := fileCallback(file.Name, mode, tarReader)
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			err = symlinkCallback(file.Linkname, file.Name)
			if err != nil {
				return err
			}
		default:
			return errors.New("unexpected type in tar")
		}
	}
	return nil
}

// ExtractToDirectory takes a tar, gzipped archive and extracts it to a directory on the filesystem
func ExtractToDirectory(content io.Reader, dir string) error {
	createDirectory := func(relativePath string, mode os.FileMode) error {
		// Only try to create the directory if this is a new one
		if filepath.Clean(relativePath) == "." {
			return nil
		}
		path := filepath.Join(dir, relativePath)
		return os.Mkdir(path, mode)
	}

	createFile := func(relativePath string, mode os.FileMode, content io.Reader) error {
		path := filepath.Join(dir, relativePath)
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

	createSymlink := func(relativeLinkPath string, relativePath string) error {
		path := filepath.Join(dir, relativePath)
		linkPath := filepath.Join(filepath.Dir(path), relativeLinkPath)
		return os.Symlink(linkPath, path)
	}

	return walk(content, createDirectory, createFile, createSymlink)
}

func node(path string, info os.FileInfo, dir string, tarWriter *tar.Writer) error {
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
	err = tarWriter.WriteHeader(header)
	if err != nil {
		return err
	}

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
}

// CreateFromDirectory creates an archive from a directory on the filesystem
// ignorePaths is a list of paths (relative to dir) that should be ignored and not archived
func CreateFromDirectory(dir string, ignorePaths []string) (io.Reader, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
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
		return node(path, info, dir, tarWriter)
	})
	if err != nil {
		return nil, err
	}
	return &buffer, nil
}
