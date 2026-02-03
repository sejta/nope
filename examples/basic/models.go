package main

// CreatePostRequest описывает входной запрос на создание поста.
type CreatePostRequest struct {
	Title string `json:"title"`
}

// PostResponse описывает ответ с данными поста.
type PostResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
