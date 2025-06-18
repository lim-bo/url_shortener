package cache_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	cache "github.com/limbo/url_shortener/internal/url_cache_manager"
	"github.com/stretchr/testify/assert"
)

var (
	testCode = "abcd1234"
	testUrl  = "https://google.com"
)

func TestCachingURL(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	manager := cache.New(cache.RedisConfig{
		Address: mr.Addr(),
	})

	err = manager.CacheLink(testCode, testUrl)
	if err != nil {
		t.Error(err)
	}

	link, err := manager.GetLink(testCode)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, testUrl, link)
}
