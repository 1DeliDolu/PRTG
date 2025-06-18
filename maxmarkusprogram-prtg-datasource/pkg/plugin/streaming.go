package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Performance-optimized constants
const (
	// Maximum data points to buffer in memory per stream
	DefaultBufferSize = 500

	// Maximum active streams per panel
	MaxStreamsPerPanel = 5

	// Reasonable interval limits
	MinUpdateInterval = 1000 * time.Millisecond
	MaxUpdateInterval = 60000 * time.Millisecond
	DefaultInterval   = 5000 * time.Millisecond

	// Default cache time to prevent excessive API calls
	DefaultCacheTime = 5 * time.Second

	// Initial buffer capacity for better memory allocation
	InitialBufferCapacity = 32
)

func (d *Datasource) SubscribeStream(ctx context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	d.logger.Debug("Subscribe to stream", "path", req.Path)

	// Fast path - quick prefix check
	if !strings.HasPrefix(req.Path, "prtg-stream/") {
		return &backend.SubscribeStreamResponse{Status: backend.SubscribeStreamStatusNotFound}, nil
	}

	// Parse subscription data for validation
	var query queryModel
	if err := json.Unmarshal(req.Data, &query); err != nil {
		d.logger.Error("Invalid subscription data", "error", err)
		return &backend.SubscribeStreamResponse{Status: backend.SubscribeStreamStatusPermissionDenied}, nil
	}

	// Required fields validation
	if query.SensorId == "" || (len(query.ChannelArray) == 0 && query.Channel == "") {
		d.logger.Error("Missing required fields", "sensorId", query.SensorId)
		return &backend.SubscribeStreamResponse{Status: backend.SubscribeStreamStatusPermissionDenied}, nil
	}

	// Check stream quota per panel
	panelId := fmt.Sprintf("%v", query.PanelID)
	if panelStreams := d.getStreamsByPanel(panelId); len(panelStreams) >= MaxStreamsPerPanel {
		d.logger.Warn("Maximum streams reached for panel", "panelId", panelId)
		return &backend.SubscribeStreamResponse{Status: backend.SubscribeStreamStatusPermissionDenied}, nil
	}

	// Log time range information if available
	if timeRangeInfo, err := extractTimeRangeInfo(req.Data); err == nil {
		d.logger.Debug("Stream time range",
			"windowSize", time.Duration(timeRangeInfo.To-timeRangeInfo.From)*time.Millisecond)
	}

	return &backend.SubscribeStreamResponse{Status: backend.SubscribeStreamStatusOK}, nil
}

