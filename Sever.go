package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Comment struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Comment  string `json:"comment"`
}

var (
	comments []Comment
	mu       sync.Mutex
)

func main() {
	loadCommentsFromFile()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/comments", commentsHandler)
	http.HandleFunc("/api/add-comment", addCommentHandler)
	http.HandleFunc("/api/delete-comment/", deleteCommentHandler)

	fs := http.FileServer(http.Dir("."))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// 首页处理函数，返回静态文件 test.html
func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "test.html")
}

// 获取所有评论的处理函数
func commentsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// 添加评论的处理函数
func addCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var comment Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	comment.ID = len(comments) + 1
	comments = append(comments, comment)
	saveCommentsToFile()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// 删除评论的处理函数
func deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	commentID := r.URL.Path[len("/api/delete-comment/"):]
	id, err := strconv.Atoi(commentID)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i, comment := range comments {
		if comment.ID == id {
			comments = append(comments[:i], comments[i+1:]...)
			break
		}
	}
	saveCommentsToFile()

	w.WriteHeader(http.StatusNoContent)
}

// 从文件加载评论数据
func loadCommentsFromFile() {
	file, err := os.Open("comments.json")
	if err != nil {
		log.Println("Failed to open comments file:", err)
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&comments)
	if err != nil {
		log.Println("Failed to decode comments file:", err)
	}
}

// 将评论数据保存到文件
func saveCommentsToFile() {
	file, err := os.Create("comments.json")
	if err != nil {
		log.Println("Failed to create comments file:", err)
		return
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(comments)
	if err != nil {
		log.Println("Failed to encode comments to file:", err)
	}
}
