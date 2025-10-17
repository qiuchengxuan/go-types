package ttlmap

import (
	"iter"
	"maps"
	"time"
)

var boot = time.Now()

type entry[K comparable, T any] struct {
	expireAt   time.Duration
	key        K
	value      T
	prev, next *entry[K, T]
}

type TTLMap[K comparable, V any] struct {
	ttl     time.Duration
	entries map[K]*entry[K, V]
	head    *entry[K, V]
}

func (m *TTLMap[K, V]) detach(entry *entry[K, V]) {
	if entry == entry.next {
		entry.prev, entry.next = nil, nil
		m.head = nil
		return
	}
	entry.prev.next, entry.next.prev = entry.next, entry.prev
}

func (m *TTLMap[K, V]) expireOne() bool {
	if entry := m.head; entry != nil && time.Since(boot) > entry.expireAt {
		delete(m.entries, entry.key)
		m.detach(entry)
		return true
	}
	return false
}

func (m *TTLMap[K, V]) tail() *entry[K, V] { return m.head.prev }

func (m *TTLMap[K, V]) attach(key K, value V) {
	entry := &entry[K, V]{time.Since(boot) + m.ttl, key, value, nil, m.head}
	m.entries[key] = entry
	if m.head == nil {
		entry.prev, entry.next = entry, entry
		m.head = entry
	} else {
		tail := m.tail()
		m.head.prev = entry
		entry.prev, tail.next = tail, entry
	}
}

func (m *TTLMap[K, V]) Put(key K, value V) {
	if m.expireOne() {
		m.expireOne() // extra call to ensure convergence
	}
	if existing, ok := m.entries[key]; ok {
		if existing == m.tail() {
			existing.expireAt, existing.value = time.Since(boot)+m.ttl, value
			return
		} else {
			m.detach(existing)
		}
	}
	m.attach(key, value)
}

func (m *TTLMap[K, V]) Len() int { return len(m.entries) }

func (m *TTLMap[K, V]) Get(key K) (V, bool) {
	entry, ok := m.entries[key]
	if !ok {
		var value V
		return value, false
	}
	if time.Since(boot) > entry.expireAt {
		delete(m.entries, key)
		m.detach(entry)
		return entry.value, false
	}
	if entry == m.tail() {
		entry.expireAt = time.Since(boot) + m.ttl
		return entry.value, true
	}
	m.detach(entry)
	m.attach(key, entry.value)
	return entry.value, true
}

func (m *TTLMap[K, V]) Delete(key K) bool {
	entry, ok := m.entries[key]
	if !ok {
		return false
	}
	delete(m.entries, key)
	m.detach(entry)
	return true
}

func (m *TTLMap[K, V]) Keys() iter.Seq[K] {
	return maps.Keys(m.entries)
}

// ttl should not be too long, otherwise may cause too much memory consumption
func New[K comparable, V any](ttl time.Duration) TTLMap[K, V] {
	return TTLMap[K, V]{ttl, make(map[K]*entry[K, V]), nil}
}
