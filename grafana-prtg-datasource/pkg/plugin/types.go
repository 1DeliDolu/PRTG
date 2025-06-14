package plugin

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

/* =================================== GROUP LIST RESPONSE ======================================== */
type PrtgGroupListResponse struct {
	PrtgVersion string                    `json:"prtg-version"`
	TreeSize    int64                     `json:"treesize"`
	Groups      []PrtgGroupListItemStruct `json:"groups"`
}

type PrtgGroupListItemStruct struct {
	Active         bool    `json:"active"`
	ActiveRAW      int     `json:"active_raw"`
	Channel        string  `json:"channel"`
	ChannelRAW     string  `json:"channel_raw"`
	Datetime       string  `json:"datetime"`
	DatetimeRAW    float64 `json:"datetime_raw"`
	Device         string  `json:"device"`
	DeviceRAW      string  `json:"device_raw"`
	Downsens       string  `json:"downsens"`
	DownsensRAW    int     `json:"downsens_raw"`
	Group          string  `json:"group"`
	GroupRAW       string  `json:"group_raw"`
	Message        string  `json:"message"`
	MessageRAW     string  `json:"message_raw"`
	ObjectId       int64   `json:"objid"`
	ObjectIdRAW    int64   `json:"objid_raw"`
	Pausedsens     string  `json:"pausedsens"`
	PausedsensRAW  int     `json:"pausedsens_raw"`
	Priority       string  `json:"priority"`
	PriorityRAW    int     `json:"priority_raw"`
	Sensor         string  `json:"sensor"`
	SensorRAW      string  `json:"sensor_raw"`
	Status         string  `json:"status"`
	StatusRAW      int     `json:"status_raw"`
	Tags           string  `json:"tags"`
	TagsRAW        string  `json:"tags_raw"`
	Totalsens      string  `json:"totalsens"`
	TotalsensRAW   int     `json:"totalsens_raw"`
	Unusualsens    string  `json:"unusualsens"`
	UnusualsensRAW int     `json:"unusualsens_raw"`
	Upsens         string  `json:"upsens"`
	UpsensRAW      int     `json:"upsens_raw"`
	Warnsens       string  `json:"warnsens"`
	WarnsensRAW    int     `json:"warnsens_raw"`
}

/* =================================== DEVICE LIST RESPONSE ====================================== */
type PrtgDevicesListResponse struct {
	PrtgVersion string                     `json:"prtg-version"`
	TreeSize    int64                      `json:"treesize"`
	Devices     []PrtgDeviceListItemStruct `json:"devices"`
}

type PrtgDeviceListItemStruct struct {
	Active         bool    `json:"active"`
	ActiveRAW      int     `json:"active_raw"`
	Channel        string  `json:"channel"`
	ChannelRAW     string  `json:"channel_raw"`
	Datetime       string  `json:"datetime"`
	DatetimeRAW    float64 `json:"datetime_raw"`
	Device         string  `json:"device"`
	DeviceRAW      string  `json:"device_raw"`
	Downsens       string  `json:"downsens"`
	DownsensRAW    int     `json:"downsens_raw"`
	Group          string  `json:"group"`
	GroupRAW       string  `json:"group_raw"`
	Message        string  `json:"message"`
	MessageRAW     string  `json:"message_raw"`
	ObjectId       int64   `json:"objid"`
	ObjectIdRAW    int64   `json:"objid_raw"`
	Pausedsens     string  `json:"pausedsens"`
	PausedsensRAW  int     `json:"pausedsens_raw"`
	Priority       string  `json:"priority"`
	PriorityRAW    int     `json:"priority_raw"`
	Sensor         string  `json:"sensor"`
	SensorRAW      string  `json:"sensor_raw"`
	Status         string  `json:"status"`
	StatusRAW      int     `json:"status_raw"`
	Tags           string  `json:"tags"`
	TagsRAW        string  `json:"tags_raw"`
	Totalsens      string  `json:"totalsens"`
	TotalsensRAW   int     `json:"totalsens_raw"`
	Unusualsens    string  `json:"unusualsens"`
	UnusualsensRAW int     `json:"unusualsens_raw"`
	Upsens         string  `json:"upsens"`
	UpsensRAW      int     `json:"upsens_raw"`
	Warnsens       string  `json:"warnsens"`
	WarnsensRAW    int     `json:"warnsens_raw"`
}

