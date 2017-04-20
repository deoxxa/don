package bcache

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

type Option func(c *Cache)
type Worker func(key string, userdata interface{}) ([]byte, error)
type Strategy func(t time.Time, a, b Meta) bool
type EvictionFunc func(m Meta, t time.Time, d []byte, err []byte)

type Cache struct {
	name       string
	worker     Worker
	path       string
	keepErrors bool
	lowMark    int
	highMark   int
	maxAge     time.Duration
	strategy   Strategy

	dbLock     sync.RWMutex
	db         *bolt.DB
	dbError    error
	dbReady    bool
	dbCanClose bool

	runningLock sync.RWMutex
	running     map[string]*sync.RWMutex
	runningRefs map[*sync.RWMutex]int

	onEviction EvictionFunc
}

func New(name string, options ...Option) *Cache {
	c := Cache{
		name:        name,
		running:     make(map[string]*sync.RWMutex),
		runningRefs: make(map[*sync.RWMutex]int),
	}

	for _, fn := range options {
		fn(&c)
	}

	return &c
}

func (c *Cache) ForceInit() error {
	return c.ensureOpen()
}

func (c *Cache) Close() error {
	if c.db == nil || !c.dbCanClose {
		return nil
	}

	return c.db.Close()
}

func SetName(name string) Option {
	return func(c *Cache) {
		c.name = name
	}
}

func SetWorker(worker Worker) Option {
	return func(c *Cache) {
		c.worker = worker
	}
}

func SetPath(path string) Option {
	return func(c *Cache) {
		c.path = path
	}
}

func SetDB(db *bolt.DB) Option {
	return func(c *Cache) {
		c.db = db
		c.dbCanClose = false
	}
}

func SetKeepErrors(keepErrors bool) Option {
	return func(c *Cache) {
		c.keepErrors = keepErrors
	}
}

func SetHighMark(highMark int) Option {
	return func(c *Cache) {
		c.highMark = highMark
	}
}

func SetLowMark(lowMark int) Option {
	return func(c *Cache) {
		c.lowMark = lowMark
	}
}

func SetLimit(maxEntries int, overCorrect float64) Option {
	return func(c *Cache) {
		c.highMark = maxEntries
		c.lowMark = maxEntries - int(float64(maxEntries)*overCorrect)
	}
}

func SetMaxAge(age time.Duration) Option {
	return func(c *Cache) {
		c.maxAge = age
	}
}

func SetStrategy(strategy Strategy) Option {
	return func(c *Cache) {
		c.strategy = strategy
	}
}

func SetOnEviction(evictionFunc EvictionFunc) Option {
	return func(c *Cache) {
		c.onEviction = evictionFunc
	}
}

func (c *Cache) ensureOpen() error {
	c.dbLock.RLock()
	if c.dbReady || c.dbError != nil {
		c.dbLock.RUnlock()
		return c.dbError
	}
	c.dbLock.RUnlock()

	c.dbLock.Lock()
	defer c.dbLock.Unlock()

	if c.dbReady || c.dbError != nil {
		return c.dbError
	}

	db := c.db
	dbCanClose := c.dbCanClose

	if db == nil {
		if c.path == "" {
			c.dbError = errors.New("Cache.ensureOpen: no db or path provided; can't continue")
			return c.dbError
		}

		newDB, err := bolt.Open(c.path, 0644, nil)
		if err != nil {
			c.dbError = err
			return err
		}

		db = newDB
		dbCanClose = true
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(c.name + "#meta")); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists([]byte(c.name + "#data")); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists([]byte(c.name + "#errs")); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if c.db == nil {
			// there was already an error we can't recover from, so this is just
			// housekeeping. this is why there's no error check on db.Close().
			_ = db.Close()
		}

		c.dbError = err
		return c.dbError
	}

	c.db = db
	c.dbReady = true
	c.dbCanClose = dbCanClose

	return nil
}

type Meta struct {
	Hash        [20]byte
	CreatedAt   uint64
	AccessedAt  uint64
	AccessCount uint64
}

func (m *Meta) decode(b []byte) error {
	if len(b) != 24 {
		return errors.New("invalid data length")
	}

	m.CreatedAt = binary.BigEndian.Uint64(b[0:8])
	m.AccessedAt = binary.BigEndian.Uint64(b[8:16])
	m.AccessCount = binary.BigEndian.Uint64(b[16:24])

	return nil
}

func (m *Meta) encode() []byte {
	var b [24]byte
	binary.BigEndian.PutUint64(b[0:8], m.CreatedAt)
	binary.BigEndian.PutUint64(b[8:16], m.AccessedAt)
	binary.BigEndian.PutUint64(b[16:24], m.AccessCount)
	return b[:]
}

