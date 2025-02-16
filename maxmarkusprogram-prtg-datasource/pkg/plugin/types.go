package plugin

import (
	"encoding/json"
	"time"
)

// PrtgTableListResponse represents the response from PRTG Table List API.
// Note: Check if "prtg-version" is delivered as array or string.
type PrtgTableListResponse struct {
	PrtgVersion []PrtgStatusListResponse   `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                      `json:"treesize" xml:"treesize"`
	Groups      []PrtgGroupListResponse    `json:"groups,omitempty" xml:"groups,omitempty"`
	Devices     []PrtgDevicesListResponse  `json:"devices,omitempty" xml:"devices,omitempty"`
	Sensors     []PrtgSensorsListResponse  `json:"sensors,omitempty" xml:"sensors,omitempty"`
	Values      []PrtgChannelsListResponse `json:"channels,omitempty" xml:"channels,omitempty"`
}

//############################# GROUP LIST RESPONSE ####################################

// PrtgGroupListResponse represents the response for groups.
type PrtgGroupListResponse struct {
	PrtgVersion string                    `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                     `json:"treesize" xml:"treesize"`
	Groups      []PrtgGroupListItemStruct `json:"groups" xml:"groups"`
}

// PrtgGroupListItemStruct contains details for a single group.
type PrtgGroupListItemStruct struct {
	Active         bool    `json:"active" xml:"active"`
	ActiveRAW      int     `json:"active_raw" xml:"active_raw"`
	Channel        string  `json:"channel" xml:"channel"`
	ChannelRAW     string  `json:"channel_raw" xml:"channel_raw"`
	Datetime       string  `json:"datetime" xml:"datetime"`
	DatetimeRAW    float64 `json:"datetime_raw" xml:"datetime_raw"`
	Device         string  `json:"device" xml:"device"`
	DeviceRAW      string  `json:"device_raw" xml:"device_raw"`
	Downsens       string  `json:"downsens" xml:"downsens"`
	DownsensRAW    int     `json:"downsens_raw" xml:"downsens_raw"`
	Group          string  `json:"group" xml:"group"`
	GroupRAW       string  `json:"group_raw" xml:"group_raw"`
	Message        string  `json:"message" xml:"message"`
	MessageRAW     string  `json:"message_raw" xml:"message_raw"`
	ObjectId       int64   `json:"objid" xml:"objid"`
	ObjectIdRAW    int64   `json:"objid_raw" xml:"objid_raw"`
	Pausedsens     string  `json:"pausedsens" xml:"pausedsens"`
	PausedsensRAW  int     `json:"pausedsens_raw" xml:"pausedsens_raw"`
	Priority       string  `json:"priority" xml:"priority"`
	PriorityRAW    int     `json:"priority_raw" xml:"priority_raw"`
	Sensor         string  `json:"sensor" xml:"sensor"`
	SensorRAW      string  `json:"sensor_raw" xml:"sensor_raw"`
	Status         string  `json:"status" xml:"status"`
	StatusRAW      int     `json:"status_raw" xml:"status_raw"`
	Tags           string  `json:"tags" xml:"tags"`
	TagsRAW        string  `json:"tags_raw" xml:"tags_raw"`
	Totalsens      string  `json:"totalsens" xml:"totalsens"`
	TotalsensRAW   int     `json:"totalsens_raw" xml:"totalsens_raw"`
	Unusualsens    string  `json:"unusualsens" xml:"unusualsens"`
	UnusualsensRAW int     `json:"unusualsens_raw" xml:"unusualsens_raw"`
	Upsens         string  `json:"upsens" xml:"upsens"`
	UpsensRAW      int     `json:"upsens_raw" xml:"upsens_raw"`
	Warnsens       string  `json:"warnsens" xml:"warnsens"`
	WarnsensRAW    int     `json:"warnsens_raw" xml:"warnsens_raw"`
}

//############################# DEVICE LIST RESPONSE ####################################

