package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Video struct {
	Title       string
	Description *string // Use pointer to handle NULL
	FilePath    string
}

func init() {
	var err error
	// Change the DSN according to your MySQL setup
	dsn := "root:@tcp(127.0.0.1:3306)/video_db"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", uploadForm)
	http.HandleFunc("/upload", uploadVideo)
	http.HandleFunc("/videos", listVideos) // New route for listing videos
	http.HandleFunc("/video/", serveVideo) // Route to serve video files
	log.Println("Server started at port: 8080")
	http.ListenAndServe(":8080", nil)
}

func uploadForm(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("upload.html")
	t.Execute(w, nil)
}

func uploadVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		description := r.FormValue("description")

		// Handle file upload
		file, fileHeader, err := r.FormFile("video") // Capture all three return values
		if err != nil {
			http.Error(w, "Unable to get file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Save the file to the server
		filePath := filepath.Join("uploads", fileHeader.Filename) // Use fileHeader.Filename for the file name
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
		_, err = db.Exec("INSERT INTO videos (title, description, file_path) VALUES (?, ?, ?)", title, description, filePath)
		if err != nil {
			http.Error(w, "Unable to save video metadata", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Video uploaded successfully: %s", fileHeader.Filename)
	}
}

func listVideos(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT title, description, file_path FROM videos") // Removed created_at
	if err != nil {
		http.Error(w, "Unable to retrieve videos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var video Video
		if err := rows.Scan(&video.Title, &video.Description, &video.FilePath); err != nil {
			fmt.Println("Error scanning video data:", err)
			http.Error(w, "Unable to scan video data", http.StatusInternalServerError)
			return
		}
		videos = append(videos, video)
	}

	t, err := template.ParseFiles("videos.html")
	if err != nil {
		fmt.Println("Unable to load template:", err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	t.Execute(w, videos)
}

func serveVideo(w http.ResponseWriter, r *http.Request) {
	// Get the video file path from the URL
	videoFilePath := filepath.Join("uploads", filepath.Base(r.URL.Path[len("/video/"):])) // Extract filename from URL

	// Check if the video file exists
	if _, err := os.Stat(videoFilePath); os.IsNotExist(err) {
		http.Error(w, "Video file does not exist.", http.StatusNotFound)
		return
	}

	// Serve the video file
	w.Header().Set("Content-Type", "video/mp4")
	http.ServeFile(w, r, videoFilePath)
}
