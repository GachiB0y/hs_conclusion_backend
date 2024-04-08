package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"hs-conclusion/internal/storage/mysql"
	"hs-conclusion/model"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// pallets slice to seed record album data.
// var pallets = []pallet{
// 	{ID: "1", Title: "Blue Train", Items: []box{{BoxID: "1"}, {BoxID: "2"}}, Price: 56.99},
// 	{ID: "2", Title: "Jeru", Items: []box{{BoxID: "1"}, {BoxID: "2"}}, Price: 17.99},
// 	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Items: []box{{BoxID: "1"}, {BoxID: "2"}}, Price: 39.99},
// }

func getPallets(barcode string) []model.Pallet {
	var pallets = []model.Pallet{{
		Barcode:        barcode,
		CountItemInBox: 4,
		Items: []model.Box{
			{
				Barcode: "043456789123456783",
				Items: []model.Item{
					{Barcode: "item1"},
					{Barcode: "item2"},
					{Barcode: "item3"},
					{Barcode: "item4"},
				},
			},
			{
				Barcode: "043456789123456782",
				Items: []model.Item{
					{Barcode: "item1"},
					{Barcode: "item2"},
					{Barcode: "item3"},
					{Barcode: "item4"},
				},
			},
		},
	}}

	return pallets
}

func main() {

	storage, err := mysql.New()
	if err != nil {
		fmt.Errorf("%s: %w", err)
	}
	storage.PingDB()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", getAlbums)
	r.Get("/pallet", func(w http.ResponseWriter, r *http.Request) {
		getAlbumsByID(w, r, storage)
	})
	r.Post("/pallet", postAlbums)

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Serve requests concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		http.ListenAndServe(":3000", r)
	}()

	// Wait for all goroutines to finish
	wg.Wait()
}

func getAlbums(w http.ResponseWriter, r *http.Request) {
	pallets := getPallets("143456789123456999")

	data, err := json.Marshal(pallets)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Write([]byte(data))
	fmt.Printf("%s", data)
}

func getAlbumsByID(w http.ResponseWriter, r *http.Request, storage *mysql.Storage) {
	idPllet := r.URL.Query().Get("barcode")

	// pallets := getPallets(idPllet)
	pallets, err := storage.GetPallets(idPllet)

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, pallet := range pallets {

		data, err := json.Marshal(pallet)
		if err != nil {
			fmt.Println(err)
			return
		}

		w.Write([]byte(data))
		fmt.Printf("%s", data)

	}

}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(w http.ResponseWriter, r *http.Request) {
	var newAlbum model.Pallet
	err := json.NewDecoder(r.Body).Decode(&newAlbum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var cachePallets = make([]model.Pallet, 0, 1)

	cachePallets = append(cachePallets, newAlbum)
	// Далее можно использовать newAlbum для выполнения необходимых операций, например сохранения в базу данных

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newAlbum)
}