// PrtgDevicesListResponse represents the response for devices.
type PrtgDevicesListResponse struct {
	PrtgVersion string                     `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                      `json:"treesize" xml:"treesize"`
	Devices     []PrtgDeviceListItemStruct `json:"devices" xml:"devices"`
}

// PrtgDeviceListItemStruct contains details for a single device.
type PrtgDeviceListItemStruct struct {
	Active         bool    `json:"active" xml:"active"`
	ActiveRAW      int     `json:"active_raw" xml:"active_raw"`
	Channel        string  `json:"channel" xml:"channel"`
	ChannelRAW     string  `json:"channel_raw" xml:"channel_raw"`
	Datetime       string  `json:"datetime" xml:"datetime"`
	DatetimeRAW    float64 `json:"datetime_raw" xml:"datetime_raw"`
	Device         string  `json:"device" xml:"device"`
	DeviceRAW      string  `json:"device_raw" xml:"device_raw"`
	Downsens       string  `json:"downsens" xml:"downsens"`
	DownsensRAW    int     `json:"downsens_raw" xml:"downsens_raw"`
	Group          string  `json:"group" xml:"group"`
	GroupRAW       string  `json:"group_raw" xml:"group_raw"`
	Message        string  `json:"message" xml:"message"`
	MessageRAW     string  `json:"message_raw" xml:"message_raw"`
	ObjectId       int64   `json:"objid" xml:"objid"`
	ObjectIdRAW    int64   `json:"objid_raw" xml:"objid_raw"`
	Pausedsens     string  `json:"pausedsens" xml:"pausedsens"`
	PausedsensRAW  int     `json:"pausedsens_raw" xml:"pausedsens_raw"`
	Priority       string  `json:"priority" xml:"priority"`
	PriorityRAW    int     `json:"priority_raw" xml:"priority_raw"`
	Sensor         string  `json:"sensor" xml:"sensor"`
	SensorRAW      string  `json:"sensor_raw" xml:"sensor_raw"`
	Status         string  `json:"status" xml:"status"`
	StatusRAW      int     `json:"status_raw" xml:"status_raw"`
	Tags           string  `json:"tags" xml:"tags"`
	TagsRAW        string  `json:"tags_raw" xml:"tags_raw"`
	Totalsens      string  `json:"totalsens" xml:"totalsens"`
	TotalsensRAW   int     `json:"totalsens_raw" xml:"totalsens_raw"`
	Unusualsens    string  `json:"unusualsens" xml:"unusualsens"`
	UnusualsensRAW int     `json:"unusualsens_raw" xml:"unusualsens_raw"`
	Upsens         string  `json:"upsens" xml:"upsens"`
	UpsensRAW      int     `json:"upsens_raw" xml:"upsens_raw"`
	Warnsens       string  `json:"warnsens" xml:"warnsens"`
	WarnsensRAW    int     `json:"warnsens_raw" xml:"warnsens_raw"`
}

//############################# SENSOR LIST RESPONSE ####################################

// PrtgSensorsListResponse represents the response for sensors.
type PrtgSensorsListResponse struct {
	PrtgVersion string                     `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                      `json:"treesize" xml:"treesize"`
	Sensors     []PrtgSensorListItemStruct `json:"sensors" xml:"sensors"`
}

