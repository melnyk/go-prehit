package prehit

import (
	"reflect"
	"testing"
	"time"
)

type basicmetrics struct {
	count   int
	hits    int
	miss    int
	updates int
	evicted int
	errors  int
}

func (m *basicmetrics) Hit() {
	m.hits++
}
func (m *basicmetrics) Miss() {
	m.miss++
}
func (m *basicmetrics) Error() {
	m.errors++
}
func (m *basicmetrics) Add() {
	m.count++
}
func (m *basicmetrics) Update() {
	m.updates++
}
func (m *basicmetrics) Evict() {
	m.evicted++
}
func (m *basicmetrics) Delete() {
	m.count--
}

func TestNewCache(t *testing.T) {
	metrics := &basicmetrics{}
	c := NewCache[string, int](WithMaxSize(20), WithMetrics(metrics))
	if c == nil {
		t.Error("Cache is nil")
	}
}

func TestCacheBasic(t *testing.T) {
	c := NewCache[string, int](WithMaxSize(20))

	// Add case
	c.Set("test", 1, time.Second)
	if v, ok := c.Get("test"); !ok || v != 1 {
		t.Error("Cache set failed")
	}

	// Update case
	c.Set("test", 1, time.Second)

	// 2nd add case
	c.Set("test2", 2, time.Second)
	if v, ok := c.Get("test2"); !ok || v != 2 {
		t.Error("Cache set failed")
	}

	// Delete case
	c.Delete("test")
	if _, ok := c.Get("test"); ok {
		t.Error("Cache delete failed")
	}

	// Get case - test2
	if v, ok := c.Get("test2"); !ok || v != 2 {
		t.Error("Cache get failed")
	}

	// Corruption case
	c.index["test2"] = nil
	if _, ok := c.Get("test2"); ok {
		t.Error("Cache get failed")
	}

	c.Reset()
}

func TestCacheErrors(t *testing.T) {
	c := NewCache[string, int](WithMaxSize(1))

	// Add case
	c.Set("test", 1, time.Second)
	if v, ok := c.Get("test"); !ok || v != 1 {
		t.Error("Cache set failed")
	}

	// 2nd add case
	c.Set("test2", 2, time.Second)
	if v, ok := c.Get("test2"); !ok || v != 2 {
		t.Error("Cache set failed")
	}

	// Delete case
	c.Delete("test2")
	if _, ok := c.Get("test"); ok {
		t.Error("Cache delete failed")
	}

	// Corruption
	c.index["test"] = nil
	if _, ok := c.Get("test"); ok {
		t.Error("Cache delete failed")
	}

	// Corruption
	c.index["test"] = nil
	c.Delete("test")

	// Corruption
	c.Set("test", 2, 0*time.Second)
	c.size = 0
	if _, ok := c.Get("test"); ok {
		t.Error("Cache set failed")
	}

	// Corruption
	c.index["test"] = nil
	c.deleteexpired("test", time.Now())
}

func TestCacheSet(t *testing.T) {
	metrics := &basicmetrics{}
	c := NewCache[string, int](WithMaxSize(20), WithMetrics(metrics))

	// Add case
	c.Set("test", 1, time.Second)
	if v, ok := c.Get("test"); !ok || v != 1 {
		t.Error("Cache set failed")
	}

	if metrics.count != 1 {
		t.Error("Cache metrics count failed")
	}

	// Update case
	c.Set("test", 1, time.Second)

	if metrics.count != 1 {
		t.Error("Cache metrics count failed")
	}

	if metrics.updates != 1 {
		t.Error("Cache metrics updates failed")
	}

	// 2nd add case
	c.Set("test2", 2, time.Second)
	if v, ok := c.Get("test2"); !ok || v != 2 {
		t.Error("Cache set failed")
	}

	if metrics.count != 2 {
		t.Error("Cache metrics count failed")
	}

	if metrics.updates != 1 {
		t.Error("Cache metrics updates failed")
	}
}

