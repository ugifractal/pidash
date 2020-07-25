package player

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// Song struct
type Song struct {
	Filename string `json:"filename"`
}

type PlayListReply struct {
	Status string `json:"status"`
	Songs  []Song `json:"songs"`
}

func Test() {
	fmt.Println("Just Test")
}

func HandlePause(w http.ResponseWriter, r *http.Request) {
	HandleCommand("pause")
}

// Hanlde
func HandleLoadMusicFolder(w http.ResponseWriter, r *http.Request) {
	root := os.Getenv("MUSIC_PATH")
	var songs []Song
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if (filepath.Ext(path)) != ".mp3" {
			return nil
		}
		var s Song
		s.Filename = path
		songs = append(songs, s)
		return nil
	})
	var reply PlayListReply
	reply.Status = "ok"
	reply.Songs = songs

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	response, _ := json.Marshal(reply)
	w.Write(response)
}

// test
func HandleCommand(str string) string {
	sock := os.Getenv("SOCK_PATH")
	text := "echo '" + str + "' | nc -U " + sock
	out, err := exec.Command("bash", "-c", text).Output()

	if err != nil {
		//log.Fatalf("error %s\n", err)
	}
	fmt.Printf("%s\n", out)
	return string(out)
}

func HandleOpenFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("file")
	if path == "" {
		http.Error(w, "'file' params is required", 400)
		return
	}
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		http.Error(w, "File Not Found", 400)
		return
	}
	header := make([]byte, 512)
	file.Read(header)
	contentType := http.DetectContentType(header)
	stat, _ := file.Stat()
	size := strconv.FormatInt(stat.Size(), 10)

	//w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(path))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", size)
	file.Seek(0, 0)
	io.Copy(w, file)
	return
}

func HandleBrightness(w http.ResponseWriter, r *http.Request) {
	value := r.URL.Query().Get("value")
	file := "/sys/class/backlight/rpi_backlight/brightness"
	// file := "text.txt"
	f, err := os.OpenFile(file, os.O_WRONLY, 666)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprint(f, value)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	response, _ := json.Marshal(map[string]string{"brightness": value, "status": "ok"})
	w.Write(response)
}
