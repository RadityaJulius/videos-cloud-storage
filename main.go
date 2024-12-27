package main

import (
	"log"
	"net/http"
	"video-upload/config"
	videoscontrollers "video-upload/controllers/videoscontroller"
)

func main() {
	err := config.InitDB("root:@tcp(localhost:3306)/video_db")
	if err != nil {
		log.Fatal(err)
	}
	defer config.DB.Close()

	http.HandleFunc("/", videoscontrollers.UploadForm)
	http.HandleFunc("/upload", videoscontrollers.UploadVideo)
	http.HandleFunc("/videos", videoscontrollers.ListVideos)
	http.HandleFunc("/video/", videoscontrollers.ServeVideo)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