func TestCacheSetAtMax(t *testing.T) {
	metrics := &basicmetrics{}
	c := NewCache[string, int](WithMaxSize(3), WithMetrics(metrics))

	// Add case
	c.Set("test", 1, time.Second)

	if metrics.count != 1 {
		t.Error("Cache metrics count failed")
	}

	// Update case
	c.Set("test", 1, time.Second)

	if metrics.count != 1 {
		t.Error("Cache metrics count failed")
	}

	if metrics.updates != 1 {
		t.Error("Cache metrics updates failed")
	}

	// 2nd add case
	c.Set("test2", 2, time.Second)
	if v, ok := c.Get("test2"); !ok || v != 2 {
		t.Error("Cache set failed")
	}

	if metrics.count != 2 {
		t.Error("Cache metrics count failed")
	}

	if metrics.updates != 1 {
		t.Error("Cache metrics updates failed")
	}

	if metrics.evicted != 0 {
		t.Error("Cache metrics evicted failed")
	}

	// 3nd add case
	c.Set("test3", 2, time.Second)
	if v, ok := c.Get("test3"); !ok || v != 2 {
		t.Error("Cache set failed")
	}

	if metrics.count != 3 {
		t.Error("Cache metrics count failed")
	}

	if metrics.evicted != 0 {
		t.Error("Cache metrics evicted failed")
	}

	// 4nd add case
	c.Set("test4", 2, time.Second)
	if v, ok := c.Get("test4"); !ok || v != 2 {
		t.Error("Cache set failed")
	}

	if metrics.count != 3 {
		t.Error("Cache metrics count failed")
	}

	if metrics.count != int(c.size) {
		t.Error("Cache metrics count failed")
	}

	if metrics.evicted != 1 {
		t.Error("Cache metrics evicted failed")
	}

	// 5nd add case
	c.Set("test5", 2, time.Second)
	if v, ok := c.Get("test5"); !ok || v != 2 {
		t.Error("Cache set failed")
	}

	if metrics.count != 3 {
		t.Error("Cache metrics count failed")
	}

	if metrics.miss != 0 {
		t.Error("Cache metrics miss failed")
	}

	if metrics.hits != 4 {
		t.Error("Cache metrics hits failed")
	}

	if metrics.evicted != 2 {
		t.Error("Cache metrics evicted failed")
	}

	if v, ok := c.Get("test2"); ok || v != 0 {
		t.Error("Cache set failed")
	}

	if metrics.count != 3 {
		t.Error("Cache metrics count failed")
	}

	if metrics.miss != 1 {
		t.Error("Cache metrics miss failed")
	}

	if metrics.hits != 4 {
		t.Error("Cache metrics hits failed")
	}

	if metrics.count != int(c.size) {
		t.Error("Cache metrics count failed")
	}

	if metrics.errors != 0 {
		t.Error("Cache metrics errors failed")
	}
}

func TestCacheGet(t *testing.T) {
	metrics := &basicmetrics{}
	c := NewCache[string, int](WithMaxSize(20), WithMetrics(metrics))

	// Miss case
	if v, ok := c.Get("test"); ok || v != 0 {
		t.Error("Cache get failed")
	}

	if metrics.count != 0 {
		t.Error("Cache metrics count failed")
	}

	if metrics.hits != 0 {
		t.Error("Cache metrics hits failed")
	}

	if metrics.miss != 1 {
		t.Error("Cache metrics miss failed")
	}

	// Add
	c.Set("test", 1, time.Second)

	// Hit case
	if v, ok := c.Get("test"); !ok || v != 1 {
		t.Error("Cache get failed")
	}

	if metrics.count != 1 {
		t.Error("Cache metrics count failed")
	}

	if metrics.updates != 0 {
		t.Error("Cache metrics updates failed")
	}

	if metrics.hits != 1 {
		t.Error("Cache metrics hits failed")
	}

	c.Set("test", 1, time.Second)

	if metrics.count != 1 {
		t.Error("Cache metrics count failed")
	}

	if metrics.updates != 1 {
		t.Error("Cache metrics updates failed")
	}

	if v, ok := c.Get("test"); !ok || v != 1 {
		t.Error("Cache get failed")
	}

	if metrics.count != 1 {
		t.Error("Cache metrics count failed")
	}

	if metrics.hits != 2 {
		t.Error("Cache metrics hits failed")
	}

	if metrics.miss != 1 {
		t.Error("Cache metrics miss failed")
	}

	if metrics.errors != 0 {
		t.Error("Cache metrics errors failed")
	}
}

