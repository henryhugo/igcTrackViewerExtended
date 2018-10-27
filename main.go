package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	igc "github.com/marni/goigc"
)

type igcTrack struct {
	H_date        string  //"H_date": <date from File Header, H-record>,
	Pilot         string  //"pilot": <pilot>,
	Glider        string  //"glider": <glider>,
	Glider_id     string  //"glider_id": <glider_id>,
	Track_length  float64 //"track_length": <calculated total track length>
	Track_src_url string  //"track_src_url": <the original URL used to upload the track, ie. the URL used with POST>
}

type API struct {
	Uptime  time.Time //"uptime": <uptime>
	Info    string    //"info": "Service for IGC tracks."
	Version string    //"version": "v1"

}
type igcFile struct {
	Url string //a valid igc URL
}

type igcDB struct {
	igcs map[string]igcFile
}

func (db *igcDB) add(igc igcFile, id string) {
	for _, file := range db.igcs {
		if igc == file {
			return
		}
	}
	db.igcs[id] = igc
}

func (db igcDB) Count() int {
	return len(db.igcs)
}

func (db igcDB) Get(idWanted string) igcFile {
	for id, file := range db.igcs {
		if idWanted == id {
			return file
		}
	}
	return igcFile{}
}

func getApi(w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")
	//io.WriteString(w, "Api information :\n")
	api := &API{}
	api.Uptime = time.Now()
	api.Info = "Service for IGC tracks."
	api.Version = "version : v1"
	//fmt.Fprintf(w, "%s\n%s\n%s", api.Uptime, api.Info, api.Version)
	json.NewEncoder(w).Encode(api)

}

func igcHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		{

			if r.Body == nil {
				http.Error(w, "no JSON body", http.StatusBadRequest)
				return
			}
			var igc igcFile
			//TODO check correct igc URL
			err := json.NewDecoder(r.Body).Decode(&igc)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			fmt.Fprintf(w, "URL : %s\n", igc.Url)
			Idstr := "id"
			strValue := fmt.Sprintf("%d", idCount)
			newId := Idstr + strValue
			ids = append(ids, newId)
			idCount += 1
			db.add(igc, newId)
			json.NewEncoder(w).Encode(newId)
		}
	case "GET":
		{
			//GET case
			http.Header.Add(w.Header(), "content-type", "application/json")
			parts := strings.Split(r.URL.Path, "/")
			//fmt.Fprintf(w, "longueur : %d\n", len(parts))
			//fmt.Fprintln(w, parts)
			if len(parts) < 5 || len(parts) > 6 {
				//deal with errors
				fmt.Fprintln(w, "wrong numbers of parameters")
				return
			}
			if parts[4] == "" {
				//deal with the array
				json.NewEncoder(w).Encode(ids)

			}
			if parts[4] != "" && parts[5] == "" {
				fmt.Fprintln(w, "Information about the id")
				//deal with the id
				var igcWanted igcFile
				rgx, _ := regexp.Compile("^id[0-9]*")
				id := parts[4]
				if rgx.MatchString(id) == true {
					igcWanted = db.Get(id)

					//then encode the igcFile
					url := igcWanted.Url
					track, err := igc.ParseLocation(url)
					if err != nil {
						//fmt.Errorf("Problem reading the track", err)
					}
					igcT := igcTrack{}
					igcT.Glider = track.GliderType
					igcT.Glider_id = track.GliderID
					igcT.Pilot = track.Pilot
					igcT.Track_length = track.Task.Distance()
					igcT.H_date = track.Date.String()
					igcT.Track_src_url = igcWanted.Url
					json.NewEncoder(w).Encode(igcT)
				}
				if rgx.MatchString(id) == false {
					fmt.Fprintln(w, "Use format id0 or id21 for exemple")
				}
			}
			if parts[5] != "" && parts[4] != "" {
				//TODO parse field track_lenghtto float64, return the value asked
				//fmt.Fprintln(w, parts)
				var igcW igcFile
				infoWanted := parts[5]
				id := parts[4]
				igcW = db.Get(id)
				url := igcW.Url
				track, err := igc.ParseLocation(url)
				if err != nil {
					//fmt.Errorf("Problem reading the track", err)
				}
				igcT := igcTrack{}
				switch infoWanted {
				case "pilot":
					{
						igcT.Pilot = track.Pilot
						json.NewEncoder(w).Encode(igcT.Pilot)
					}
				case "glider":
					{
						igcT.Glider = track.GliderType
						json.NewEncoder(w).Encode(igcT.Glider)
					}
				case "glider_id":
					{
						igcT.Glider_id = track.GliderID
						json.NewEncoder(w).Encode(igcT.Glider_id)
					}
				case "H_date":
					{
						igcT.H_date = track.Date.String()
						json.NewEncoder(w).Encode(igcT.H_date)
					}
				case "track_src_url":
					{
						igcT.Track_src_url = igcW.Url
						json.NewEncoder(w).Encode(igcW.Url)
					}
				}
			}

		}
	default:

		http.Error(w, "not implemented yet", http.StatusNotImplemented)

	}
}

var path, _ = regexp.Compile("/paragliding[/]{1}$")

func router(w http.ResponseWriter, r *http.Request) {
	switch {
	case path.MatchString(r.URL.Path):
		{
			getApi(w, r)
		}
	default:
		http.NotFound(w, r)
	}
}

var db igcDB
var ids []string
var idCount int

func main() {
	db = igcDB{}
	db.igcs = map[string]igcFile{}
	idCount = 0
	ids = nil
	port := os.Getenv("PORT")
	http.HandleFunc("/", router)
	http.HandleFunc("/paragliding/api", getApi)
	http.HandleFunc("/paragliding/api/track/", igcHandler)
	http.ListenAndServe(":"+port, nil)
}
