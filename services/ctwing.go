package services

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
	httpclient "plugin_chirpstack/http_client"
	"plugin_chirpstack/model"
	"plugin_chirpstack/mqtt"
)

type CtwingService struct {
	mux *http.ServeMux
}

func NewCtwing() *CtwingService {
	return &CtwingService{
		mux: http.NewServeMux(),
	}
}

func (ctw *CtwingService) Init() *http.ServeMux {
	ctw.mux.HandleFunc("/accept/telemetry", ctw.telemetry)
	ctw.mux.HandleFunc("/accept/command-response", ctw.commandResponse)
	ctw.mux.HandleFunc("/accept/event", ctw.event)
	ctw.mux.HandleFunc("/accept/online", ctw.online)
	return ctw.mux
}

func (ctw *CtwingService) telemetry(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logrus.Debug("Raw Body:", string(body))

	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	var msg CtwingTelemetry
	err = decoder.Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logrus.Debug("telemetry:", msg)
	deviceNumber := fmt.Sprintf(viper.GetString("ctwing.device_number_key"), msg.ProductId, msg.DeviceId)
	// 读取设备信息
	deviceInfo, err := httpclient.GetDeviceConfig(deviceNumber)
	if err != nil {
		// 获取设备信息失败，请检查连接包是否正确
		logrus.Error(err)
		return
	}
	err = mqtt.PublishTelemetry(deviceInfo.Data.ID, msg.Payload)
	if err != nil {
		logrus.Error(err)
	}
}

func (ctw *CtwingService) commandResponse(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logrus.Debug("Raw Body:", string(body))
}
func (ctw *CtwingService) event(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	var msg CtwingEvent
	err := decoder.Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logrus.Debug("event:", msg)
	deviceNumber := fmt.Sprintf(viper.GetString("ctwing.device_number_key"), msg.ProductId, msg.DeviceId)
	// 读取设备信息
	deviceInfo, err := httpclient.GetDeviceConfig(deviceNumber)
	if err != nil {
		// 获取设备信息失败，请检查连接包是否正确
		logrus.Error(err)
		return
	}
	data := model.EventInfo{
		Method: msg.ServiceId,
		Params: msg.EventContent,
	}
	err = mqtt.PublishEvent(deviceInfo.Data.ID, data)
	if err != nil {
		logrus.Error(err)
	}
}
func (ctw *CtwingService) online(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	var msg CtwingOnline
	err := decoder.Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logrus.Debug("telemetry:", msg)
	deviceNumber := fmt.Sprintf(viper.GetString("ctwing.device_number_key"), msg.ProductId, msg.DeviceId)
	// 读取设备信息
	deviceInfo, err := httpclient.GetDeviceConfig(deviceNumber)
	if err != nil {
		// 获取设备信息失败，请检查连接包是否正确
		logrus.Error(err)
		return
	}

	err = mqtt.DeviceStatusUpdate(deviceInfo.Data.ID, msg.EventType)
	if err != nil {
		logrus.Error(err)
	}
}

type CtwingMessage struct {
	MessageType string `json:"messageType"`
	DeviceId    string `json:"deviceId"`
	ProductId   string `json:"productId"`
}

type CtwingOnline struct {
	CtwingMessage
	EventType int `json:"eventType"` //1上线 0下线
}

type CtwingTelemetry struct {
	CtwingMessage
	Payload map[string]interface{} `json:"payload"`
}

type CtwingEvent struct {
	CtwingMessage
	EventContent map[string]interface{} `json:"eventContent"`
	EventType    int                    `json:"eventType"`
	ServiceId    string                 `json:"serviceId"`
}
