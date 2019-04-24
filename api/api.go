package api

func GetResources(w http.ResponseWriter, r *http.Request) {
	var resources []Resource
	db.Find(&resources)
	json.NewEncoder(w).Encode(&resources)
}
