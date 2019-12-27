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

// ExtractToDirectory takes a tar, gzipped archive and extracts it to a directory on the filesystem
func ExtractToDirectory(content io.Reader, dir string) error {
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

		nameAbsolute := filepath.Join(dir, file.Name)
		linkNameAbsolute := filepath.Join(filepath.Dir(nameAbsolute), file.Linkname)

		switch file.Typeflag {
		case tar.TypeDir:
			// Only try to create the directory if this is a new one
			if nameAbsolute != dir {
				err := os.Mkdir(nameAbsolute, 0755)
				if err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(
				nameAbsolute,
				os.O_RDWR|os.O_CREATE|os.O_TRUNC,
				file.FileInfo().Mode(),
			)
			if err != nil {
				return err
			}
			_, err = io.Copy(f, tarReader)
			if err != nil {
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			err = os.Symlink(linkNameAbsolute, nameAbsolute)
			if err != nil {
				return err
			}
		default:
			return errors.New("Unexpected type in tar")
		}
	}
	return nil
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