// Fix for error with TimeRange structure
func extractTimeRangeInfo(data []byte) (struct{ From, To int64 }, error) {
	var result struct {
		TimeRange struct {
			From int64 `json:"from"`
			To   int64 `json:"to"`
		} `json:"timeRange"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return struct{ From, To int64 }{0, 0}, err
	}

	// Return a struct with the correct structure
	return struct{ From, To int64 }{
		From: result.TimeRange.From,
		To:   result.TimeRange.To,
	}, nil
}

func (d *Datasource) PublishStream(ctx context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return &backend.PublishStreamResponse{Status: backend.PublishStreamStatusPermissionDenied}, nil
}

func (d *Datasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	// Parse query model with error handling
	var query queryModel
	if err := json.Unmarshal(req.Data, &query); err != nil {
		return fmt.Errorf("failed to parse stream data: %w", err)
	}

	// Validation with early return
	if query.SensorId == "" {
		return fmt.Errorf("missing required field: sensorId")
	}

	// Use channelArray or channel, ensuring we have at least one
	channels := getChannels(query)
	if len(channels) == 0 {
		return fmt.Errorf("missing required field: channel or channelArray")
	}

	// Apply interval limits
	interval := getBoundedInterval(query.StreamInterval)

	// Generate stable stream ID
	streamID := generateStreamID(query, channels)

	// Set up time range
	timeRangeFrom, timeRangeTo := getTimeRange(query)

	// Configure optimized caching
	cacheDuration := getCacheDuration(query)

	// Use appropriate buffer size
	bufferSize := getBufferSize(query)

	// Create stream instance
	stream := createStream(query, channels, streamID, timeRangeFrom, timeRangeTo,
		interval, cacheDuration, bufferSize)

	d.logger.Info("Stream starting",
		"streamID", streamID,
		"intervalMs", interval.Milliseconds())

	// Check for existing stream - fast path
	if existingStream := d.getExistingStream(streamID); existingStream != nil {
		d.updateExistingStream(existingStream, timeRangeFrom, timeRangeTo, cacheDuration)
		return nil
	}

	// Register new stream
	d.registerNewStream(stream, streamID)

	// Run the stream
	return d.runStreamLoop(ctx, stream, query, sender, timeRangeFrom, timeRangeTo)
}

// Helper functions for better organization and testability

func getChannels(query queryModel) []string {
	if len(query.ChannelArray) > 0 {
		return query.ChannelArray
	}
	if query.Channel != "" {
		return []string{query.Channel}
	}
	return nil
}

func getBoundedInterval(requestedInterval int64) time.Duration {
	if requestedInterval <= 0 {
		return DefaultInterval
	}

	interval := time.Duration(requestedInterval) * time.Millisecond
	if interval < MinUpdateInterval {
		return MinUpdateInterval
	}
	if interval > MaxUpdateInterval {
		return MaxUpdateInterval
	}
	return interval
}

func generateStreamID(query queryModel, channels []string) string {
	channelKey := strings.Join(channels, "_")
	return fmt.Sprintf("%v_%s_%s_%s",
		query.PanelID,
		query.RefID,
		query.SensorId,
		channelKey)
}

func getTimeRange(query queryModel) (time.Time, time.Time) {
	now := time.Now()
	if query.From > 0 && query.To > 0 {
		return time.Unix(0, query.From*int64(time.Millisecond)),
			time.Unix(0, query.To*int64(time.Millisecond))
	}
	return now.Add(-30 * time.Minute), now
}

func getCacheDuration(query queryModel) time.Duration {
	if query.StreamInterval > 0 {
		// Use half the stream interval as cache time
		cacheDuration := time.Duration(query.StreamInterval/2) * time.Millisecond
		// With reasonable minimum
		if cacheDuration < time.Second {
			return time.Second
		}
		return cacheDuration
	}
	return DefaultCacheTime
}

// Fix for the BufferSize field that doesn't exist in queryModel
func getBufferSize(_ queryModel) int64 {
	// Since queryModel doesn't have BufferSize field, always use the default
	return DefaultBufferSize
}

func createStream(query queryModel, channels []string, streamID string,
	fromTime, toTime time.Time, interval, cacheDuration time.Duration, bufferSize int64) *activeStream {

	// Create stream with optimized initial settings
	stream := &activeStream{
		sensorId:          query.SensorId,
		channelArray:      channels,
		interval:          interval,
		group:             query.Group,
		device:            query.Device,
		sensor:            query.Sensor,
		includeGroupName:  query.IncludeGroupName,
		includeDeviceName: query.IncludeDeviceName,
		includeSensorName: query.IncludeSensorName,
		fromTime:          fromTime,
		toTime:            toTime,
		lastUpdate:        time.Now().Add(-cacheDuration), // Ensure immediate first update
		cacheTime:         cacheDuration,
		isActive:          true,
		refID:             query.RefID,
		streamID:          streamID,
		panelId:           fmt.Sprintf("%v", query.PanelID),
		queryId:           query.RefID,
		multiChannelKey:   strings.Join(channels, "_"),
		channelStates:     make(map[string]*channelState, len(channels)), // Preallocate map
		updateChan:        make(chan struct{}, 1),
		updateMode:        query.UpdateMode, // Now this will reference the correct field
		bufferSize:        bufferSize,
		status: &streamStatus{
			active:    true,
			updating:  false,
			lastError: nil,
		},
		lastDataTimestamp: time.Now().UnixMilli(),
	}

	// Default updateMode if not specified
	if stream.updateMode == "" {
		stream.updateMode = "append"
	}

	// Initialize channel states with optimized initial capacity
	for _, channelName := range channels {
		stream.channelStates[channelName] = &channelState{
			lastValue: 0,
			isActive:  true,
			buffer: &dataBuffer{
				times:  make([]time.Time, 0, InitialBufferCapacity),
				values: make([]float64, 0, InitialBufferCapacity),
				size:   bufferSize,
			},
		}
	}

	return stream
}

func (d *Datasource) getExistingStream(streamID string) *activeStream {
	d.streamManager.mu.RLock()
	existingStream, exists := d.streamManager.streams[streamID]
	d.streamManager.mu.RUnlock()

	if !exists {
		return nil
	}
	return existingStream
}

func (d *Datasource) updateExistingStream(stream *activeStream, from time.Time, to time.Time, cacheDuration time.Duration) {
	d.streamManager.mu.Lock()
	stream.timeRange = &backend.TimeRange{From: from, To: to}
	stream.isActive = true
	stream.lastUpdate = time.Now().Add(-cacheDuration) // Force immediate update
	d.streamManager.mu.Unlock()

	// Trigger update without blocking
	select {
	case stream.updateChan <- struct{}{}:
	default:
	}
}

func (d *Datasource) registerNewStream(stream *activeStream, streamID string) {
	d.streamManager.mu.Lock()
	d.streamManager.streams[streamID] = stream
	d.streamManager.mu.Unlock()

	d.trackStream(stream.panelId, streamID, stream)
}

func (d *Datasource) runStreamLoop(
	ctx context.Context,
	stream *activeStream,
	query queryModel,
	sender *backend.StreamSender,
	timeRangeFrom, timeRangeTo time.Time,
) error {
	// Set up stream cleanup
	defer d.cleanupStream(stream)

	// Create time range object for metrics query
	timeRange := backend.TimeRange{
		From: timeRangeFrom,
		To:   timeRangeTo,
	}

	// Add jitter to prevent synchronized API calls
	jitter := time.Duration(rand.Int63n(250)) * time.Millisecond
	ticker := time.NewTicker(stream.interval + jitter)
	defer ticker.Stop()

	// Get initial data immediately
	if err := d.updateStreamWithMetricsQuery(ctx, stream, sender, query, timeRange); err != nil {
		d.logger.Error("Initial stream update failed", "error", err)
	}

	// Main stream loop with controlled update frequency
	return d.streamUpdateLoop(ctx, stream, sender, query, timeRange, ticker)
}

func (d *Datasource) cleanupStream(stream *activeStream) {
	d.streamManager.mu.Lock()
	delete(d.streamManager.streams, stream.streamID)

	if panelStreams, exists := d.streamManager.activeStreams[stream.panelId]; exists {
		delete(panelStreams, stream.streamID)
		if len(panelStreams) == 0 {
			delete(d.streamManager.activeStreams, stream.panelId)
		}
	}
	d.streamManager.mu.Unlock()

	d.logger.Debug("Stream closed", "streamID", stream.streamID)
}

func (d *Datasource) streamUpdateLoop(
	ctx context.Context,
	stream *activeStream,
	sender *backend.StreamSender,
	query queryModel,
	timeRange backend.TimeRange,
	ticker *time.Ticker,
) error {
	for {
		select {
		case <-ctx.Done():
			stream.isActive = false
			return nil

		case <-ticker.C:
			// Use sliding window approach for time range
			if stream.updateMode == "sliding" {
				now := time.Now()
				windowSize := timeRange.To.Sub(timeRange.From)
				timeRange.From = now.Add(-windowSize)
				timeRange.To = now
			}

			// Create limited context for update
			updateCtx, cancel := context.WithTimeout(ctx, stream.interval/2)
			err := d.updateStreamWithMetricsQuery(updateCtx, stream, sender, query, timeRange)
			cancel()

			// Handle errors with exponential backoff logging
			d.handleStreamError(stream, err)

		case <-stream.updateChan:
			if err := d.updateStreamWithMetricsQuery(ctx, stream, sender, query, timeRange); err != nil {
				d.logger.Error("Manual update failed", "error", err)
			}
		}
	}
}

func (d *Datasource) handleStreamError(stream *activeStream, err error) {
	if err != nil {
		stream.errorCount++
		// Log less frequently as error count increases
		if stream.errorCount <= 3 || stream.errorCount%10 == 0 {
			d.logger.Error("Stream update failed",
				"error", err,
				"count", stream.errorCount,
				"streamID", stream.streamID)
		}
	} else {
		stream.errorCount = 0
	}
}

// updateStreamWithMetricsQuery uses the standard metrics query handler to avoid code duplication
func (d *Datasource) updateStreamWithMetricsQuery(
	ctx context.Context,
	stream *activeStream,
	sender *backend.StreamSender,
	query queryModel,
	timeRange backend.TimeRange,
) error {
	// Respect cache time
	if time.Since(stream.lastUpdate) < stream.cacheTime {
		return nil
	}

	stream.status.updating = true
	defer func() { stream.status.updating = false }()

	// Update lastUpdate timestamp immediately to prevent racing updates
	stream.lastUpdate = time.Now()

	// Create unique frame name for metrics query
	baseFrameName := fmt.Sprintf("stream_%s", stream.refID)

	// Get metrics data using existing query handler
	response := d.handleMetricsQuery(ctx, query, timeRange, baseFrameName)

	// Check for errors
	if response.Error != nil {
		stream.status.lastError = response.Error
		return response.Error
	}

	// No frames returned case
	if len(response.Frames) == 0 {
		return nil
	}

	// Process each frame
	return d.processResponseFrames(stream, sender, response, timeRange)
}

func (d *Datasource) processResponseFrames(
	stream *activeStream,
	sender *backend.StreamSender,
	response backend.DataResponse,
	timeRange backend.TimeRange,
) error {
	for _, frame := range response.Frames {
		// Skip invalid frames
		if len(frame.Fields) < 2 {
			continue
		}

		// Extract channel name
		channelName := extractChannelName(frame)
		if channelName == "" {
			continue
		}

		// Get channel state
		channelState, exists := stream.channelStates[channelName]
		if !exists {
			continue
		}

		// Extract data
		times, values := extractFrameData(frame)
		if len(times) == 0 {
			continue
		}

		// Update buffer
		updateChannelBuffer(stream, channelState, times, values)

		// Create streaming frame
		streamFrame := createStreamingFrame(stream, channelName, channelState, timeRange.From, timeRange.To)

		// Send frame
		if err := sender.SendFrame(streamFrame, data.IncludeAll); err != nil {
			d.logger.Error("Failed to send frame", "error", err, "channel", channelName)
			continue
		}
	}

	// Mark successful update time
	stream.lastDataTimestamp = time.Now().UnixMilli()
	return nil
}

// Helper functions for frame processing

func extractChannelName(frame *data.Frame) string {
	// First try metadata
	if frame.Meta != nil && frame.Meta.Custom != nil {
		if metaMap, ok := frame.Meta.Custom.(map[string]interface{}); ok {
			if ch, exists := metaMap["channel"]; exists {
				return fmt.Sprint(ch)
			}
		}
	}

	// Then try frame name parsing
	parts := strings.Split(frame.Name, "_")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}

	return ""
}

func extractFrameData(frame *data.Frame) ([]time.Time, []float64) {
	if len(frame.Fields) < 2 || frame.Fields[0].Len() == 0 {
		return nil, nil
	}

	timeField := frame.Fields[0]
	valueField := frame.Fields[1]

	times := make([]time.Time, timeField.Len())
	values := make([]float64, valueField.Len())

	// Extract with type checking
	for i := 0; i < timeField.Len(); i++ {
		if t, ok := timeField.At(i).(time.Time); ok {
			times[i] = t
		}
		if v, ok := valueField.At(i).(float64); ok {
			values[i] = v
		}
	}

	return times, values
}

func updateChannelBuffer(stream *activeStream, state *channelState, times []time.Time, values []float64) {
	// Update last value
	if len(values) > 0 {
		state.lastValue = values[len(values)-1]
	}

	// Update buffer based on mode
	if stream.updateMode == "append" {
		if len(times) > 0 {
			appendBufferData(state.buffer, times, values, stream.bufferSize)
		}
	} else {
		state.buffer.times = times
		state.buffer.values = values

		// Ensure buffer size limits
		if int64(len(state.buffer.times)) > state.buffer.size {
			excess := int64(len(state.buffer.times)) - state.buffer.size
			state.buffer.times = state.buffer.times[excess:]
			state.buffer.values = state.buffer.values[excess:]
		}
	}
}

// Optimized buffer data append
func appendBufferData(buffer *dataBuffer, newTimes []time.Time, newValues []float64, maxSize int64) {
	curLen := len(buffer.times)
	newLen := curLen + len(newTimes)

	// Fast path: if new data is larger than max size, just keep the newest portion
	if int64(len(newTimes)) >= maxSize {
		startIdx := len(newTimes) - int(maxSize)
		buffer.times = newTimes[startIdx:]
		buffer.values = newValues[startIdx:]
		return
	}

	// If combined data exceeds max size, trim old data first
	if int64(newLen) > maxSize {
		excess := int64(newLen) - maxSize
		buffer.times = buffer.times[int(excess):]
		buffer.values = buffer.values[int(excess):]
	}

	// Now append new data
	buffer.times = append(buffer.times, newTimes...)
	buffer.values = append(buffer.values, newValues...)
}

// Helper function to create a streaming frame with proper metadata
func createStreamingFrame(stream *activeStream, channelName string, state *channelState, from, to time.Time) *data.Frame {
	// Build display name
	displayName := buildDisplayName(stream, channelName)

	// Create frame with buffer data
	frameName := fmt.Sprintf("stream_%s_%s", stream.sensorId, channelName)
	frame := data.NewFrame(frameName,
		data.NewField("Time", nil, state.buffer.times),
		data.NewField("Value", nil, state.buffer.values).SetConfig(&data.FieldConfig{
			DisplayName: displayName,
		}),
	)

	// Set optimized metadata for live indicators
	now := time.Now().UnixMilli()
	streamingStatus := map[string]interface{}{
		"active":      true,
		"lastUpdate":  now,
		"lastValue":   state.lastValue,
		"dataPoints":  len(state.buffer.times),
		"streamId":    stream.streamID,
		"sensorId":    stream.sensorId,
		"channelName": channelName,
		"isLive":      true,
		"state":       "streaming",
	}

	// Frame metadata with streaming indicators
	frame.Meta = &data.FrameMeta{
		Type: data.FrameTypeTimeSeriesMulti,
		Custom: map[string]interface{}{
			"from":           from.UnixMilli(),
			"to":             to.UnixMilli(),
			"channel":        channelName,
			"updating":       true,
			"streaming":      true, // Required for streaming indicators
			"live":           true, // Shows live status
			"streaming_rate": stream.interval.Milliseconds(),
			"isActive":       true,
			"stable":         true,
			"timezone":       "UTC",
			"state":          "streaming",
			"streamStatus":   streamingStatus,
		},
	}

	return frame
}

// Helper for building consistent display names
func buildDisplayName(stream *activeStream, channelName string) string {
	displayName := channelName

	if stream.includeGroupName && stream.group != "" {
		displayName = fmt.Sprintf("%s - %s", stream.group, displayName)
	}
	if stream.includeDeviceName && stream.device != "" {
		displayName = fmt.Sprintf("%s - %s", stream.device, displayName)
	}
	if stream.includeSensorName && stream.sensor != "" {
		displayName = fmt.Sprintf("%s - %s", stream.sensor, displayName)
	}

	return displayName
}