// PrtgSensorListItemStruct contains details for a single sensor.
type PrtgSensorListItemStruct struct {
	Active         bool    `json:"active" xml:"active"`
	ActiveRAW      int     `json:"active_raw" xml:"active_raw"`
	Channel        string  `json:"channel" xml:"channel"`
	ChannelRAW     int     `json:"channel_raw" xml:"channel_raw"`
	Datetime       string  `json:"datetime" xml:"datetime"`
	DatetimeRAW    float64 `json:"datetime_raw" xml:"datetime_raw"`
	Device         string  `json:"device" xml:"device"`
	DeviceRAW      string  `json:"device_raw" xml:"device_raw"`
	Downsens       string  `json:"downsens" xml:"downsens"`
	DownsensRAW    int     `json:"downsens_raw" xml:"downsens_raw"`
	Group          string  `json:"group" xml:"group"`
	GroupRAW       string  `json:"group_raw" xml:"group_raw"`
	Message        string  `json:"message" xml:"message"`
	MessageRAW     string  `json:"message_raw" xml:"message_raw"`
	ObjectId       int64   `json:"objid" xml:"objid"`
	ObjectIdRAW    int64   `json:"objid_raw" xml:"objid_raw"`
	Pausedsens     string  `json:"pausedsens" xml:"pausedsens"`
	PausedsensRAW  int     `json:"pausedsens_raw" xml:"pausedsens_raw"`
	Priority       string  `json:"priority" xml:"priority"`
	PriorityRAW    int     `json:"priority_raw" xml:"priority_raw"`
	Sensor         string  `json:"sensor" xml:"sensor"`
	SensorRAW      string  `json:"sensor_raw" xml:"sensor_raw"`
	Status         string  `json:"status" xml:"status"`
	StatusRAW      int     `json:"status_raw" xml:"status_raw"`
	Tags           string  `json:"tags" xml:"tags"`
	TagsRAW        string  `json:"tags_raw" xml:"tags_raw"`
	Totalsens      string  `json:"totalsens" xml:"totalsens"`
	TotalsensRAW   int     `json:"totalsens_raw" xml:"totalsens_raw"`
	Unusualsens    string  `json:"unusualsens" xml:"unusualsens"`
	UnusualsensRAW int     `json:"unusualsens_raw" xml:"unusualsens_raw"`
	Upsens         string  `json:"upsens" xml:"upsens"`
	UpsensRAW      int     `json:"upsens_raw" xml:"upsens_raw"`
	Warnsens       string  `json:"warnsens" xml:"warnsens"`
	WarnsensRAW    int     `json:"warnsens_raw" xml:"warnsens_raw"`
}

//############################# STATUS LIST RESPONSE ####################################

// PrtgStatusListResponse contains system-wide status information.
type PrtgStatusListResponse struct {
	PrtgVersion          string `json:"prtgversion" xml:"prtg-version"`
	AckAlarms            string `json:"ackalarms" xml:"ackalarms"`
	Alarms               string `json:"alarms" xml:"alarms"`
	AutoDiscoTasks       string `json:"autodiscotasks" xml:"autodiscotasks"`
	BackgroundTasks      string `json:"backgroundtasks" xml:"backgroundtasks"`
	Clock                string `json:"clock" xml:"clock"`
	ClusterNodeName      string `json:"clusternodename" xml:"clusternodename"`
	ClusterType          string `json:"clustertype" xml:"clustertype"`
	CommercialExpiryDays int    `json:"commercialexpirydays" xml:"commercialexpirydays"`
	CorrelationTasks     string `json:"correlationtasks" xml:"correlationtasks"`
	DaysInstalled        int    `json:"daysinstalled" xml:"daysinstalled"`
	EditionType          string `json:"editiontype" xml:"editiontype"`
	Favs                 int    `json:"favs" xml:"favs"`
	JsClock              int64  `json:"jsclock" xml:"jsclock"`
	LowMem               bool   `json:"lowmem" xml:"lowmem"`
	MaintExpiryDays      string `json:"maintexpirydays" xml:"maintexpirydays"`
	MaxSensorCount       string `json:"maxsensorcount" xml:"maxsensorcount"`
	NewAlarms            string `json:"newalarms" xml:"newalarms"`
	NewMessages          string `json:"newmessages" xml:"newmessages"`
	NewTickets           string `json:"newtickets" xml:"newtickets"`
	Overloadprotection   bool   `json:"overloadprotection" xml:"overloadprotection"`
	PartialAlarms        string `json:"partialalarms" xml:"partialalarms"`
	PausedSens           string `json:"pausedsens" xml:"pausedsens"`
	PRTGUpdateAvailable  bool   `json:"prtgupdateavailable" xml:"prtgupdateavailable"`
	ReadOnlyUser         string `json:"readonlyuser" xml:"readonlyuser"`
	ReportTasks          string `json:"reporttasks" xml:"reporttasks"`
	TotalSens            int    `json:"totalsens"`
	TrialExpiryDays      int    `json:"trialexpirydays"`
	UnknownSens          string `json:"unknownsens"`
	UnusualSens          string `json:"unusualsens"`
	UpSens               string `json:"upsens"`
	Version              string `json:"version"`
	WarnSens             string `json:"warnsens"`
}