func TestCacheGetExpired(t *testing.T) {
	metrics := &basicmetrics{}
	c := NewCache[string, int](WithMaxSize(20), WithMetrics(metrics))
	c.Set("test", 1, time.Second)
	c.Set("test2", 2, time.Second)
	c.Set("test3", 2, time.Second)

	if metrics.count != 3 {
		t.Error("Cache metrics count failed")
	}

	if v, ok := c.Get("test"); !ok || v != 1 {
		t.Error("Cache get failed")
	}

	time.Sleep(2 * time.Second)

	if v, ok := c.Get("test"); ok || v != 0 {
		t.Error("Cache get failed")
	}

	if metrics.count != 2 {
		t.Error("Cache metrics count failed")
	}

	if metrics.evicted != 1 {
		t.Error("Cache metrics evicted failed")
	}

	if v, ok := c.Get("test2"); ok || v != 0 {
		t.Error("Cache get failed")
	}

	if metrics.count != 1 {
		t.Error("Cache metrics count failed")
	}

	if metrics.hits != 1 {
		t.Error("Cache metrics hits failed")
	}

	if metrics.miss != 2 {
		t.Error("Cache metrics miss failed")
	}

	if metrics.evicted != 2 {
		t.Error("Cache metrics evicted failed")
	}

	if metrics.errors != 0 {
		t.Error("Cache metrics errors failed")
	}
}

func TestCacheDelete(t *testing.T) {
	metrics := &basicmetrics{}
	c := NewCache[string, int](WithMaxSize(20), WithMetrics(metrics))
	c.Set("test", 1, time.Second)
	c.Delete("test")
	if v, ok := c.Get("test"); ok || v != 0 {
		t.Error("Cache delete failed")
	}

	if metrics.count != 0 {
		t.Error("Cache metrics count failed")
	}

	if metrics.hits != 0 {
		t.Error("Cache metrics hits failed")
	}

	if metrics.miss != 1 {
		t.Error("Cache metrics miss failed")
	}

	c.Set("test", 1, time.Second)
	c.Set("test2", 1, time.Second)
	c.Set("test3", 1, time.Second)
	c.Set("test4", 1, time.Second)
	if metrics.count != 4 {
		t.Error("Cache metrics count failed")
	}

	c.Delete("test1")
	if metrics.count != 4 {
		t.Error("Cache metrics count failed")
	}

	c.Delete("test")
	if metrics.count != 3 {
		t.Error("Cache metrics count failed")
	}

	c.Delete("test3")
	if metrics.count != 2 {
		t.Error("Cache metrics count failed")
	}

	c.Delete("test2")
	if metrics.count != 1 {
		t.Error("Cache metrics count failed")
	}

	c.Delete("test4")
	if metrics.count != 0 {
		t.Error("Cache metrics count failed")
	}
}

func TestCacheReset(t *testing.T) {
	metrics := &basicmetrics{}
	c := NewCache[string, int](WithMaxSize(20), WithMetrics(metrics))
	c.Set("test", 1, time.Second)

	c.Reset()

	if c.head != nil || c.tail != nil {
		t.Error("Cache reset failed")
	}

	if metrics.count != 0 {
		t.Error("Cache metrics count failed")
	}

	if v, ok := c.Get("test"); ok || v != 0 {
		t.Error("Cache clear failed")
	}

	c.Set("test1", 1, time.Second)
	c.Set("test2", 2, time.Second)
	c.Set("test3", 3, time.Second)
	c.Set("test4", 4, time.Second)

	c.Reset()

	if c.head != nil || c.tail != nil {
		t.Error("Cache reset failed")
	}

	if metrics.count != 0 {
		t.Error("Cache metrics count failed")
	}
}