func StrategyFIFO() Strategy {
	return func(_ time.Time, a, b Meta) bool {
		return a.CreatedAt < b.CreatedAt
	}
}

func StrategyLRU() Strategy {
	return func(_ time.Time, a, b Meta) bool {
		return a.AccessedAt < b.AccessedAt
	}
}

func StrategyLFU() Strategy {
	return func(_ time.Time, a, b Meta) bool {
		return a.AccessCount < b.AccessCount
	}
}

func (c *Cache) Get(key string, userdata interface{}) ([]byte, bool, error) {
	return c.GetAt(key, time.Now(), userdata)
}

func (c *Cache) readJobState(tx *bolt.Tx, k []byte, t time.Time, m *Meta, rres *[]byte, rerr *error, found *bool) error {
	mb := tx.Bucket([]byte(c.name + "#meta"))
	db := tx.Bucket([]byte(c.name + "#data"))
	eb := tx.Bucket([]byte(c.name + "#errs"))

	md := mb.Get(k)
	if len(md) == 0 {
		return nil
	}

	if err := m.decode(md); err != nil {
		return err
	}

	// don't return items that have expired
	if c.maxAge != 0 && time.Duration(t.UnixNano()-int64(m.CreatedAt)) > c.maxAge {
		return nil
	}

	*found = true

	if d := db.Get(k); len(d) > 0 {
		*rres = d
	}

	if c.keepErrors {
		if e := eb.Get(k); len(e) > 0 {
			*rerr = errors.New(string(e))
		}
	}

	return nil
}

func (c *Cache) getWorkerLock(key string) *sync.RWMutex {
	c.runningLock.Lock()
	defer c.runningLock.Unlock()

	if l, ok := c.running[key]; ok {
		c.runningRefs[l] = c.runningRefs[l] + 1
		return l
	}

	l := &sync.RWMutex{}

	c.running[key] = l
	c.runningRefs[l] = 1

	return l
}

func (c *Cache) clearWorkerLock(key string, l *sync.RWMutex) {
	c.runningLock.Lock()
	defer c.runningLock.Unlock()

	if n, ok := c.runningRefs[l]; ok {
		c.runningRefs[l] = n - 1
	}

	if c.runningRefs[l] == 0 {
		delete(c.runningRefs, l)

		if c.running[key] == l {
			delete(c.running, key)
		}
	}
}

func (c *Cache) increment(k []byte, t time.Time) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		var m Meta

		b := tx.Bucket([]byte(c.name + "#meta"))

		d := b.Get(k)
		if len(d) == 0 {
			return nil
		}

		if err := m.decode(d); err != nil {
			return err
		}

		m.AccessCount++
		if n := uint64(t.UnixNano()); n > m.AccessedAt {
			m.AccessedAt = n
		}

		return b.Put(k, m.encode())
	})
}

func withRLock(m *sync.RWMutex, fn func() error) error {
	m.RLock()
	defer m.RUnlock()
	return fn()
}

func withLock(m *sync.RWMutex, fn func() error) error {
	m.Lock()
	defer m.Unlock()
	return fn()
}