/* =================================== SENSOR LIST RESPONSE ===================================== */
type PrtgSensorsListResponse struct {
	PrtgVersion string                     `json:"prtg-version"`
	TreeSize    int64                      `json:"treesize"`
	Sensors     []PrtgSensorListItemStruct `json:"sensors"`
}

// Mixed type for handling both string and number values
type StringOrNumber struct {
	String string
}

func (s *StringOrNumber) UnmarshalJSON(data []byte) error {
	// Try string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.String = str
		return nil
	}

	// Try number if string fails
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		s.String = fmt.Sprintf("%v", num)
		return nil
	}

	return fmt.Errorf("value must be string or number")
}

type PrtgSensorListItemStruct struct {
	Active         bool           `json:"active"`
	ActiveRAW      int            `json:"active_raw"`
	Channel        string         `json:"channel"`
	ChannelRAW     StringOrNumber `json:"channel_raw"` // Changed from string to StringOrNumber
	Datetime       string         `json:"datetime"`
	DatetimeRAW    float64        `json:"datetime_raw"`
	Device         string         `json:"device"`
	DeviceRAW      string         `json:"device_raw"`
	Downsens       string         `json:"downsens"`
	DownsensRAW    int            `json:"downsens_raw"`
	Group          string         `json:"group"`
	GroupRAW       string         `json:"group_raw"`
	Message        string         `json:"message"`
	MessageRAW     string         `json:"message_raw"`
	ObjectId       int64          `json:"objid"`
	ObjectIdRAW    int64          `json:"objid_raw"`
	Pausedsens     string         `json:"pausedsens"`
	PausedsensRAW  int            `json:"pausedsens_raw"`
	Priority       string         `json:"priority"`
	PriorityRAW    int            `json:"priority_raw"`
	Sensor         string         `json:"sensor"`
	SensorRAW      string         `json:"sensor_raw"`
	Status         string         `json:"status"`
	StatusRAW      int            `json:"status_raw"`
	Tags           string         `json:"tags"`
	TagsRAW        string         `json:"tags_raw"`
	Totalsens      string         `json:"totalsens"`
	TotalsensRAW   int            `json:"totalsens_raw"`
	Unusualsens    string         `json:"unusualsens"`
	UnusualsensRAW int            `json:"unusualsens_raw"`
	Upsens         string         `json:"upsens"`
	UpsensRAW      int            `json:"upsens_raw"`
	Warnsens       string         `json:"warnsens"`
	WarnsensRAW    int            `json:"warnsens_raw"`
}

/* =================================== STATUS LIST RESPONSE ===================================== */
type PrtgStatusListResponse struct {
	PrtgVersion          string `json:"prtgversion"`
	AckAlarms            string `json:"ackalarms"`
	Alarms               string `json:"alarms"`
	AutoDiscoTasks       string `json:"autodiscotasks"`
	BackgroundTasks      string `json:"backgroundtasks"`
	Clock                string `json:"clock"`
	ClusterNodeName      string `json:"clusternodename"`
	ClusterType          string `json:"clustertype"`
	CommercialExpiryDays int    `json:"commercialexpirydays"`
	CorrelationTasks     string `json:"correlationtasks"`
	DaysInstalled        int    `json:"daysinstalled"`
	EditionType          string `json:"editiontype"`
	Favs                 int    `json:"favs"`
	JsClock              int64  `json:"jsclock" `
	LowMem               bool   `json:"lowmem"`
	MaintExpiryDays      string `json:"maintexpirydays"`
	MaxSensorCount       string `json:"maxsensorcount"`
	NewAlarms            string `json:"newalarms"`
	NewMessages          string `json:"newmessages"`
	NewTickets           string `json:"newtickets"`
	Overloadprotection   bool   `json:"overloadprotection"`
	PartialAlarms        string `json:"partialalarms"`
	PausedSens           string `json:"pausedsens"`
	PRTGUpdateAvailable  bool   `json:"prtgupdateavailable"`
	ReadOnlyUser         string `json:"readonlyuser"`
	ReportTasks          string `json:"reporttasks"`
	TotalSens            int    `json:"totalsens"`
	TrialExpiryDays      int    `json:"trialexpirydays"`
	UnknownSens          string `json:"unknownsens"`
	UnusualSens          string `json:"unusualsens"`
	UpSens               string `json:"upsens"`
	Version              string `json:"version"`
	WarnSens             string `json:"warnsens"`
}

