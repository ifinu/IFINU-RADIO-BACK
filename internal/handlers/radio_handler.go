package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/radio-api/internal/models"
	"github.com/ifinu/radio-api/internal/services"
	"github.com/ifinu/radio-api/internal/stream"
	"github.com/rs/zerolog/log"
)

type RadioHandler struct {
	service  services.RadioService
	streamer *stream.Streamer
}

func NewRadioHandler(service services.RadioService, streamer *stream.Streamer) *RadioHandler {
	return &RadioHandler{
		service:  service,
		streamer: streamer,
	}
}

// ListRadios godoc
// @Summary List radios
// @Description Get paginated list of radios
// @Tags radios
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ListRadiosResponse
// @Router /radios [get]
func (h *RadioHandler) ListRadios(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	radios, total, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list radios")
		c.JSON(http.StatusInternalServerError, gin.H{
			"sucesso": false,
			"mensagem": "Failed to fetch radios",
		})
		return
	}

	// Convert to DTOs
	dtos := make([]interface{}, len(radios))
	for i, radio := range radios {
		dtos[i] = radio.ToDTO()
	}

	c.JSON(http.StatusOK, gin.H{
		"sucesso": true,
		"dados":   dtos,
		"total":   total,
	})
}

// SearchRadios godoc
// @Summary Search radios
// @Description Search radios by name, country or tags
// @Tags radios
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} []models.Radio
// @Router /radios/search [get]
func (h *RadioHandler) SearchRadios(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	radios, err := h.service.Search(c.Request.Context(), query, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("query", query).Msg("Failed to search radios")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search radios"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  radios,
		"query": query,
	})
}

// GetRadio godoc
// @Summary Get radio by ID
// @Description Get radio details by ID
// @Tags radios
// @Accept json
// @Produce json
// @Param id path int true "Radio ID"
// @Success 200 {object} models.Radio
// @Router /radios/{id} [get]
func (h *RadioHandler) GetRadio(c *gin.Context) {
	// Try to get by UUID first, fallback to numeric ID
	idParam := c.Param("id")

	// Check if it's a UUID or numeric ID
	var radio *models.Radio
	var err error

	if id, parseErr := strconv.ParseUint(idParam, 10, 32); parseErr == nil {
		radio, err = h.service.GetByID(c.Request.Context(), uint(id))
	} else {
		// Try UUID
		radio, err = h.service.GetByUUID(c.Request.Context(), idParam)
	}

	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Failed to get radio")
		c.JSON(http.StatusNotFound, gin.H{
			"sucesso": false,
			"mensagem": "Radio not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sucesso": true,
		"dados": radio.ToDTO(),
	})
}

// StreamRadio godoc
// @Summary Stream radio
// @Description Proxy stream from radio station
// @Tags radios
// @Accept json
// @Produce audio/mpeg
// @Param id path int true "Radio ID"
// @Router /radios/{id}/stream [get]
func (h *RadioHandler) StreamRadio(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid radio ID"})
		return
	}

	radio, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Msg("Failed to get radio for streaming")
		c.JSON(http.StatusNotFound, gin.H{"error": "Radio not found"})
		return
	}

	if radio.StreamURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Radio has no stream URL"})
		return
	}

	log.Info().
		Uint("radio_id", radio.ID).
		Str("radio_name", radio.Name).
		Str("stream_url", radio.StreamURL).
		Msg("Starting stream proxy")

	// Stream the radio
	if err := h.streamer.Stream(c.Request.Context(), c.Writer, radio.StreamURL); err != nil {
		log.Error().
			Err(err).
			Uint("radio_id", radio.ID).
			Msg("Stream failed")
		// Don't send JSON if we've already started streaming
		return
	}
}

// Health godoc
// @Summary Health check
// @Description Health check endpoint
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *RadioHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
