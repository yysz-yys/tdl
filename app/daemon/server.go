package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/iyear/tdl/pkg/ps"
	"github.com/iyear/tdl/pkg/utils"
)

type Server struct {
	httpServer *http.Server
	router     *mux.Router
	hub        *Hub
	taskMgr    *TaskManager
	logger     *zap.Logger
}

func NewServer(port int, logger *zap.Logger) *Server {
	router := mux.NewRouter()
	hub := NewHub(logger)
	taskMgr := NewTaskManager(hub)

	srv := &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		},
		router:  router,
		hub:     hub,
		taskMgr: taskMgr,
		logger:  logger,
	}

	srv.setupRoutes()
	return srv
}

func (s *Server) setupRoutes() {
	// CORS Middleware for development
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	api := s.router.PathPrefix("/api/v1").Subrouter()
	
	// WebSocket endpoint
	api.HandleFunc("/ws", s.hub.ServeWS)

	// Task management
	api.HandleFunc("/tasks", s.handleGetTasks).Methods("GET")
	api.HandleFunc("/tasks", s.handleAddDownloadTask).Methods("POST")
	api.HandleFunc("/tasks/{id}/pause", s.handlePauseTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}/resume", s.handleResumeTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", s.handleDeleteTask).Methods("DELETE")

	// System Status
	api.HandleFunc("/system/stats", s.handleSystemStats).Methods("GET")
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting daemon server", zap.String("addr", s.httpServer.Addr))
	
	go s.hub.Run(ctx)
	go s.taskMgr.Run(ctx)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	s.logger.Info("Shutting down daemon server...")
	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	return s.httpServer.Shutdown(shutdownCtx)
}

// API Responses
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.taskMgr.GetAllTasks()
	sendJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "success", Data: tasks})
}

func (s *Server) handleAddDownloadTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Code: 400, Message: "invalid request body"})
		return
	}

	task := &Task{
		Type: "download",
		Name: "New Task " + req.URL, // Simple mock name
		Size: 1024 * 1024 * 100,     // Mock size: 100MB
	}
	s.taskMgr.AddTask(task)
	
	sendJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "task added", Data: task})
}

func (s *Server) handlePauseTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	s.taskMgr.UpdateTaskStatus(id, StatusPaused, nil)
	sendJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "task paused"})
}

func (s *Server) handleResumeTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	s.taskMgr.UpdateTaskStatus(id, StatusRunning, nil)
	sendJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "task resumed"})
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	s.taskMgr.DeleteTask(id)
	sendJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "task deleted"})
}

func (s *Server) handleSystemStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cpu, _ := ps.GetSelfCPU(ctx)
	mem, _ := ps.GetSelfMem(ctx)
	goroutines := ps.GetGoroutineNum()

	stats := map[string]interface{}{
		"cpu":        fmt.Sprintf("%.2f%%", cpu),
		"goroutines": goroutines,
	}
	if mem != nil {
		stats["memory"] = utils.Byte.FormatBinaryBytes(int64(mem.RSS))
		stats["memory_bytes"] = mem.RSS
	}

	sendJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "success", Data: stats})
}
