package relawanapps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

// Relawan struktur
// struktur ini dipakai sebagai entity oleh DataStore
// dan setiap fieldnya merupakan properti entity
type Relawan struct {
	Id          int64     `json:"id"`
	JumlahSuara int       `json:"jumlah_suara"`
	Pihak       int       `json:"pihak"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// Kandidat struktur
// hanya di pakai untuk response, tidak masuk entity
type Kandidat struct {
	Nama       string `json:"nama"`
	NomorUrut  int    `json:"nomor_urut"`
	TotalSuara int    `json:"total_suara"`
}

// defaultSuaraRelawan digunakan untuk mendefine parent key semua
// entity SuaraRelawan
func defaultSuaraRelawan(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "SuaraRelawan", "default", 0, nil)
}

// generate key untuk setiap entity Relawan
func (s *Relawan) key(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "SuaraRelawan", "", 0, defaultSuaraRelawan(c))
}

// menyimpan data dan merespon data yg sudah kesimpan
func (s *Relawan) save(c appengine.Context) (*Relawan, error) {
	s.SubmittedAt = time.Now()
	_, err := datastore.Put(c, s.key(c), s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// mengubah JSON yang masuk menjadi data struktur Relawan
func decodeSuaraJSON(r io.ReadCloser) (*Relawan, error) {
	defer r.Close()
	var suara Relawan
	err := json.NewDecoder(r).Decode(&suara)
	return &suara, err
}

// Get semua suara relawan
func getAllSuaraRelawan(c appengine.Context) ([]Relawan, error) {
	suara := []Relawan{}
	_, err := datastore.NewQuery("SuaraRelawan").Ancestor(defaultSuaraRelawan(c)).Order("SubmittedAt").GetAll(c, &suara)
	if err != nil {
		return nil, err
	}
	return suara, nil
}

func init() {
	http.HandleFunc("/", redirectToRepos)
	http.HandleFunc("/suara", handler)
	http.HandleFunc("/suara/prabowo", Prabowo)
	http.HandleFunc("/suara/jokowi", Jokowi)
}

// Setiap request '/' akan di redirect ke https://github.com/pyk/relawanapps
func redirectToRepos(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://github.com/pyk/relawanapps", http.StatusFound)
	return
}

// url '/suara' di handle oleh fungsi ini
func handler(w http.ResponseWriter, r *http.Request) {
	// buat konteks baru
	c := appengine.NewContext(r)
	// handle request lewat handleSuara()
	val, err := handleSuara(c, r)
	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}
	if err != nil {
		c.Errorf("Relawan error: %#v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleSuara menghandle request sesuai method yang di pakai client
// merespon dalam bentuk struktur
func handleSuara(c appengine.Context, r *http.Request) (interface{}, error) {
	switch r.Method {
	// jika method yg di pakai adalah 'POST' maka akan merespon JSON data
	// yang sukses di simpan
	case "POST":
		suara, err := decodeSuaraJSON(r.Body)
		if err != nil {
			return nil, err
		}
		// call method .save()
		return suara.save(c)
	case "GET":
		return getAllSuaraRelawan(c)
	}
	return nil, fmt.Errorf("method not implemented")
}

// menghandle request di '/suara/prabowo'
func Prabowo(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	val, err := totalSuara(1, c, r)
	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}
	if err != nil {
		c.Errorf("Relawan error: %#v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// menghandle request di '/suara/jokowi'
func Jokowi(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	val, err := totalSuara(2, c, r)
	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}
	if err != nil {
		c.Errorf("Relawan error: %#v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// count total suara masing-masing kandidat
func totalSuara(n int, c appengine.Context, r *http.Request) (interface{}, error) {
	suara := []Relawan{}
	var kandidat Kandidat
	if n == 1 {
		kandidat.Nama = "Prabowo"
		kandidat.NomorUrut = 1
	}
	if n == 2 {
		kandidat.Nama = "Jokowi"
		kandidat.NomorUrut = 2
	}
	_, err := datastore.NewQuery("SuaraRelawan").Ancestor(defaultSuaraRelawan(c)).Filter("Pihak =", n).GetAll(c, &suara)
	if err != nil {
		return nil, err
	}
	if err == nil {
		for i := 0; i < len(suara); i++ {
			kandidat.TotalSuara += suara[i].JumlahSuara
		}
		return kandidat, nil
	}

	return nil, fmt.Errorf("nomor urut kandidat")

}
