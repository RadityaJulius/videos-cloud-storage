package videosmodel

import (
	"database/sql"
	"video-upload/entities"
)

func GetAllVideos(db *sql.DB) ([]entities.Video, error) {
	rows, err := db.Query("SELECT title, description, file_path FROM videos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []entities.Video
	for rows.Next() {
		var video entities.Video
		if err := rows.Scan(&video.Title, &video.Description, &video.FilePath); err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}
	return videos, nil
}