/* =================================== CHANNEL LIST RESPONSE ==================================== */
type PrtgChannelsListResponse struct {
	PrtgVersion string                   `json:"prtg-version"`
	TreeSize    int64                    `json:"treesize"`
	Values      []PrtgChannelValueStruct `json:"values"`
}

type PrtgChannelValueStruct map[string]interface{}

/* =================================== CHANNEL VALUE RESPONSE =================================== */
type PrtgHistoricalDataResponse struct {
	PrtgVersion string       `json:"prtg-version"`
	TreeSize    int64        `json:"treesize"`
	HistData    []PrtgValues `json:"histdata"`
}

type PrtgValues struct {
	Datetime string                 `json:"datetime"`
	Value    map[string]interface{} `json:"-"`
}

func (p *PrtgValues) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if dt, ok := raw["datetime"].(string); ok {
		p.Datetime = dt
	}
	delete(raw, "datetime")
	p.Value = raw
	return nil
}

/* =================================== DATASOURCE INTERFACE ==================================== */
type PRTGAPI interface {
	GetGroups() (*PrtgGroupListResponse, error)
	GetStatusList() (*PrtgStatusListResponse, error)
	GetDevices(groupId string) (*PrtgDevicesListResponse, error)
	GetSensors(deviceId string) (*PrtgSensorsListResponse, error)
	GetChannels(sensorId string) (*PrtgChannelValueStruct, error)
	GetHistoricalData(sensorId string, from time.Time, to time.Time) (*PrtgHistoricalDataResponse, error)
	ExecuteManualMethod(method string, objectId string) (*PrtgManualMethodResponse, error)
	GetAnnotationData(query *AnnotationQuery) (*AnnotationResponse, error)
	GetCacheTime() time.Duration
}

type Group struct {
	Group string `json:"group"`
}

type Device struct {
	Device string `json:"device"`
}

type Sensor struct {
	Sensor string `json:"sensor"`
}

type queryModel struct {
	QueryType         string   `json:"queryType"`
	SensorId          string   `json:"sensorId"`
	DeviceId          string   `json:"deviceId"`
	GroupId           string   `json:"groupId"`
	Group             string   `json:"group"`
	Device            string   `json:"device"`
	Sensor            string   `json:"sensor"`
	Channel           string   `json:"channel"`
	ChannelArray      []string `json:"channelArray"`
	Property          string   `json:"property"`
	FilterProperty    string   `json:"filterProperty"`
	IncludeGroupName  bool     `json:"includeGroupName"`
	IncludeDeviceName bool     `json:"includeDeviceName"`
	IncludeSensorName bool     `json:"includeSensorName"`
	From              int64    `json:"from"`
	To                int64    `json:"to"`
	ManualMethod      string   `json:"manualMethod"`
	ManualObjectId    string   `json:"manualObjectId"`
	Limit             int64    `json:"limit"`
	Tags              []string `json:"tags"`
	DashboardID       int64    `json:"dashboardId"`
	DashboardUID      string   `json:"dashboardUid"`
	PanelID           int64    `json:"panelId"`
	IsStreaming       bool     `json:"isStreaming"`
	StreamInterval    int64    `json:"streamInterval"`
	UpdateMode        string   `json:"updateMode"` // Add this field for stream update mode
	RefID             string   `json:"refId"`
}

/* =================================== DATASOURCE ============================================== */

/* =================================== CACHE ITEM ============================================== */
type cacheItem struct {
	data   []byte
	expiry time.Time
}

/* =================================== API ==================================================== */
type ApiInterface interface {
	GetCacheTime() time.Duration
	SetTimeout(timeout time.Duration)
	GetStatusList() (*PrtgStatusListResponse, error)
	GetGroups() (*PrtgGroupListResponse, error)
	GetDevices(group string) (*PrtgDevicesListResponse, error)
	GetSensors(device string) (*PrtgSensorsListResponse, error)
	GetChannels(sensorId string) (*PrtgChannelValueStruct, error)
	GetHistoricalData(sensorID string, startDate, endDate time.Time) (*PrtgHistoricalDataResponse, error)
	ExecuteManualMethod(method string, objectId string) (*PrtgManualMethodResponse, error)
	GetAnnotationData(query *AnnotationQuery) (*AnnotationResponse, error)
}

type Api struct {
	baseURL   string
	apiKey    string
	timeout   time.Duration
	cacheTime time.Duration
	cache     map[string]cacheItem
	cacheMu   sync.RWMutex
}

