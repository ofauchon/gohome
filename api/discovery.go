package api

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/markdaws/gohome"
	"github.com/markdaws/gohome/log"
	"github.com/markdaws/gohome/zone"
)

// RegisterDiscoveryHandlers registers all of the discovery specific API REST routes
func RegisterDiscoveryHandlers(r *mux.Router, s *apiServer) {

	// Get a list of all the devices that we can discover
	r.HandleFunc("/api/v1/discovery/discoverers",
		apiListDiscoveryHandler(s.system)).Methods("GET")

	// Scan the network for all devices corresponding to the discovery ID
	r.HandleFunc("/api/v1/discovery/discoverers/{discovererID}",
		apiDiscoveryHandler(s.system)).Methods("POST")
}

func apiListDiscoveryHandler(system *gohome.System) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		infos := system.Extensions.ListDiscoverers(system)

		jsonInfos := make([]jsonDiscovererInfo, len(infos))
		for i, info := range infos {
			jsonUIFields := make([]jsonUIField, len(info.UIFields))
			for j, field := range info.UIFields {
				jsonUIFields[j] = jsonUIField{
					ID:          field.ID,
					Label:       field.Label,
					Description: field.Description,
					Default:     field.Default,
					Required:    field.Required,
				}
			}

			jsonInfos[i] = jsonDiscovererInfo{
				ID:          info.ID,
				Name:        info.Name,
				Description: info.Description,
				PreScanInfo: info.PreScanInfo,
				UIFields:    jsonUIFields,
			}
		}
		if err := json.NewEncoder(w).Encode(jsonInfos); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func writeDiscoveryResults(sys *gohome.System, result *gohome.DiscoveryResults, w http.ResponseWriter) {
	// Need to serialize the scenes, use handy functions from scenes.go
	inputScenes := make(map[string]*gohome.Scene)
	for _, scene := range result.Scenes {
		inputScenes[scene.ID] = scene
	}

	inputDevices := make(map[string]*gohome.Device)
	dupeDevices := make(map[string]*gohome.Device)

	// Given the discovery results, we search the existing system to see if the discovery results
	// are returning any duplicate device/zone/sensor entries, if so we mark them appropriately
	// and return the existing devices to the user, with any new devices/zones/sensors appended

	for _, device := range result.Devices {
		if dupeDevice, isDupe := sys.IsDupeDevice(device); isDupe {
			dupeDevices[dupeDevice.ID] = device
		} else {
			inputDevices[device.ID] = device
		}
	}

	// JSONify all the non dupe devices
	jsonDevices := DevicesToJSON(inputDevices)

	// For all the devices we found that were dupes, we need to JSONify those separately
	// along with merging the zones + sensors of the current discovery with zones/sensors
	// already attached to the device.  For example the user may have already imported a
	// device and zone previously, then added a new zone and done a rescan, we need to
	// return the existing device and zone but also append the new zone so the user has
	// change to import the new zone
	for existingDeviceID, dupeDevice := range dupeDevices {
		existingDevice := sys.Devices[existingDeviceID]

		// JSONify the existing device, since this is a dupe we want to send back the
		// current device to the user
		jsonDupeDevice := DevicesToJSON(map[string]*gohome.Device{existingDevice.ID: existingDevice})[0]
		jsonDupeDevice.IsDupe = true

		// Have to mark zones/sensors as dupes
		for i, _ := range jsonDupeDevice.Zones {
			jsonDupeDevice.Zones[i].IsDupe = true
		}
		for i, _ := range jsonDupeDevice.Sensors {
			jsonDupeDevice.Sensors[i].IsDupe = true
		}

		// Now if we discovered any new zones/sensors we need to add those to the JSON
		// and send those back
		for _, zn := range dupeDevice.Zones {
			if _, isDupe := existingDevice.IsDupeZone(zn); !isDupe {
				jsonZone := ZonesToJSON(map[string]*zone.Zone{zn.ID: zn})[0]
				jsonZone.IsDupe = false
				jsonZone.DeviceID = existingDevice.ID
				jsonDupeDevice.Zones = append(jsonDupeDevice.Zones, jsonZone)
			}
		}

		for _, sen := range dupeDevice.Sensors {
			if _, isDupe := existingDevice.IsDupeSensor(sen); !isDupe {
				jsonSensor := SensorsToJSON(map[string]*gohome.Sensor{sen.ID: sen})[0]
				jsonSensor.IsDupe = false
				jsonSensor.DeviceID = existingDevice.ID
				jsonDupeDevice.Sensors = append(jsonDupeDevice.Sensors, jsonSensor)
			}
		}

		jsonDevices = append(jsonDevices, jsonDupeDevice)
	}

	jsonScenes := ScenesToJSON(inputScenes)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(struct {
		Devices []jsonDevice `json:"devices"`
		Scenes  []jsonScene  `json:"scenes"`
	}{
		Devices: jsonDevices,
		Scenes:  jsonScenes,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func apiDiscoveryHandler(sys *gohome.System) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		discovererID := vars["discovererID"]

		discoverer := sys.Extensions.FindDiscovererFromID(sys, discovererID)
		if discoverer == nil {
			log.V("unknown discoverer id %s", discovererID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024*1024))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var uiFields map[string]string
		if len(body) > 0 {
			if err := json.Unmarshal(body, &uiFields); err != nil {
				log.V("error unmarhsaling uiFields %s", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		res, err := discoverer.ScanDevices(sys, uiFields)
		if err != nil {
			log.V("error scanning devices %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		writeDiscoveryResults(sys, res, w)
	}
}

//TODO: Remove...
/*
func apiFromStringDiscoveryHandler(sys *gohome.System) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		discovererID := vars["discovererID"]

		discoverer := sys.Extensions.FindDiscovererFromID(sys, discovererID)
		if discoverer == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024*1024))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Println(string(body))

		unquotedBody, err := strconv.Unquote(string(body))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		res, err := discoverer.FromString(unquotedBody)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Printf("%+v\n", res)
		writeDiscoveryResults(sys, res, w)
	}
}
*/
