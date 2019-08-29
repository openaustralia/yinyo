package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	Describe("storagePath", func() {
		It("should convert app to a standard path", func() {
			Expect(storagePath("abc", "app", "tgz")).To(Equal("app/abc.tgz"))
		})

		It("should convert output to a standard path", func() {
			Expect(storagePath("def", "output", "")).To(Equal("output/def"))
		})
	})
})
