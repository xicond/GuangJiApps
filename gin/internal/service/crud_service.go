package service

import (
	"fmt"
	"sync"
	"time"

	"guangjiapps/gin/internal/domain"
)

type CRUDService struct {
	mu        sync.RWMutex
	resources map[string][]domain.Resource
}

func NewCRUDService() *CRUDService {
	return &CRUDService{resources: map[string][]domain.Resource{}}
}

func (s *CRUDService) List(resource string) []domain.Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Resource(nil), s.resources[resource]...)
}

func (s *CRUDService) Create(resource string, payload domain.Resource) (domain.Resource, error) {
	if resource == "" || payload.Code == "" || payload.Name == "" {
		return domain.Resource{}, fmt.Errorf("resource, code and name are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	payload.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	payload.Category = resource
	payload.Active = true
	payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	payload.UpdatedAt = payload.CreatedAt
	payload.CreatedBy = "system"
	payload.UpdatedBy = "system"

	s.resources[resource] = append(s.resources[resource], payload)
	return payload, nil
}

func (s *CRUDService) Update(resource string, id string, payload domain.Resource) (domain.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.resources[resource]
	for i, item := range items {
		if item.ID == id {
			items[i].Code = payload.Code
			items[i].Name = payload.Name
			items[i].Active = payload.Active
			items[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			items[i].UpdatedBy = "system"
			s.resources[resource] = items
			return items[i], nil
		}
	}

	return domain.Resource{}, fmt.Errorf("resource %s with id %s not found", resource, id)
}

func (s *CRUDService) Delete(resource string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.resources[resource]
	for i, item := range items {
		if item.ID == id {
			s.resources[resource] = append(items[:i], items[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("resource %s with id %s not found", resource, id)
}
