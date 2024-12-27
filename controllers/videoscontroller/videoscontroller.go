package videoscontrollers

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"video-upload/config"
	"video-upload/models/videosmodel"
)

func UploadForm(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("views/upload.html")
	t.Execute(w, nil)
}

func UploadVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		description := r.FormValue("description")

		// Handle file upload
		file, fileHeader, err := r.FormFile("video")
		if err != nil {
			http.Error(w, "Unable to get file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Save the file to the server
		filePath := filepath.Join("uploads", fileHeader.Filename)
		os.MkdirAll("uploads", os.ModePerm)
		out, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		// Copy the uploaded file to the destination
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}

		// Insert video metadata into the database
		_, err = config.DB.Exec("INSERT INTO videos (title, description, file_path) VALUES (?, ?, ?)", title, description, filePath)
		if err != nil {
			http.Error(w, "Unable to save video metadata", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Video uploaded successfully: %s", fileHeader.Filename)
	}
}

func ListVideos(w http.ResponseWriter, r *http.Request) {
	videos, err := videosmodel.GetAllVideos(config.DB)
	if err != nil {
		http.Error(w, "Unable to retrieve videos", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("views/videos.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	t.Execute(w, videos)
}

func ServeVideo(w http.ResponseWriter, r *http.Request) {
	videoFilePath := filepath.Join("uploads", filepath.Base(r.URL.Path[len("/video/"):]))
	if _, err := os.Stat(videoFilePath); os.IsNotExist(err) {
		http.Error(w, "Video file does not exist.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	http.ServeFile(w, r, videoFilePath)
}