//############################# CHANNEL LIST RESPONSE ####################################

// PrtgChannelsListResponse represents the response for channel values.
type PrtgChannelsListResponse struct {
	PrtgVersion string                   `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                    `json:"treesize" xml:"treesize"`
	Values      []PrtgChannelValueStruct `json:"-" xml:"-"`
}

// PrtgChannelValueStruct is used as a dynamic map for storing channel data.
type PrtgChannelValueStruct map[string]interface{}

// UnmarshalJSON implements a custom unmarshal method,
// which handles the "datetime" value separately and packs the rest into the Value field.
func (p *PrtgChannelValueStruct) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Create a new map if p is nil
	if *p == nil {
		*p = make(PrtgChannelValueStruct)
	}

	// Copy all values to the map
	for k, v := range raw {
		(*p)[k] = v
	}

	return nil
}

//############################# CHANNEL VALUE RESPONSE ####################################

// PrtgHistoricalDataResponse contains historical values of a sensor.
type PrtgHistoricalDataResponse struct {
	PrtgVersion string       `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64        `json:"treesize" xml:"treesize"`
	HistData    []PrtgValues `json:"histdata" xml:"histdata"`
}

// PrtgValues contains the timestamp and dynamic values.
type PrtgValues struct {
	Datetime string                 `json:"datetime"`
	Value    map[string]interface{} `json:"-"`
}

// UnmarshalJSON implements a custom unmarshal method,
// which handles the "datetime" value separately and packs the rest into the Value field.
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

/* #################################  MESSAGE RESPONSE #################################### */

// PrtgMessageResponse contains the response for messages.
type PrtgMessageResponse struct {
	PrtgVersion string                      `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                       `json:"treesize" xml:"treesize"`
	Messages    []PrtgMessageListItemStruct `json:"messages" xml:"messages"`
}

// PrtgMessageListItemStruct contains details for a single message.
type PrtgMessageListItemStruct struct {
	ObjectId    int64   `json:"objid"`
	Datetime    string  `json:"datetime"`
	DatetimeRAW float64 `json:"datetime_raw"`
	Parent      string  `json:"parent"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Message     string  `json:"message"`
}

// Annotation represents a single annotation event
type Annotation struct {
	Time time.Time
	Text string
	Tags []string
}

/* ##################################### QUERY MODEL #################################### */

// Datasource defines basic parameters for the datasource.
type Datasource struct {
	baseURL string
	api     *Api
}

// Group, Device and Sensor serve as simple structures for filtering.
type Group struct {
	Group string `json:"group"`
}

type Device struct {
	Device string `json:"device"`
}

type Sensor struct {
	Sensor string `json:"sensor"`
}

// queryModel defines the data model for queries.
type queryModel struct {
	QueryType         string   `json:"queryType"`
	ObjectId          string   `json:"objid"`
	Group             string   `json:"group"`
	Device            string   `json:"device"`
	Sensor            string   `json:"sensor"`
	Channel           string   `json:"channel"`
	Property          string   `json:"property"`
	FilterProperty    string   `json:"filterProperty"`
	IncludeGroupName  bool     `json:"includeGroupName"`
	IncludeDeviceName bool     `json:"includeDeviceName"`
	IncludeSensorName bool     `json:"includeSensorName"`
	Groups            []string `json:"groups,omitempty"`
	Devices           []string `json:"devices,omitempty"`
	Sensors           []string `json:"sensors,omitempty"`
	From              int64    `json:"from"`
	To                int64    `json:"to"`
}

// MyDatasource can be used for further internal purposes.
type MyDatasource struct{}

// 14.02.2025 13:49:00
// Api holds API-related configurations.
type Api struct {
	baseURL string
	apiKey  string
	timeout time.Duration
}