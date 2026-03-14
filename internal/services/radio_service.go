package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ifinu/radio-api/internal/cache"
	"github.com/ifinu/radio-api/internal/models"
	"github.com/ifinu/radio-api/internal/repository"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

const (
	RadioBrowserAPIBase = "https://de1.api.radio-browser.info/json/stations"
)

type RadioService interface {
	GetByID(ctx context.Context, id uint) (*models.Radio, error)
	GetByUUID(ctx context.Context, uuid string) (*models.Radio, error)
	List(ctx context.Context, limit, offset int) ([]*models.Radio, int64, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Radio, error)
	Count(ctx context.Context) (int64, error)
	SyncRadios(ctx context.Context) error
	StartPeriodicSync(interval time.Duration)
}

type radioService struct {
	repo   repository.RadioRepository
	cache  *cache.RadioCache
	client *http.Client
	cron   *cron.Cron
}

func NewRadioService(repo repository.RadioRepository, cache *cache.RadioCache) RadioService {
	return &radioService{
		repo:  repo,
		cache: cache,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cron: cron.New(),
	}
}

func (s *radioService) GetByID(ctx context.Context, id uint) (*models.Radio, error) {
	// Check cache first
	if radio, ok := s.cache.Get(id); ok {
		log.Debug().Uint("id", id).Msg("Cache hit")
		return radio, nil
	}

	// Get from database
	radio, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(id, radio)

	return radio, nil
}

func (s *radioService) GetByUUID(ctx context.Context, uuid string) (*models.Radio, error) {
	// Get from database
	radio, err := s.repo.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(radio.ID, radio)

	return radio, nil
}

func (s *radioService) List(ctx context.Context, limit, offset int) ([]*models.Radio, int64, error) {
	radios, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return radios, count, nil
}

func (s *radioService) Search(ctx context.Context, query string, limit, offset int) ([]*models.Radio, error) {
	return s.repo.Search(ctx, query, limit, offset)
}

func (s *radioService) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

func (s *radioService) SyncRadios(ctx context.Context) error {
	log.Info().Msg("Starting radio synchronization")

	// Fetch from Radio Browser API
	stations, err := s.fetchStations(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch stations from API")
		return err
	}

	log.Info().Int("count", len(stations)).Msg("Fetched stations from API")

	// Convert to our model
	radios := make([]*models.Radio, 0, len(stations))
	for _, station := range stations {
		radio := &models.Radio{
			UUID:      station.StationUUID,
			Name:      station.Name,
			StreamURL: station.URLResolved,
			Country:   station.Country,
			Language:  station.Language,
			Tags:      station.Tags,
			Bitrate:   station.Bitrate,
			Favicon:   station.Favicon,
			Homepage:  station.Homepage,
			Active:    true,
			Listeners: 0,
		}

		// Fallback to URL if URLResolved is empty
		if radio.StreamURL == "" {
			radio.StreamURL = station.URL
		}

		radios = append(radios, radio)
	}

	// Batch upsert
	if err := s.repo.UpsertBatch(ctx, radios); err != nil {
		log.Error().Err(err).Msg("Failed to upsert radios")
		return err
	}

	// Clear cache after sync
	s.cache.Clear()

	log.Info().Int("synced", len(radios)).Msg("Radio synchronization completed")

	return nil
}

func (s *radioService) fetchStations(ctx context.Context) ([]*models.RadioBrowserStation, error) {
	// Fetch Brazilian stations as default
	url := fmt.Sprintf("%s/bycountrycodeexact/BR", RadioBrowserAPIBase)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "IFINU-Radio/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stations []*models.RadioBrowserStation
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, err
	}

	return stations, nil
}

func (s *radioService) StartPeriodicSync(interval time.Duration) {
	log.Info().Dur("interval", interval).Msg("Starting periodic radio sync")

	// Run initial sync
	go func() {
		ctx := context.Background()
		if err := s.SyncRadios(ctx); err != nil {
			log.Error().Err(err).Msg("Initial sync failed")
		}
	}()

	// Schedule periodic sync
	_, err := s.cron.AddFunc(fmt.Sprintf("@every %s", interval), func() {
		ctx := context.Background()
		if err := s.SyncRadios(ctx); err != nil {
			log.Error().Err(err).Msg("Periodic sync failed")
		}
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to schedule periodic sync")
		return
	}

	s.cron.Start()
}
