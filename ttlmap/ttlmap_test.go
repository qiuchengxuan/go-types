package ttlmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTLMap(t *testing.T) {
	ttlMap := New[int, int](time.Second)
	v, ok := ttlMap.Get(0)
	assert.False(t, ok)
	assert.Equal(t, v, 0)

	ttlMap.Put(0, 10)
	v, ok = ttlMap.Get(0)
	assert.True(t, ok)
	assert.Equal(t, v, 10)
}

func TestExpire(t *testing.T) {
	ttlMap := New[int, int](time.Second)
	ttlMap.Put(0, 0)
	expireAt := ttlMap.head.expireAt

	_, ok := ttlMap.Get(0)
	assert.True(t, ok)
	assert.Greater(t, ttlMap.head.expireAt, expireAt)

	assert.NotZero(t, ttlMap.Len())
	_, ok = ttlMap.Get(0)
	assert.True(t, ok)

	ttlMap.head.expireAt -= 2 * time.Second
	_, ok = ttlMap.Get(0)
	assert.False(t, ok)
	assert.Equal(t, ttlMap.Len(), 0)
}

func TestExpireOnPut(t *testing.T) {
	ttlMap := New[int, int](time.Second)
	ttlMap.Put(0, 10)
	ttlMap.Put(1, 11)
	ttlMap.head.expireAt -= ttlMap.ttl * 2
	ttlMap.Put(2, 12)
	assert.Equal(t, ttlMap.Len(), 2)
	assert.True(t, ttlMap.Delete(1))
	assert.Equal(t, ttlMap.Len(), 1)
	assert.True(t, ttlMap.Delete(2))
	assert.Equal(t, ttlMap.Len(), 0)
}
