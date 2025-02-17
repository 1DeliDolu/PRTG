package plugin

import (
	"encoding/json"
)

// PrtgTableListResponse repräsentiert die Antwort der PRTG Table List API.
// Hinweis: Prüfe, ob "prtg-version" als Array oder String geliefert wird.
type PrtgTableListResponse struct {
	PrtgVersion []PrtgStatusListResponse   `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                      `json:"treesize" xml:"treesize"`
	Groups      []PrtgGroupListResponse    `json:"groups,omitempty" xml:"groups,omitempty"`
	Devices     []PrtgDevicesListResponse  `json:"devices,omitempty" xml:"devices,omitempty"`
	Sensors     []PrtgSensorsListResponse  `json:"sensors,omitempty" xml:"sensors,omitempty"`
	Values      []PrtgChannelsListResponse `json:"channels,omitempty" xml:"channels,omitempty"`
}

//############################# GROUP LIST RESPONSE ####################################

// PrtgGroupListResponse repräsentiert die Antwort für Gruppen.
type PrtgGroupListResponse struct {
	PrtgVersion string                    `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                     `json:"treesize" xml:"treesize"`
	Groups      []PrtgGroupListItemStruct `json:"groups" xml:"groups"`
}

// PrtgGroupListItemStruct enthält Details zu einer einzelnen Gruppe.
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

// PrtgDevicesListResponse repräsentiert die Antwort für Geräte.
type PrtgDevicesListResponse struct {
	PrtgVersion string                     `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                      `json:"treesize" xml:"treesize"`
	Devices     []PrtgDeviceListItemStruct `json:"devices" xml:"devices"`
}

// PrtgDeviceListItemStruct enthält Details zu einem einzelnen Gerät.
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

// PrtgSensorsListResponse repräsentiert die Antwort für Sensoren.
type PrtgSensorsListResponse struct {
	PrtgVersion string                     `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                      `json:"treesize" xml:"treesize"`
	Sensors     []PrtgSensorListItemStruct `json:"sensors" xml:"sensors"`
}

// PrtgSensorListItemStruct enthält Details zu einem einzelnen Sensor.
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

// PrtgStatusListResponse enthält systemweite Statusinformationen.
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

// PrtgChannelsListResponse repräsentiert die Antwort für Channel-Werte.
type PrtgChannelsListResponse struct {
	PrtgVersion string                   `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64                    `json:"treesize" xml:"treesize"`
	Values      []PrtgChannelValueStruct `json:"values" xml:"values"`
}

// PrtgChannelValueStruct wird als dynamische Map zur Speicherung von Channel-Daten verwendet.
type PrtgChannelValueStruct map[string]interface{}

//############################# CHANNEL VALUE RESPONSE ####################################

// PrtgHistoricalDataResponse enthält historische Werte eines Sensors.
type PrtgHistoricalDataResponse struct {
	PrtgVersion string       `json:"prtg-version" xml:"prtg-version"`
	TreeSize    int64        `json:"treesize" xml:"treesize"`
	HistData    []PrtgValues `json:"histdata" xml:"histdata"`
}

// PrtgValues enthält den Zeitstempel und dynamische Werte.
type PrtgValues struct {
	Datetime string                 `json:"datetime"`
	Value    map[string]interface{} `json:"-"`
}

// UnmarshalJSON implementiert ein benutzerdefiniertes Unmarshal-Verfahren,
// das den "datetime"-Wert separat behandelt und den Rest in das Value-Feld packt.
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

/* ##################################### QUERY MODEL #################################### */

// Datasource definiert grundlegende Parameter für die Datasource.
type Datasource struct {
	baseURL string
	api     *Api
}

// Group, Device und Sensor dienen als einfache Strukturen zur Filterung.
type Group struct {
	Group string `json:"group"`
}

type Device struct {
	Device string `json:"device"`
}

type Sensor struct {
	Sensor string `json:"sensor"`
}

// queryModel definiert das Datenmodell für Abfragen.
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

// MyDatasource kann für weitere interne Zwecke verwendet werden.
type MyDatasource struct{}

// 14.02.2025 13:49:00
