package cache_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/nepeta70/ride-hailing/internal/pkg/mocks"
	. "github.com/nepeta70/ride-hailing/services/ride/internal/adapters/cache"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

// StubCountryRepo is a lean, non-mock implementation for testing
type StubCountryRepo struct {
	data      map[string]*domain.Country
	err       error
	callCount int
	mu        sync.Mutex
}

func (s *StubCountryRepo) GetAllEnabled(ctx context.Context) (map[string]*domain.Country, error) {
	s.mu.Lock()
	s.callCount++
	s.mu.Unlock()
	return s.data, s.err
}

func TestCountryCache_GetCountryByCode(t *testing.T) {
	ctx := context.Background()

	// Reusable test data
	spain := &domain.Country{Code: "ES", Currency: "EUR"}
	uk := &domain.Country{Code: "GB", Currency: "GBP"}
	validMap := map[string]*domain.Country{"ES": spain, "GB": uk}

	tests := []struct {
		name        string
		repoData    map[string]*domain.Country
		repoErr     error
		countryCode string
		wantCountry *domain.Country
		wantFound   bool
	}{
		{
			name:        "Success: return country from lazy load",
			repoData:    validMap,
			countryCode: "ES",
			wantCountry: spain,
			wantFound:   true,
		},
		{
			name:        "Fail: country code missing in map",
			repoData:    validMap,
			countryCode: "US",
			wantCountry: nil,
			wantFound:   false,
		},
		{
			name:        "Fail: repository returns error",
			repoErr:     errors.New("database connection failed"),
			countryCode: "ES",
			wantCountry: nil,
			wantFound:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup stub
			stub := &StubCountryRepo{data: tt.repoData, err: tt.repoErr}
			cache := NewCountryCache(stub, &mocks.MockLogger{})

			gotCountry, gotFound := cache.GetCountryByCode(ctx, tt.countryCode)

			assert.Equal(t, tt.wantFound, gotFound)
			assert.Equal(t, tt.wantCountry, gotCountry)
		})
	}
}

func TestCountryCache_ConcurrencyAndOnce(t *testing.T) {
	ctx := context.Background()

	// Real data stub
	stub := &StubCountryRepo{
		data: map[string]*domain.Country{"ES": {Code: "ES", Currency: "EUR"}},
	}
	cache := NewCountryCache(stub, &mocks.MockLogger{})

	const workers = 100
	var wg sync.WaitGroup
	wg.Add(workers)

	// Fire 100 simultaneous requests
	for range workers {
		go func() {
			defer wg.Done()
			_, _ = cache.GetCountryByCode(ctx, "ES")
		}()
	}

	wg.Wait()

	// PROOF of sync.Once: Repo should have been called exactly once
	assert.Equal(t, 1, stub.callCount, "Database should only be hit once despite 100 requests")
}
