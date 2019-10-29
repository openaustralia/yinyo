package clayclient

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArchive(t *testing.T) {
	// First create a test directory which we want to archive
	err := os.Mkdir("test", 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile("test/foo.txt", []byte("foobar"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create an archive
	reader, err := createArchiveFromDirectory("test")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create("test.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(file, reader)
	file.Close()
	err = os.RemoveAll("test")
	if err != nil {
		t.Fatal(err)
	}

	// Extract the archive
	file, err = os.Open("test.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir("test", 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = extractArchiveToDirectory(file, "test")
	if err != nil {
		t.Fatal(err)
	}

	// Now check the result
	c, err := ioutil.ReadFile("test/foo.txt")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "foobar", string(c))

	// And finally tidy up
	err = os.RemoveAll("test")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove("test.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
}