/* =================================== MANUAL STRUCT =========================================== */
type PrtgManualMethodResponse struct {
	Manuel    map[string]interface{} `json:"raw"`
	KeyValues []KeyValue             `json:"keyValues"`
}

type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

/* =================================== ANNOTATION STRUCTS ====================================== */
type AnnotationQuery struct {
	From         int64    `json:"from,omitempty"`  // milliseconds epoch
	To           int64    `json:"to,omitempty"`    // milliseconds epoch
	Limit        int64    `json:"limit,omitempty"` // default 100
	AlertID      int64    `json:"alertId,omitempty"`
	DashboardID  int64    `json:"dashboardId,omitempty"`
	DashboardUID string   `json:"dashboardUID,omitempty"`
	PanelID      int64    `json:"panelId,omitempty"`
	UserID       int64    `json:"userId,omitempty"`
	Type         string   `json:"type,omitempty"`     // alert or annotation
	Tags         []string `json:"tags,omitempty"`     // AND filtering
	SensorID     string   `json:"sensorId,omitempty"` // PRTG specific
}

type Annotation struct {
	ID      string                 `json:"id"` // Changed to string for UID
	Time    int64                  `json:"time"`
	TimeEnd int64                  `json:"timeEnd"`
	Title   string                 `json:"title"`
	Text    string                 `json:"text"`
	Tags    []string               `json:"tags"`
	Type    string                 `json:"type,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type AnnotationResponse struct {
	Annotations []Annotation `json:"annotations"`
	Total       int          `json:"total"`
}

type PrtgAnnotationResponse struct {
	Annotations []PrtgAnnotation `json:"annotations"`
}

type PrtgAnnotation struct {
	ID      int64    `json:"id"`
	Time    int64    `json:"time"`
	TimeEnd int64    `json:"timeEnd"`
	Text    string   `json:"text"`
	Tags    []string `json:"tags"`
}

/* =================================== QUERY CACHE ============================================== */
type QueryCacheKey struct {
	RefID      string
	QueryType  string
	SensorID   string
	Channel    string
	TimeRange  string
	Property   string
	Parameters string
}

func (k QueryCacheKey) String() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s",
		k.RefID,
		k.QueryType,
		k.SensorID,
		k.Channel,
		k.TimeRange,
		k.Property,
		k.Parameters,
	)
}

type QueryCacheEntry struct {
	Response   backend.DataResponse
	ValidUntil time.Time
	Updating   bool
}

// Add after QueryCacheEntry struct...

type streamManager struct {
	streams          map[string]*activeStream
	mu               sync.RWMutex
	defaultCacheTime time.Duration
	activeStreams    map[string]map[string]*activeStream // panelId -> streamId -> stream
}

type streamStatus struct {
	active    bool
	updating  bool
	lastError error
}

type channelState struct {
	lastValue float64
	isActive  bool
	buffer    *dataBuffer // Reference to dataBuffer type
}

// Use this dataBuffer definition and remove the one in streaming.go
type dataBuffer struct {
	times  []time.Time
	values []float64
	size   int64
}

type activeStream struct {
	sensorId          string
	channelArray      []string
	interval          time.Duration
	lastUpdate        time.Time
	group             string
	device            string
	sensor            string
	includeGroupName  bool
	includeDeviceName bool
	includeSensorName bool
	fromTime          time.Time
	toTime            time.Time
	cacheTime         time.Duration
	timeRange         *backend.TimeRange
	isActive          bool
	updateChan        chan struct{}
	status            *streamStatus
	refID             string // Add RefID for multiple streams support
	streamID          string // Unique identifier for the stream
	panelId           string // Panel identifier
	queryId           string // Query identifier within the panel
	multiChannelKey   string
	channelStates     map[string]*channelState
	updateMode        string
	bufferSize        int64
	errorCount        int
	lastDataTimestamp int64 // Track when data was last sent successfully
}

/* =================================== QUERY MODEL ============================================== */
type Datasource struct {
	baseURL       string
	api           PRTGAPI
	logger        PrtgLogger
	tracer        *Tracer
	metrics       *Metrics
	mux           backend.QueryDataHandler
	queryCache    map[string]*QueryCacheEntry
	cacheMutex    sync.RWMutex
	cacheTime     time.Duration
	streamManager *streamManager
}
