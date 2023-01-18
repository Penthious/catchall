package adapters

import (
	"fmt"
	"github.com/mailgun/catchall"
	"github.com/penthious/catchall/business/models"
	"github.com/penthious/catchall/business/ports"
	"sync"
)

var _ ports.DB = MemoryRepo{}
var mut sync.RWMutex

// NewMemoryRepo returns a new MemoryRepo.
// We create a new map to act as the DB for the application, we also create a mutex
// to handle concurrent access to the map. Without it we would have a race condition on the map.
// I do not want to put the mutex on the MemoryRepo struct because I dont want to have to use pointer semantics, nor
// do I want to change the function signature to return a pointer.
func NewMemoryRepo() MemoryRepo {
	d := make(map[string]models.Domain)
	mut = sync.RWMutex{}
	return MemoryRepo{
		Storage: d,
	}
}

type MemoryRepo struct {
	Storage map[string]models.Domain
}

// Query searches the map for the domain and returns the domain if found.
func (mr MemoryRepo) Query(domain string) (models.Domain, error) {
	mut.RLock()
	defer mut.RUnlock()
	return mr.Storage[domain], nil
}

// Insert adds the domain to the map and increments the count based on the event type.
func (mr MemoryRepo) Insert(event catchall.Event) error {
	mut.Lock()
	defer mut.Unlock()
	if _, ok := mr.Storage[event.Domain]; !ok {
		mr.Storage[event.Domain] = models.Domain{}
	}

	domain, err := increment(mr, event)
	if err != nil {
		return fmt.Errorf("error incrementing domain: %w", err)
	}

	mr.Storage[event.Domain] = domain

	return nil
}

func increment(mr MemoryRepo, event catchall.Event) (models.Domain, error) {
	switch event.Type {
	case catchall.TypeBounced:
		return models.Domain{
			Bounced:   mr.Storage[event.Domain].Bounced + 1,
			Delivered: mr.Storage[event.Domain].Delivered,
		}, nil
	case catchall.TypeDelivered:
		return models.Domain{
			Bounced:   mr.Storage[event.Domain].Bounced,
			Delivered: mr.Storage[event.Domain].Delivered + 1,
		}, nil

	default:
		return models.Domain{}, fmt.Errorf("unknown status: %s", event.Type)
	}
}