func TestCacheInternal(t *testing.T) {
	c := NewCache[string, int](WithMaxSize(10))
	c.Set("test1", 1, time.Second)
	c.Set("test2", 2, time.Second)
	c.Set("test3", 3, time.Second)
	c.Set("test4", 4, time.Second)

	expected := []string{"test4", "test3", "test2", "test1"}
	list := make([]string, 0)
	next := c.head
	for next != nil {
		list = append(list, next.key)
		next = next.next
	}

	if !reflect.DeepEqual(expected, list) {
		t.Error("Cache internal failed")
	}

	c.Set("test1", 1, time.Second)
	expected = []string{"test1", "test4", "test3", "test2"}
	list = make([]string, 0)
	next = c.head
	for next != nil {
		list = append(list, next.key)
		next = next.next
	}

	if !reflect.DeepEqual(expected, list) {
		t.Error("Cache internal failed")
	}

	c.Set("test3", 2, time.Second)
	expected = []string{"test1", "test4", "test3", "test2"}
	list = make([]string, 0)
	next = c.head
	for next != nil {
		list = append(list, next.key)
		next = next.next
	}

	if !reflect.DeepEqual(expected, list) {
		t.Error("Cache internal failed")
	}

	c.Set("test2", 3, time.Second)
	expected = []string{"test2", "test1", "test4", "test3"}
	list = make([]string, 0)
	next = c.head
	for next != nil {
		list = append(list, next.key)
		next = next.next
	}

	if !reflect.DeepEqual(expected, list) {
		t.Error("Cache internal failed")
	}

	c.Set("test4", 4, time.Second)
	expected = []string{"test2", "test1", "test4", "test3"}
	list = make([]string, 0)
	next = c.head
	for next != nil {
		list = append(list, next.key)
		next = next.next
	}

	if !reflect.DeepEqual(expected, list) {
		t.Error("Cache internal failed")
	}

	c.Delete("test4")
	expected = []string{"test2", "test1", "test3"}
	list = make([]string, 0)
	next = c.head
	for next != nil {
		list = append(list, next.key)
		next = next.next
	}

	if !reflect.DeepEqual(expected, list) {
		t.Error("Cache internal failed")
	}

	c.Delete("test4", "test2", "test3")
	expected = []string{"test1"}
	list = make([]string, 0)
	next = c.head
	for next != nil {
		list = append(list, next.key)
		next = next.next
	}

	if !reflect.DeepEqual(expected, list) {
		t.Error("Cache internal failed")
	}
}

// cache benchmark test
func BenchmarkCacheSetOnly(b *testing.B) {
	c := NewCache[int, int](WithMaxSize((uint)(b.N)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i, time.Second)
	}
}

func BenchmarkCacheSetOnLimit(b *testing.B) {
	c := NewCache[int, int](WithMaxSize(3))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i, time.Second)
	}
}

func BenchmarkCacheUpdate(b *testing.B) {
	c := NewCache[int, int](WithMaxSize((uint)(b.N)))
	for i := 0; i < b.N; i++ {
		c.Set(i, i, 100*time.Second)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i, time.Second)
	}
}

func BenchmarkCacheReset(b *testing.B) {
	c := NewCache[int, int](WithMaxSize((uint)(b.N)))
	for i := 0; i < b.N; i++ {
		c.Set(i, i, time.Second)
	}
	b.ResetTimer()
	c.Reset()
}

func BenchmarkCacheSetPrealloc(b *testing.B) {
	c := NewCache[int, int](WithMaxSize((uint)(b.N / 10000)))
	for i := 0; i < b.N; i++ {
		c.Set(i, i, time.Second)
	}
	c.Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i, time.Second)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	c := NewCache[int, int](WithMaxSize((uint)(b.N)))
	for i := 0; i < b.N; i++ {
		c.Set(i, i, 100*time.Second)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(i)
	}
}

func BenchmarkCacheGetExpired(b *testing.B) {
	c := NewCache[int, int](WithMaxSize((uint)(b.N)))
	for i := 0; i < b.N; i++ {
		c.Set(i, i, time.Microsecond)
	}
	time.Sleep(10 * time.Millisecond)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(i)
	}
}

func BenchmarkTimeNow(b *testing.B) {
	var res time.Time
	for i := 0; i < b.N; i++ {
		res = time.Now()
	}
	_ = res
}
