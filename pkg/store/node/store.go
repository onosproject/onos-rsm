// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package nodestore

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	storeutil "github.com/onosproject/onos-rsm/pkg/store"
	"sync"
)

var log = logging.GetLogger("store", "node-store")

func NewStore() Store {
	watchers := storeutil.NewWatchers()
	return &store{
		watchers: watchers,
		storage:  make(map[string]*Entry),
	}
}

type Store interface {
	Put(ctx context.Context, key string, value interface{}) (*Entry, error)

	Get(ctx context.Context, key string) (*Entry, error)

	UpdateSliceCreated(ctx context.Context, key string, value interface{}) (*Entry, error)

	UpdateSliceUpdated(ctx context.Context, key string, value interface{}) (*Entry, error)

	UpdateSliceDeleted(ctx context.Context, key string, value interface{}) (*Entry, error)

	Delete(ctx context.Context, key string) error

	Entries(ctx context.Context, ch chan<- *Entry) error

	Watch(ctx context.Context, ch chan<- storeutil.Event) error
}

type store struct {
	mu       sync.RWMutex
	watchers *storeutil.Watchers
	// key is node id
	storage map[string]*Entry
}

func (s *store) Put(ctx context.Context, key string, value interface{}) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var event storeutil.Event

	entry := &Entry{
		Key:   key,
		Value: value,
	}

	if _, ok := s.storage[key]; ok {
		return nil, errors.New(errors.AlreadyExists, "the key already exists")
	}

	event = storeutil.Event{
		Key:   key,
		Value: entry,
		Type:  Created,
	}

	s.storage[key] = entry
	s.watchers.Send(event)
	return entry, nil
}

func (s *store) Get(ctx context.Context, key string) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.storage[key]; ok {
		return v, nil
	}
	return nil, errors.New(errors.NotFound, "the entry does not exist")
}

func (s *store) UpdateSliceCreated(ctx context.Context, key string, value interface{}) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var event storeutil.Event

	entry := &Entry{
		Key:   key,
		Value: value,
	}

	if _, ok := s.storage[key]; !ok {
		return nil, errors.New(errors.NotFound, "the entry does not exist")
	}

	event = storeutil.Event{
		Key:   key,
		Value: entry,
		Type:  UpdatedSliceCreated,
	}

	s.storage[key] = entry
	s.watchers.Send(event)
	return entry, nil
}

func (s *store) UpdateSliceUpdated(ctx context.Context, key string, value interface{}) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var event storeutil.Event

	entry := &Entry{
		Key:   key,
		Value: value,
	}

	if _, ok := s.storage[key]; !ok {
		return nil, errors.New(errors.NotFound, "the entry does not exist")
	}

	event = storeutil.Event{
		Key:   key,
		Value: entry,
		Type:  UpdatedSliceUpdated,
	}

	s.storage[key] = entry
	s.watchers.Send(event)
	return entry, nil
}

func (s *store) UpdateSliceDeleted(ctx context.Context, key string, value interface{}) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var event storeutil.Event

	entry := &Entry{
		Key:   key,
		Value: value,
	}

	if _, ok := s.storage[key]; !ok {
		return nil, errors.New(errors.NotFound, "the entry does not exist")
	}

	event = storeutil.Event{
		Key:   key,
		Value: entry,
		Type:  UpdatedSliceDeleted,
	}

	s.storage[key] = entry
	s.watchers.Send(event)
	return entry, nil
}

func (s *store) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storage[key]; !ok {
		return fmt.Errorf("key %s is not in the store", key)
	}
	delete(s.storage, key)
	s.watchers.Send(storeutil.Event{
		Key:   key,
		Value: s.storage[key],
		Type:  Deleted,
	})
	return nil
}

func (s *store) Entries(ctx context.Context, ch chan<- *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range s.storage {
		ch <- entry
	}

	close(ch)
	return nil
}

func (s *store) Watch(ctx context.Context, ch chan<- storeutil.Event) error {
	id := uuid.New()
	err := s.watchers.AddWatcher(id, ch)
	if err != nil {
		log.Error(err)
		close(ch)
		return err
	}
	go func() {
		<-ctx.Done()
		err = s.watchers.RemoveWatcher(id)
		if err != nil {
			log.Error(err)
		}
		close(ch)
	}()
	return nil
}

var _ Store = &store{}
