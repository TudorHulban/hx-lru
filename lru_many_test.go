package hxlru

import (
	"testing"
	"time"

	goerrors "github.com/TudorHulban/go-errors"
	"github.com/stretchr/testify/require"
)

func TestCacheManyLRUPutRetrievesCorrectValues(t *testing.T) {
	cache := NewCacheManyLRU[paramsTestLRU, member](
		&ParamsNewCacheLRU{
			Capacity: 3,
		},
	)
	require.NotNil(t, cache)

	key1 := paramsTestLRU{ProjectID: 1, ProjectMembers: 10}
	key2 := paramsTestLRU{ProjectID: 2, ProjectMembers: 20}
	key3 := paramsTestLRU{ProjectID: 3, ProjectMembers: 30}

	value1 := []member{
		{Name: "Alice", Skills: []string{"Go"}},
		{Name: "Bob", Skills: []string{"Java"}},
	}
	value2 := []member{
		{Name: "Charlie", Skills: []string{"Python"}},
	}
	value3 := []member{
		{Name: "David", Skills: []string{"C++"}},
		{Name: "Eve", Skills: []string{"C#", "Go"}},
	}

	cache.Put(key1, value1)
	cache.Put(key2, value2)
	cache.Put(key3, value3)

	retrievedValue1, errGetValue1 := cache.Get(key1)
	require.NoError(t, errGetValue1)
	require.Equal(t, value1, retrievedValue1)

	retrievedValue2, errGetValue2 := cache.Get(key2)
	require.NoError(t, errGetValue2)
	require.Equal(t, value2, retrievedValue2)

	retrievedValue3, errGetValue3 := cache.Get(key3)
	require.NoError(t, errGetValue3)
	require.Equal(t, value3, retrievedValue3)

	// Eviction test
	key4 := paramsTestLRU{ProjectID: 4, ProjectMembers: 40}
	value4 := []member{{Name: "Frank", Skills: []string{"Rust"}}}
	cache.Put(key4, value4)

	evicted, errGetEvicted := cache.Get(key1)
	require.Error(t, errGetEvicted)
	require.Zero(t, evicted)

	cache.DeleteSilent(key4)

	deleted, errGetValue4 := cache.Get(key4)
	require.Error(t, errGetValue4)
	require.Zero(t, deleted)
}

func TestCacheManyLRUPutTTLExpiresEntriesCorrectly(t *testing.T) {
	cache := NewCacheManyLRU[paramsTestLRU, member](
		&ParamsNewCacheLRU{
			Capacity: 2,
			TTL:      100 * time.Millisecond,
		},
	)
	require.NotNil(t, cache)

	key1 := paramsTestLRU{ProjectID: 1}
	key2 := paramsTestLRU{ProjectID: 2}

	value1 := []member{
		{Name: "Alice", Skills: []string{"Go"}},
	}
	value2 := []member{
		{Name: "Bob", Skills: []string{"Java"}},
	}

	cache.PutTTL(key1, value1)
	cache.PutTTL(key2, value2)

	retrievedValue1, errGetValue1 := cache.Get(key1)
	require.NoError(t, errGetValue1)
	require.Equal(t, value1, retrievedValue1)

	time.Sleep(150 * time.Millisecond)

	retrievedValue2, errGetValue1AfterSleep := cache.Get(key1)
	require.Error(t, errGetValue1AfterSleep)
	require.ErrorIs(t, errGetValue1AfterSleep, goerrors.ErrEntryNotFound{Key: key1})
	require.Zero(t, retrievedValue2)

	retrievedValue3, errGetValue2AfterSleep := cache.Get(key2)
	require.Error(t, errGetValue2AfterSleep)
	require.ErrorIs(t, errGetValue2AfterSleep, goerrors.ErrEntryNotFound{Key: key2})
	require.Zero(t, retrievedValue3)
}

func TestCacheManyLRUPutOverwritesExistingEntry(t *testing.T) {
	cache := NewCacheManyLRU[paramsTestLRU, member](
		&ParamsNewCacheLRU{
			Capacity: 1,
		},
	)
	require.NotNil(t, cache)

	key := paramsTestLRU{ProjectID: 1}

	initialValue := []member{{Name: "first value", Skills: []string{}}}
	updatedValue := []member{{Name: "updated value", Skills: []string{}}}

	cache.Put(key, initialValue)

	retrievedValue1, errGet1 := cache.Get(key)
	require.NoError(t, errGet1)
	require.Equal(t, initialValue, retrievedValue1)

	cache.Put(key, updatedValue)

	retrievedValue2, errGet2 := cache.Get(key)
	require.NoError(t, errGet2)
	require.Equal(t, updatedValue, retrievedValue2)
}

func TestCacheManyLRUPutTTLOverwritesExistingEntry(t *testing.T) {
	cache := NewCacheManyLRU[paramsTestLRU, member](
		&ParamsNewCacheLRU{
			TTL:      time.Second * 60,
			Capacity: 1,
		},
	)
	require.NotNil(t, cache)

	key := paramsTestLRU{ProjectID: 1}

	initialValue := []member{{Name: "first value", Skills: []string{}}}
	updatedValue := []member{{Name: "updated value", Skills: []string{}}}

	cache.PutTTL(key, initialValue)

	retrievedValue1, errGet1 := cache.Get(key)
	require.NoError(t, errGet1)
	require.Equal(t, initialValue, retrievedValue1)

	cache.PutTTL(key, updatedValue)

	retrievedValue2, errGet2 := cache.Get(key)
	require.NoError(t, errGet2)
	require.Equal(t, updatedValue, retrievedValue2)
}

func TestCacheManyLRU_String(t *testing.T) {
	cache := NewCacheManyLRU[int, string](
		&ParamsNewCacheLRU{
			Capacity: 2,
		},
	)

	expectedEmpty := "Capacity: 2\nCached:\n"
	require.Equal(t, expectedEmpty, cache.String())

	cache.Put(1, []string{"one", "two"})
	expectedOne := "Capacity: 2\nCached:\nkey: 1, values: [one two]\n"
	require.Equal(t, expectedOne, cache.String())

	cache.Put(2, []string{"three"})

	spool := cache.String()

	expectedTitle := "Capacity: 2\nCached:\n"
	expectedKey1 := "key: 1, values: [one two]\n"
	expectedKey2 := "key: 2, values: [three]\n"

	require.Contains(t, spool, expectedTitle)
	require.Contains(t, spool, expectedKey1)
	require.Contains(t, spool, expectedKey2)
}
