package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoragePath(t *testing.T) {
	assert.Equal(t, storagePath("abc", "app.tgz"), "abc/app.tgz")
	assert.Equal(t, storagePath("def", "output"), "def/output")
}
