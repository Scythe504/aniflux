package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/scythe504/aniflux/internal/utils"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(s.corsMiddleware)

	r.HandleFunc("/", s.HelloWorldHandler)

	aniflux := r.PathPrefix("/aniflux").Subrouter()

	aniflux.HandleFunc("/trending", s.trendingAnime)

	aniflux.HandleFunc("/anime/{id}", s.getAnime)

	return r
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS Headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Wildcard allows all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Credentials not allowed with wildcard origins

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


// func (s *Server) Media(w http.ResponseWriter, r *http.Request) {
// 	anilistId, err := strconv.ParseInt(mux.Vars(r)["anilistId"], 10, 32)

// 	if err != nil {
// 		utils.WriteError(w, http.StatusBadRequest, "invalid req parameter")
// 		return
// 	}

// 	page := 1
// 	perPage := 5
// 	media, err := anilist.FetchAnilistMedia(int(anilistId), &page, &perPage, r.Context())

// 	if err != nil {
// 		log.Println(err)
// 		utils.WriteError(w, http.StatusInternalServerError, "Failed To fetch anime")
// 		return
// 	}
// 	log.Println(media.ID)

// 	anizipResp, err := anizip.FetchAnizipData(anizip.AnilistID, media.ID)
// 	if err != nil {
// 		log.Println(err)
// 		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch anizip data")
// 		return
// 	}

// 	fmt.Println(anizipResp.Episodes["1"].AniDbEid)
// 	torznabResp, err := sources.FetchSources(anizipResp.Episodes["1"].AniDbEid)
// 	if err != nil {
// 		log.Println(err)
// 		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch tosho data")
// 		return	
// 	}

// 	utils.WriteJSON(w, http.StatusOK, map[string]any{
// 		"media": media,
// 		"anizipResp": anizipResp,
// 		"torznabResp": torznabResp,
// 	})
// }

func (s *Server) trendingAnime(w http.ResponseWriter, r *http.Request){
	page := 1
	perPage := 5

	trendingMedia, err := s.rs.Trending(r.Context(), page, perPage)
	if err != nil {
		log.Println(err)
		utils.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, trendingMedia)
}

func (s *Server) getAnime(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid anilistId, must be a number")
		return
	}
	media, err := s.rs.GetMedia(r.Context(), int(id))
	if err != nil {
		log.Println("[GetAnime]:",err)
		utils.WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, media)
}