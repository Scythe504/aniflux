package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/scythe504/aniflux/internal/anilist"
	"github.com/scythe504/aniflux/internal/utils"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(s.corsMiddleware)

	r.HandleFunc("/", s.HelloWorldHandler)

	r.HandleFunc("/trending", s.trendingAnime)

	r.HandleFunc("/seasonal", s.getSeasonal)

	r.HandleFunc("/search", s.searchAnime)

	r.HandleFunc("/genre", s.getGenre)

	r.HandleFunc("/{id}", s.getAnime)

	r.HandleFunc("/{id}/episodes", s.getEpisodes)

	r.HandleFunc("/{id}/episodes/{epNumber}/sources", s.getSources)

	r.HandleFunc("/{id}/recommendations", s.getRecommendations)

	return r
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dynamic CORS Origin handling
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Requested-With, Origin")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) trendingAnime(w http.ResponseWriter, r *http.Request) {
	page, perPage, err := getPageParams(r, 1, 5)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	trendingMedia, err := s.rs.Trending(r.Context(), page, perPage)
	if err != nil {
		utils.LogHandlerError(r, "TrendingAnime", err, map[string]any{
			"page":    page,
			"perPage": perPage,
		})
		utils.WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, trendingMedia)
}

func (s *Server) getSeasonal(w http.ResponseWriter, r *http.Request) {
	page, perPage, err := getPageParams(r, 1, 24)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	rawSeason := r.URL.Query().Get("season")
	var season *anilist.SEASON
	if rawSeason != "" {
		s := anilist.SEASON(strings.ToUpper(rawSeason))
		season = &s
	}

	rawYear := r.URL.Query().Get("year")
	var year *int
	if rawYear != "" {
		y, err := strconv.Atoi(rawYear)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid year query, value must be a number")
			return
		}
		year = &y
	}

	media, err := s.rs.GetMediaBySeason(r.Context(), season, year, page, perPage)
	if err != nil {
		utils.LogHandlerError(r, "GetSeasonal", err, map[string]any{
			"page":    page,
			"perPage": perPage,
			"season":  rawSeason,
			"year":    rawYear,
		})
		utils.WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, media)
}

func (s *Server) searchAnime(w http.ResponseWriter, r *http.Request) {
	page, perPage, err := getPageParams(r, 1, 24)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		utils.WriteError(w, http.StatusBadRequest, "Missing q query")
		return
	}

	media, err := s.rs.Search(r.Context(), query, page, perPage)
	if err != nil {
		utils.LogHandlerError(r, "SearchAnime", err, map[string]any{
			"page":    page,
			"perPage": perPage,
			"query":   query,
		})
		utils.WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, media)
}

func (s *Server) getAnime(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid anilistId, must be a number")
		return
	}
	media, err := s.rs.GetMedia(r.Context(), int(id))
	if err != nil {
		utils.LogHandlerError(r, "GetAnime", err, map[string]any{
			"id": id,
		})
		utils.WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, media)
}

func (s *Server) getEpisodes(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid anilist id")
		return
	}

	page, perPage, err := getPageParams(r, 1, 24)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	episodes, err := s.rs.GetEpisodes(r.Context(), int(id), int(page), int(perPage))
	if err != nil {
		utils.LogHandlerError(r, "GetEpisodes", err, map[string]any{
			"id":      id,
			"page":    page,
			"perPage": perPage,
			"errLine": 203,
		})
		utils.WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, episodes)
}

func (s *Server) getSources(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid anilist id")
		return
	}

	epNumber := mux.Vars(r)["epNumber"]
	if epNumber == "" {
		utils.WriteError(w, http.StatusBadRequest, "Invalid episode number")
		return
	}

	sources, err := s.rs.GetSources(r.Context(), id, epNumber)
	if err != nil {
		utils.LogHandlerError(r, "GetSources", err, map[string]any{
			"epNumber": epNumber,
			"id":       id,
		})
		utils.WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, sources)
}

func (s *Server) getRecommendations(w http.ResponseWriter, r *http.Request) {
	anilistId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid anilist id")
		return
	}

	page, perPage, err := getPageParams(r, 1, 5)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	media, err := s.rs.GetRecommendations(r.Context(), anilistId, page, perPage)
	if err != nil {
		utils.LogHandlerError(r, "GetRecommendations", err, map[string]any{
			"id":      anilistId,
			"page":    page,
			"perPage": perPage,
		})
		utils.WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, media)
}

func (s *Server) getGenre(w http.ResponseWriter, r *http.Request) {
	page, perPage, err := getPageParams(r, 1, 24)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	rawGenre := r.URL.Query().Get("genre")
	if rawGenre == "" {
		utils.WriteError(w, http.StatusBadRequest, "Missing genre query")
		return
	}

	media, err := s.rs.GetMediaByGenre(r.Context(), strings.Split(rawGenre, ","), page, perPage)
	if err != nil {
		utils.LogHandlerError(r, "GetGenre", err, map[string]any{
			"genre":   rawGenre,
			"page":    page,
			"perPage": perPage,
		})
		utils.WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, media)
}

func getPageParams(r *http.Request, defaultPage, defaultPerPage int) (int, int, error) {
	page, err := getPositiveIntQuery(r, "page", defaultPage)
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid page query, value must be a positive number")
	}

	perPage, err := getPositiveIntQuery(r, "perPage", defaultPerPage)
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid perPage query, value must be a positive number")
	}

	return page, perPage, nil
}

func getPositiveIntQuery(r *http.Request, key string, defaultValue int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value < 1 {
		return 0, fmt.Errorf("%s must be positive", key)
	}

	return value, nil
}
