package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ugifractal/pidash/player"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	musicPath := os.Getenv("MUSIC_PATH")
	fmt.Println(musicPath)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	godotenv.Load(".env")
	/*
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello"))
		})
	*/
	r.Get("/player/brightness", player.HandleBrightness)
	r.Get("/player/pause", player.HandlePause)
	r.Get("/player/volup", func(w http.ResponseWriter, r *http.Request) {

		player.HandleCommand("volup 1")
	})
	r.Get("/player/voldown", func(w http.ResponseWriter, r *http.Request) {
		player.HandleCommand("voldown 1")
	})
	r.Get("/player/next", func(w http.ResponseWriter, r *http.Request) {
		ret := player.HandleCommand("next")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		response, _ := json.Marshal(map[string]string{"message": ret, "status": "ok"})
		w.Write(response)
	})

	r.Get("/player/status", func(w http.ResponseWriter, r *http.Request) {
		ret := player.HandleCommand("status")
		split := strings.Split(ret, "\r\n")
		filename := ""
		for _, str := range split {
			if strings.HasPrefix(str, "status change: ( new input:") {
				path := strings.Split(str, " ")[5]
				filename = filepath.Base(path)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		response, _ := json.Marshal(map[string]string{"message": ret, "filename": filename, "status": "ok"})
		w.Write(response)
	})

	r.Get("/player/prev", func(w http.ResponseWriter, r *http.Request) {
		ret := player.HandleCommand("prev")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		response, _ := json.Marshal(map[string]string{"message": ret, "status": "ok"})
		w.Write(response)
	})
	r.Post("/player/add_song", func(w http.ResponseWriter, r *http.Request) {
		var song player.Song
		json.NewDecoder(r.Body).Decode(&song)
		fmt.Println(song)
		fmt.Println(song.Filename)
		player.HandleCommand("add " + song.Filename)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		response, _ := json.Marshal(map[string]string{"message": "good", "status": "ok"})
		w.Write(response)
	})
	r.Get("/player/load_music_folder", player.HandleLoadMusicFolder)
	r.Get("/player/playlist", func(w http.ResponseWriter, r *http.Request) {
		ret := player.HandleCommand("playlist")
		split := strings.Split(ret, "\r\n")
		var songs []player.Song
		for _, str := range split {
			if strings.HasPrefix(str, "|  -") {
				var s player.Song
				s.Filename = strings.Replace(str, "|  - ", "", 1)
				songs = append(songs, s)
			}
		}
		var reply player.PlayListReply
		reply.Status = "ok"
		reply.Songs = songs
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		//response, _ := json.Marshal(map[string]string{"output": ret, "status": "ok"})
		response, _ := json.MarshalIndent(reply, "", " ")
		w.Write(response)
	})

	r.Get("/player/clear", func(w http.ResponseWriter, r *http.Request) {
		ret := player.HandleCommand("clear")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		response, _ := json.Marshal(map[string]string{"output": ret, "status": "ok"})
		w.Write(response)
	})

	r.Get("/player/goto", func(w http.ResponseWriter, r *http.Request) {
		keys, _ := r.URL.Query()["key"]
		key := keys[0]
		ret := player.HandleCommand("goto " + key)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		response, _ := json.Marshal(map[string]string{"output": ret, "status": "ok"})
		w.Write(response)
	})
	r.Get("/player/open_file", player.HandleOpenFile)

	fmt.Println("started")
	staticPath := os.Getenv("FRONTEND_PATH")
	FileServer(r, "/", staticPath)

	http.ListenAndServe(":3333", r)
}

// FileServer is serving static files
func FileServer(r chi.Router, public string, static string) {

	if strings.ContainsAny(public, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	root, _ := filepath.Abs(static)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		panic("Static Documents Directory Not Found")
	}

	fs := http.StripPrefix(public, http.FileServer(http.Dir(root)))

	if public != "/" && public[len(public)-1] != '/' {
		r.Get(public, http.RedirectHandler(public+"/", 301).ServeHTTP)
		public += "/"
	}

	r.Get(public+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := strings.Replace(r.RequestURI, public, "/", 1)
		if _, err := os.Stat(root + file); os.IsNotExist(err) {
			http.ServeFile(w, r, path.Join(root, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	}))
}
