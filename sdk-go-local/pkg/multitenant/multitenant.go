package multitenant

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
)

type Multitenant interface {
	// New creates and stores a tenant client using the internal template options.
	// Returns an error if the template is missing or client creation fails.
	New(tenant string) error
	// NewFromOptions creates and stores a tenant client using the provided options.
	// Returns an error if client creation fails.
	NewFromOptions(tenant string, options *aruba.Options) error
	// Add stores an already initialized tenant client.
	Add(tenant string, client aruba.Client)
	// Get returns the tenant client and true if found, otherwise nil and false.
	Get(tenant string) (aruba.Client, bool)
	// MustGet returns the tenant client or terminates the process if not found.
	MustGet(tenant string) aruba.Client
	// GetOrNil returns the tenant client if found, otherwise nil.
	GetOrNil(tenant string) aruba.Client
	// CleanUp removes clients not used in the provided duration window.
	CleanUp(from time.Duration)
}

type entry struct {
	client    aruba.Client
	lastUsage time.Time
	lock      sync.Mutex
}

type multitenant struct {
	clients  map[string]*entry
	template *aruba.Options

	lock sync.RWMutex
}

var _ Multitenant = (*multitenant)(nil)

// New creates an empty multitenant client manager without template options.
func New() Multitenant {
	return &multitenant{
		clients: make(map[string]*entry),
	}
}

// NewWithTemplate creates an empty multitenant client manager with template options.
// The template is used by New to instantiate tenant clients.
func NewWithTemplate(template *aruba.Options) Multitenant {
	return &multitenant{
		clients:  make(map[string]*entry),
		template: template,
	}
}

func (m *multitenant) New(tenant string) error {
	if m.template == nil {
		return errors.New("template is missing - use the `NewFromOptions` method")
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	template := m.template.DeepCopy()

	c, err := aruba.NewClient(template)
	if err != nil {
		return err
	}

	m.clients[tenant] = &entry{client: c, lastUsage: time.Now()}

	return nil
}

func (m *multitenant) NewFromOptions(tenant string, options *aruba.Options) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	c, err := aruba.NewClient(options)
	if err != nil {
		return err
	}

	m.clients[tenant] = &entry{client: c, lastUsage: time.Now()}

	return nil
}

func (m *multitenant) Add(tenant string, client aruba.Client) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.clients[tenant] = &entry{client: client, lastUsage: time.Now()}
}

func (m *multitenant) Get(tenant string) (aruba.Client, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	e, ok := m.clients[tenant]

	if !ok {
		return nil, false
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	e.lastUsage = time.Now()

	return e.client, ok
}

func (m *multitenant) MustGet(tenant string) aruba.Client {
	m.lock.RLock()
	defer m.lock.RUnlock()

	e, ok := m.clients[tenant]
	if !ok {
		log.Fatalf("client for tenant '%s' not found", tenant)
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	e.lastUsage = time.Now()

	return e.client
}

func (m *multitenant) GetOrNil(tenant string) aruba.Client {
	m.lock.RLock()
	defer m.lock.RUnlock()

	e, ok := m.clients[tenant]
	if !ok {
		return nil
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	e.lastUsage = time.Now()

	return e.client
}

func (m *multitenant) CleanUp(from time.Duration) {
	if m == nil || m.clients == nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	cleanupTime := time.Now().Add(-1 * from)

	for t, e := range m.clients {
		if e == nil {
			delete(m.clients, t)
			continue
		}
		if e.client == nil {
			delete(m.clients, t)
			continue
		}
		if e.lastUsage.Before(cleanupTime) {
			delete(m.clients, t)
		}
	}
}
