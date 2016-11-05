package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/markdaws/gohome"
)

// RegisterDiscoveryHandlers registers all of the discovery specific API REST routes
func RegisterDiscoveryHandlers(r *mux.Router, s *apiServer) {

	// Get a list of all the devices that we can discover
	r.HandleFunc("/api/v1/discovery/discoverers",
		apiListDiscoveryHandler(s.system)).Methods("GET")

	// Scan the network for all devices corresponding to the discovery ID
	r.HandleFunc("/api/v1/discovery/discoverers/{discovererID}",
		apiDiscoveryHandler(s.system)).Methods("GET")
}

func apiListDiscoveryHandler(system *gohome.System) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		infos := system.Extensions.ListDiscoverers(system)

		jsonInfos := make([]jsonDiscovererInfo, len(infos), len(infos))
		for i, info := range infos {
			jsonInfos[i] = jsonDiscovererInfo{
				ID:          info.ID,
				Name:        info.Name,
				Description: info.Description,
				Type:        info.Type,
			}
		}
		if err := json.NewEncoder(w).Encode(jsonInfos); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func apiDiscoveryHandler(system *gohome.System) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		discovererID := vars["discovererID"]

		discoverer := system.Extensions.FindDiscovererFromID(system, discovererID)
		if discoverer == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		res, err := discoverer.ScanDevices(system)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		jsonDevices := make(devices, len(res.Devices))
		for i, device := range res.Devices {

			//TODO: This API shouldn't be sending back client ids, the client should
			//make these values up
			jsonZones := make(zones, len(device.Zones))
			var j int
			for _, zn := range device.Zones {
				jsonZones[j] = jsonZone{
					Address:     zn.Address,
					ID:          zn.ID,
					Name:        zn.Name,
					Description: zn.Description,
					DeviceID:    device.ID,
					Type:        zn.Type.ToString(),
					Output:      zn.Output.ToString(),
					ClientID:    strconv.Itoa(i) + "_zone_" + strconv.Itoa(j),
				}
				j++
			}

			jsonSensors := make(sensors, len(device.Sensors))
			j = 0
			for _, sen := range device.Sensors {
				jsonSensors[j] = jsonSensor{
					ID:          sen.ID,
					Name:        sen.Name,
					Description: sen.Description,
					Address:     sen.Address,

					//TODO: Shouldn't be setting ClientID here
					ClientID: strconv.Itoa(i) + "_sensor_" + strconv.Itoa(j),

					Attr: jsonSensorAttr{
						Name:     sen.Attr.Name,
						DataType: string(sen.Attr.DataType),
						States:   sen.Attr.States,
					},
				}
				j++
			}

			var connPoolJSON *jsonConnPool
			if device.Connections != nil {
				connCfg := device.Connections.Config
				connPoolJSON = &jsonConnPool{
					Name:     connCfg.Name,
					PoolSize: int32(connCfg.Size),
				}
			}

			var authJSON *jsonAuth
			if device.Auth != nil {
				authJSON = &jsonAuth{
					Login:    device.Auth.Login,
					Password: device.Auth.Password,
					Token:    device.Auth.Token,
				}
			}
			jsonDevices[i] = jsonDevice{
				ID:              device.ID,
				Address:         device.Address,
				Name:            device.Name,
				Description:     device.Description,
				ModelNumber:     device.ModelNumber,
				ModelName:       device.ModelName,
				SoftwareVersion: device.SoftwareVersion,
				ClientID:        "device_" + strconv.Itoa(i),
				Zones:           jsonZones,
				Sensors:         jsonSensors,
				ConnPool:        connPoolJSON,
				Auth:            authJSON,
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(jsonDevices); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
