package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1DeliDolu/PRTG/maxmarkusprogram/prtg/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
)

/*  ################################################# NewDatasource #################################################### */
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}
	baseURL := fmt.Sprintf("https://%s", config.Path)

	cacheTime := config.CacheTime
	if cacheTime <= 0 {
		cacheTime = 30 * time.Second
	}

	return &Datasource{
		baseURL: baseURL,
		api:     NewApi(baseURL, config.Secrets.ApiKey, cacheTime, 10*time.Second),
	}, nil
}

/*  ########################################### Dispose ################################################### */
func (d *Datasource) Dispose() {

}

/*  ########################################### QueryData ################################################### */
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	// Her sorgu için query metodunu çağırıyoruz.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

/* ######################################## parsePRTGDateTime ##############################################################  */
func parsePRTGDateTime(datetime string) (time.Time, string, error) {
	// Eğer datetime bir aralık içeriyorsa
	if strings.Contains(datetime, " - ") {
		parts := strings.Split(datetime, " - ")
		if len(parts) == 2 {
			datePart := strings.Split(parts[0], " ")[0]
			timePart := strings.TrimSpace(parts[1])
			datetime = datePart + " " + timePart
		}
	}

	backend.Logger.Debug(fmt.Sprintf("Parsing PRTG datetime: %s", datetime))

	layouts := []string{
		"02.01.2006 15:04:05",
		time.RFC3339,
		"2006-01-02 15:04:05",
	}

	var parseErr error
	for _, layout := range layouts {
		parsedTime, err := time.Parse(layout, datetime)
		if err == nil {
			// Saati 1 saat geri al
			adjustedTime := parsedTime.Add(-time.Hour)
			unixTime := adjustedTime.Unix()
			return adjustedTime, strconv.FormatInt(unixTime, 10), nil
		}
		parseErr = err
	}

	backend.Logger.Error("Date parsing failed for all formats",
		"datetime", datetime,
		"error", parseErr)
	return time.Time{}, "", fmt.Errorf("failed to parse time '%s': %w", datetime, parseErr)
}

/* ######################################## CheckHealth ##############################################################  */
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}

	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	if config.Secrets.ApiKey == "" {
		res.Status = backend.HealthStatusError
		res.Message = "API key is missing"
		return res, nil
	}

	status, err := d.api.GetStatusList()
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("Failed to get PRTG status: %v", err)
		return res, nil
	}

	res.Status = backend.HealthStatusOk
	res.Message = fmt.Sprintf("Data source is working. PRTG Version: %s", status.Version)
	return res, nil
}

/* ######################################## CallResource ##############################################################  */
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	pathParts := strings.Split(req.Path, "/")
	switch pathParts[0] {
	case "groups":
		return d.handleGetGroups(sender)
	case "devices":
		if len(pathParts) < 2 {
			errorResponse := map[string]string{"error": "group parameter is required"}
			errorJSON, _ := json.Marshal(errorResponse)
			return sender.Send(&backend.CallResourceResponse{
				Status:  http.StatusBadRequest,
				Headers: map[string][]string{"Content-Type": {"application/json"}},
				Body:    errorJSON,
			})
		}
		group := pathParts[1]
		return d.handleGetDevices(sender, group)
	case "sensors":
		if len(pathParts) < 2 {
			errorResponse := map[string]string{"error": "device parameter is required"}
			errorJSON, _ := json.Marshal(errorResponse)
			return sender.Send(&backend.CallResourceResponse{
				Status:  http.StatusBadRequest,
				Headers: map[string][]string{"Content-Type": {"application/json"}},
				Body:    errorJSON,
			})
		}
		device := pathParts[1]
		return d.handleGetSensors(sender, device)

	case "channels":
		if len(pathParts) < 2 {
			errorResponse := map[string]string{"error": "missing objid parameter"}
			errorJSON, _ := json.Marshal(errorResponse)
			return sender.Send(&backend.CallResourceResponse{
				Status:  http.StatusBadRequest,
				Headers: map[string][]string{"Content-Type": {"application/json"}},
				Body:    errorJSON,
			})
		}
		return d.handleGetChannel(sender, pathParts[1])
	default:
		return sender.Send(&backend.CallResourceResponse{Status: http.StatusNotFound})
	}
}

func (d *Datasource) handleGetGroups(sender backend.CallResourceResponseSender) error {
	groups, err := d.api.GetGroups()
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(err.Error()),
		})
	}
	body, err := json.Marshal(groups)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf("error marshaling groups: %v", err)),
		})
	}
	return sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    body,
	})
}

/* ######################################### handleGetDevices ############################################################*/
func (d *Datasource) handleGetDevices(sender backend.CallResourceResponseSender, group string) error {
	if group == "" {
		errorResponse := map[string]string{"error": "missing group parameter"}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusBadRequest,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}

	devices, err := d.api.GetDevices(group)
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusInternalServerError,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}

	body, err := json.Marshal(devices)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf("error marshaling devices: %v", err)),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    body,
	})
}

/* ######################################### handleGetSensors ############################################################*/
func (d *Datasource) handleGetSensors(sender backend.CallResourceResponseSender, device string) error {
	if device == "" {
		errorResponse := map[string]string{"error": "missing device parameter"}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusBadRequest,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}

	sensors, err := d.api.GetSensors(device)
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusInternalServerError,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}

	body, err := json.Marshal(sensors)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf("error marshaling sensors: %v", err)),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    body,
	})
}

/*  ########################################  handleGetChannel ########################################  */
func (d *Datasource) handleGetChannel(sender backend.CallResourceResponseSender, objid string) error {
	if objid == "" {
		errorResponse := map[string]string{"error": "missing objid parameter"}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusBadRequest,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}
	channels, err := d.api.GetChannels(objid)
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusInternalServerError,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}
	body, err := json.Marshal(channels)
	if err != nil {
		errorResponse := map[string]string{"error": fmt.Sprintf("error marshaling channels: %v", err)}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusInternalServerError,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}
	return sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    body,
	})

}