func (c *Cache) GetAt(key string, t time.Time, userdata interface{}) ([]byte, bool, error) {
	if err := c.ensureOpen(); err != nil {
		return nil, false, err
	}

	// sha1, more like shafun

	h := sha1.New()
	h.Write([]byte(key))
	k := h.Sum(nil)

	var m Meta
	var rres []byte
	var rerr error
	var found bool

	// first try to read the result - this should not block any other simple
	// read, but will be blocked by updates. this is desired behaviour.
	if err := c.db.View(func(tx *bolt.Tx) error {
		return c.readJobState(tx, k, t, &m, &rres, &rerr, &found)
	}); err != nil {
		return nil, false, err
	}

	if found {
		if err := c.increment(k, t); err != nil {
			return nil, false, err
		}

		return rres, false, rerr
	}

	// the result wasn't easily available - now we need to try to run the worker
	// function. the `getWorkerLock` and `clearWorkerLock` form a reference-
	// counting pair that will remove the lock entirely once no workers are
	// using it.

	l := c.getWorkerLock(key)
	defer c.clearWorkerLock(key, l)

	l.Lock()
	defer l.Unlock()

	// the value might have been fetched and saved while we were waiting for the
	// lock.

	if err := c.db.View(func(tx *bolt.Tx) error {
		return c.readJobState(tx, k, t, &m, &rres, &rerr, &found)
	}); err != nil {
		return nil, false, err
	}

	if found {
		if err := c.increment(k, t); err != nil {
			return nil, false, err
		}

		return rres, false, rerr
	}

	// if we don't have the value by now, it means we'll be the only one working
	// with this lock. other potential workers will be blocked above at
	// `l.Lock`, and will see the value we put into the db when they get to
	// `c.db.View`.

	copy(m.Hash[:], k[0:20])
	m.CreatedAt = uint64(t.UnixNano())
	m.AccessedAt = uint64(t.UnixNano())
	m.AccessCount = 0

	rres, rerr = c.worker(key, userdata)
	if !c.keepErrors && rerr != nil {
		return nil, false, rerr
	}

	if err := c.db.Update(func(tx *bolt.Tx) error {
		mb := tx.Bucket([]byte(c.name + "#meta"))
		db := tx.Bucket([]byte(c.name + "#data"))
		eb := tx.Bucket([]byte(c.name + "#errs"))

		if err := mb.Put(k, m.encode()); err != nil {
			return err
		}
		if err := db.Put(k, rres); err != nil {
			return err
		}
		if c.keepErrors {
			if rerr == nil {
				if err := eb.Delete(k); err != nil {
					return err
				}
			} else {
				if err := eb.Put(k, []byte(rerr.Error())); err != nil {
					return err
				}
			}
		}

		if c.highMark != 0 && mb.Stats().KeyN >= c.highMark {
			if err := c.cleanup(tx, t); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, false, err
	}

	return rres, true, rerr
}

func (c *Cache) Purge(key string) error {
	if err := c.ensureOpen(); err != nil {
		return err
	}

	h := sha1.New()
	h.Write([]byte(key))
	k := h.Sum(nil)

	return c.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket([]byte(c.name + "#meta")).Delete(k); err != nil {
			return err
		}
		if err := tx.Bucket([]byte(c.name + "#data")).Delete(k); err != nil {
			return err
		}
		if err := tx.Bucket([]byte(c.name + "#errs")).Delete(k); err != nil {
			return err
		}

		return nil
	})
}

func (c *Cache) Cleanup() error {
	return c.CleanupAt(time.Now())
}

func (c *Cache) CleanupAt(t time.Time) error {
	if err := c.ensureOpen(); err != nil {
		return err
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		return c.cleanup(tx, t)
	})
}

type metaContext struct {
	t time.Time
	a []Meta
	f Strategy
}

func (m *metaContext) add(e Meta) { m.a = append(m.a, e) }

func (m *metaContext) Len() int      { return len(m.a) }
func (m *metaContext) Swap(a, b int) { m.a[a], m.a[b] = m.a[b], m.a[a] }

func (m *metaContext) Less(a, b int) bool {
	f := m.f
	if f == nil {
		f = StrategyLRU()
	}

	return f(m.t, m.a[a], m.a[b])
}

func (c *Cache) cleanup(tx *bolt.Tx, t time.Time) error {
	mb := tx.Bucket([]byte(c.name + "#meta"))
	db := tx.Bucket([]byte(c.name + "#data"))
	eb := tx.Bucket([]byte(c.name + "#errs"))

	a := metaContext{t: t, f: c.strategy}

	mb.ForEach(func(k []byte, v []byte) error {
		if len(k) != 20 {
			return errors.New("invalid key length during iteration")
		}

		var m Meta
		copy(m.Hash[:], k[0:20])
		if err := m.decode(v); err != nil {
			return err
		}

		a.add(m)

		return nil
	})

	// first evict things that have expired
	if c.maxAge != 0 {
		for i := 0; i < len(a.a); i++ {
			if time.Duration(t.UnixNano()-int64(a.a[i].CreatedAt)) > c.maxAge {
				if c.onEviction != nil {
					c.onEviction(a.a[i], t, db.Get(a.a[i].Hash[i:]), eb.Get(a.a[i].Hash[i:]))
				}

				if err := mb.Delete(a.a[i].Hash[:]); err != nil {
					return err
				}
				if err := db.Delete(a.a[i].Hash[:]); err != nil {
					return err
				}
				if err := eb.Delete(a.a[i].Hash[:]); err != nil {
					return err
				}

				a.a = append(a.a[:i+1], a.a[i+1:]...)
			}
		}
	}

	sort.Sort(&a)
	sort.Reverse(&a)

	toRemove := len(a.a) - c.lowMark
	for i := 0; i < toRemove; i++ {
		if c.onEviction != nil {
			c.onEviction(a.a[i], t, db.Get(a.a[i].Hash[i:]), eb.Get(a.a[i].Hash[i:]))
		}

		if err := mb.Delete(a.a[i].Hash[:]); err != nil {
			return err
		}
		if err := db.Delete(a.a[i].Hash[:]); err != nil {
			return err
		}
		if err := eb.Delete(a.a[i].Hash[:]); err != nil {
			return err
		}
	}

	return nil
}
