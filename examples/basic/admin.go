package main

import (
	"net/http"
	"time"

	"github.com/sejta/nope/json"
	"github.com/sejta/nope/router"
)

type adminStatsResponse struct {
	UptimeSec int64  `json:"uptime_sec"`
	Version   string `json:"version"`
}

func adminRouter(startedAt time.Time) http.Handler {
	r := router.New()
	r.GET("/stats", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uptime := int64(time.Since(startedAt).Seconds())
		resp := adminStatsResponse{UptimeSec: uptime, Version: "dev"}
		json.WriteJSON(w, http.StatusOK, resp)
	}))
	return r
}
