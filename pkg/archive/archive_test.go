package archive

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArchive(t *testing.T) {
	// First create a test directory which we want to archive
	os.Mkdir("test", 0755)
	os.Mkdir("test/wibble", 0755)
	ioutil.WriteFile("test/foo.txt", []byte("foobar"), 0644)
	ioutil.WriteFile("test/wibble/bar.txt", []byte("bar"), 0644)
	// Use an absolute file path so we can test that this gets converted to a relative path
	// by the tarring and untarring
	abs, _ := filepath.Abs("test/foo.txt")
	os.Symlink(abs, "test/foo2.txt")
	os.Symlink("foo.txt", "test/foo3.txt")

	// Create an archive
	reader, err := CreateFromDirectory("test", []string{})
	if err != nil {
		t.Fatal(err)
	}
	file, _ := os.Create("test.tar.gz")
	io.Copy(file, reader)
	file.Close()
	os.RemoveAll("test")

	// Extract the archive
	file, _ = os.Open("test.tar.gz")
	os.Mkdir("test", 0755)
	err = ExtractToDirectory(file, "test")
	if err != nil {
		t.Fatal(err)
	}

	// Now check the result
	c, _ := ioutil.ReadFile("test/foo.txt")
	assert.Equal(t, "foobar", string(c))
	c, _ = ioutil.ReadFile("test/wibble/bar.txt")
	assert.Equal(t, "bar", string(c))
	n, _ := os.Readlink("test/foo2.txt")
	assert.Equal(t, "test/foo.txt", n)
	n, _ = os.Readlink("test/foo3.txt")
	assert.Equal(t, "test/foo.txt", n)

	// And finally tidy up
	os.RemoveAll("test")
	os.Remove("test.tar.gz")
}

func TestValidArchive(t *testing.T) {
	// Check that a zero-size file doesn't validate
	f, _ := os.Open(filepath.Join("testdata", "zero.tgz"))
	err := Validate(f)
	assert.Equal(t, err, io.EOF)
}

func TestValidArchiveEmpty(t *testing.T) {
	f, _ := os.Open(filepath.Join("testdata", "empty.tgz"))
	err := Validate(f)
	assert.Nil(t, err)
}
