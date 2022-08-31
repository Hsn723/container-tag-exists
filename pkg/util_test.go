package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertExpectedErr(t *testing.T, err error, isErr bool) {
	t.Helper()
	if isErr {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}

func TestExtractRegistryURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title  string
		image  string
		expect string
		isErr  bool
	}{
		{
			title:  "GHCR",
			image:  "ghcr.io/hsn723/hoge",
			expect: "ghcr.io",
		},
		{
			title:  "HogeRegistry",
			image:  "my-hoge.registry.dev/hsn723/hoge",
			expect: "my-hoge.registry.dev",
		},
		{
			title:  "RegistryWithPort",
			image:  "registry.dev:3000/hsn723/hoge",
			expect: "registry.dev:3000",
		},
		{
			title: "EmptyString",
			image: "",
			isErr: true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.title, func(t *testing.T) {
			t.Parallel()
			actual, err := ExtractRegistryURL(c.image)
			assertExpectedErr(t, err, c.isErr)
			assert.Equal(t, actual, c.expect)
		})
	}
}

func TestExtractImagePath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title  string
		image  string
		expect string
		isErr  bool
	}{
		{
			title:  "SingleNamespace",
			image:  "ghcr.io/hsn723/hoge",
			expect: "hsn723/hoge",
		},
		{
			title:  "SubNamespaces",
			image:  "my-hoge.registry.dev/hsn723/hoge/hige",
			expect: "hsn723/hoge/hige",
		},
		{
			title:  "RegistryWithPort",
			image:  "registry.dev:3000/hsn723/hoge",
			expect: "hsn723/hoge",
		},
		{
			title: "EmptyString",
			image: "",
			isErr: true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.title, func(t *testing.T) {
			t.Parallel()
			actual, err := ExtractImagePath(c.image)
			assertExpectedErr(t, err, c.isErr)
			assert.Equal(t, actual, c.expect)
		})
	}
}

func TestNormalizeRegistryName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title  string
		url    string
		expect string
	}{
		{
			title:  "GHCR",
			url:    "ghcr.io",
			expect: "GHCR_IO",
		},
		{
			title:  "GHCR",
			url:    "ghcr.io",
			expect: "GHCR_IO",
		},
		{
			title:  "RegistryWithPort",
			url:    "registry.dev:3000",
			expect: "REGISTRY_DEV_3000",
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.title, func(t *testing.T) {
			t.Parallel()
			actual := NormalizeRegistryName(c.url)
			assert.Equal(t, actual, c.expect)
		})
	}
}
