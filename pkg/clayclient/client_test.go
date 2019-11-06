package clayclient

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArchive(t *testing.T) {
	// First create a test directory which we want to archive
	err := os.Mkdir("test", 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir("test/wibble", 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile("test/foo.txt", []byte("foobar"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile("test/wibble/bar.txt", []byte("bar"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	// Use an absolute file path so we can test that this gets converted to a relative path
	// by the tarring and untarring
	abs, err := filepath.Abs("test/foo.txt")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Symlink(abs, "test/foo2.txt")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Symlink("foo.txt", "test/foo3.txt")
	if err != nil {
		t.Fatal(err)
	}

	// Create an archive
	reader, err := CreateArchiveFromDirectory("test")
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
	err = ExtractArchiveToDirectory(file, "test")
	if err != nil {
		t.Fatal(err)
	}

	// Now check the result
	c, err := ioutil.ReadFile("test/foo.txt")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "foobar", string(c))
	c, err = ioutil.ReadFile("test/wibble/bar.txt")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "bar", string(c))
	n, err := os.Readlink("test/foo2.txt")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test/foo.txt", n)
	n, err = os.Readlink("test/foo3.txt")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test/foo.txt", n)

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

func TestMarshalStartEvent(t *testing.T) {
	b, err := json.Marshal(StartEvent{Stage: "build"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `{"stage":"build","type":"start"}`, string(b))
}

func TestMarshalFinishEvent(t *testing.T) {
	b, err := json.Marshal(FinishEvent{Stage: "build"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `{"stage":"build","type":"finish"}`, string(b))
}

func TestMarshalLogEvent(t *testing.T) {
	b, err := json.Marshal(LogEvent{Stage: "build", Stream: "stdout", Text: "Hello"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `{"stage":"build","type":"log","stream":"stdout","text":"Hello"}`, string(b))
}
