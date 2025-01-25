package hxlru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLRU(t *testing.T) {
	cache := NewCacheOneLRU[
		paramsTestLRU,
		member,
	](
		&ParamsNewCacheLRU{
			TTL:      time.Second * 60,
			Capacity: 10,
		},
	)

	key1 := paramsTestLRU{
		ProjectID:      2,
		ProjectMembers: 13,
	}

	require.Error(t,
		cache.Delete(key1),
		"delete error",
	)

	cache.DeleteSilent(key1)

	value1 := member{
		Name: "John",
		Skills: []string{
			"manager",
			"audit",
		},
	}

	key2 := paramsTestLRU{
		ProjectID:      2,
		ProjectMembers: 16,
	}

	go cache.PutTTL(
		key2,
		value1,
	)

	cache.PutTTL(
		key1,
		value1,
	)

	results1, errGetValue := cache.Get(key1)
	require.NoError(t, errGetValue)
	require.NotEmpty(t, results1)
	require.Equal(t, *results1, value1)

	cache.DeleteSilent(key1)

	results2, errGetDeleted := cache.Get(key1)
	require.Error(t, errGetDeleted)
	require.Empty(t, results2)

	value2 := member{
		Name: "Mary",
		Skills: []string{
			"developer",
			"java",
		},
	}

	cache.PutTTL(
		key1,
		value2,
	)

	results3, errGetNewInsert := cache.Get(key1)
	require.NoError(t, errGetNewInsert)
	require.NotEmpty(t, results3)
	require.Equal(t, *results3, value2)
}

func TestEvictionLRU(t *testing.T) {
	cache := NewCacheOneLRU[int, string](
		&ParamsNewCacheLRU{
			TTL:      time.Second * 60,
			Capacity: 2,
		},
	)

	cache.PutTTL(1, "one")
	cache.PutTTL(2, "two")
	cache.PutTTL(3, "three") // Evicts 1

	value1, errGetValue1 := cache.Get(1)
	require.Error(t, errGetValue1)
	require.Nil(t, value1)

	value2, errGetValue2 := cache.Get(2)
	require.NoError(t, errGetValue2)
	require.Equal(t, "two", *value2)

	value3, errGetValue3 := cache.Get(3)
	require.NoError(t, errGetValue3)
	require.Equal(t, "three", *value3)
}

func TestOverwriteLRUPut(t *testing.T) {
	cache := NewCacheOneLRU[int, string](
		&ParamsNewCacheLRU{
			TTL:      time.Second * 60,
			Capacity: 1,
		},
	)
	require.NotNil(t, cache)

	cache.Put(1, "first value")

	value1, errGetValue1 := cache.Get(1)
	require.NoError(t, errGetValue1)
	require.Equal(t, "first value", *value1)

	cache.Put(1, "updated value")

	value2, errGetValue2 := cache.Get(1)
	require.NoError(t, errGetValue2)
	require.Equal(t, "updated value", *value2)
}

func TestOverwriteLRUPutTTL(t *testing.T) {
	cache := NewCacheOneLRU[int, string](
		&ParamsNewCacheLRU{
			TTL:      time.Second * 60,
			Capacity: 1,
		},
	)

	cache.PutTTL(1, "value1")

	value1, errGetValue1 := cache.Get(1)
	require.NoError(t, errGetValue1)
	require.Equal(t, "value1", *value1)

	cache.PutTTL(1, "value2")

	value2, errGetValue2 := cache.Get(1)
	require.NoError(t, errGetValue2)
	require.Equal(t, "value2", *value2)
}

func TestCacheOneLRU_String(t *testing.T) {
	cache := NewCacheOneLRU[int, string](
		&ParamsNewCacheLRU{
			Capacity: 2,
		},
	)

	expectedEmpty := "Capacity: 2\nCached:\n"
	require.Equal(t, expectedEmpty, cache.String())

	cache.Put(1, "one")
	expectedOne := "Capacity: 2\nCached:\nkey: 1, value: one\n"
	require.Equal(t, expectedOne, cache.String())

	cache.Put(2, "two")

	spool := cache.String()

	expectedTitle := "Capacity: 2\nCached:\n"
	expectedKey1 := "key: 1, value: one\n"
	expectedKey2 := "key: 2, value: two\n"

	require.Contains(t, spool, expectedTitle)
	require.Contains(t, spool, expectedKey1)
	require.Contains(t, spool, expectedKey2)
}
