package www

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/markdaws/gohome"
)

type Server interface {
	ListenAndServe(port string) error
}

type wwwServer struct {
	rootPath string
	system   *gohome.System
}

func NewServer(rootPath string, system *gohome.System) Server {
	return &wwwServer{rootPath: rootPath, system: system}
}

func (s *wwwServer) ListenAndServe(port string) error {
	r := mux.NewRouter()

	mime.AddExtensionType(".jsx", "text/jsx")
	cssHandler := http.FileServer(http.Dir(s.rootPath + "/assets/css/"))
	jsHandler := http.FileServer(http.Dir(s.rootPath + "/assets/js/"))
	jsxHandler := http.FileServer(http.Dir(s.rootPath + "/assets/jsx/"))
	imageHandler := http.FileServer(http.Dir(s.rootPath + "/assets/images/"))

	//TODO: Move api into separate http server
	r.HandleFunc("/api/v1/systems/{systemId}/scenes", apiScenesHandler(s.system)).Methods("GET")
	r.HandleFunc("/api/v1/systems/{systemId}/zones", apiZonesHandler(s.system)).Methods("GET")

	//TODO: GET vs. POST
	r.HandleFunc("/api/v1/systems/{systemId}/zones/{id}", apiZoneHandler(s.system))

	//TODO: Make for POST only
	//TODO: Have GET version to see the currently active scenes
	r.HandleFunc("/api/v1/systems/{systemId}/scenes/active", apiActiveScenesHandler(s.system)).Methods("POST")

	sub := r.PathPrefix("/assets").Subrouter()
	//sub.Methods("GET")
	sub.Handle("/css/{filename}", http.StripPrefix("/assets/css/", cssHandler))
	sub.Handle("/js/{filename}", http.StripPrefix("/assets/js/", jsHandler))
	sub.Handle("/jsx/{filename}", http.StripPrefix("/assets/jsx/", jsxHandler))
	sub.Handle("/images/{filename}", http.StripPrefix("/assets/images/", imageHandler))
	r.HandleFunc("/", rootHandler(s.rootPath))
	return http.ListenAndServe(port, r)
}

func rootHandler(rootPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, rootPath+"/assets/html/index.html")
	}
}

func apiScenesHandler(system *gohome.System) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")

		type jsonScene struct {
			Id          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		//TODO: Return in a consistent order
		scenes := make([]jsonScene, len(system.Scenes), len(system.Scenes))
		var i int32 = 0
		for _, scene := range system.Scenes {
			scenes[i] = jsonScene{Id: scene.Id, Name: scene.Name, Description: scene.Description}
			i++
		}
		if err := json.NewEncoder(w).Encode(scenes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		//TODO: Need ok?
	}
}

func apiZonesHandler(system *gohome.System) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		type jsonZone struct {
			Id          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		//TODO: Returns in a consistent order
		zones := make([]jsonZone, len(system.Zones), len(system.Zones))
		var i int32 = 0
		for _, zone := range system.Zones {
			zones[i] = jsonZone{Id: zone.Id, Name: zone.Name, Description: zone.Description}
			i++
		}

		if err := json.NewEncoder(w).Encode(zones); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func apiZoneHandler(system *gohome.System) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
		if err != nil {
			fmt.Println("a")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var x struct {
			Value float32 `json:"value"`
		}
		if err = json.Unmarshal(body, &x); err != nil {
			fmt.Println("b")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		zone, ok := system.Zones[vars["id"]]
		if !ok {
			fmt.Println("c")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		zone.SetCommand.Execute(x.Value)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct{}{})
	}
}

func apiActiveScenesHandler(system *gohome.System) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var x struct {
			Id string `json:"id"`
		}
		if err = json.Unmarshal(body, &x); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		scene, ok := system.Scenes[x.Id]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		scene.Execute()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct{}{})
	}
}
