package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/zackarysantana/goview/templates"
)

// ProjectData holds basic info about the Go project
// type ProjectData struct {
// 	Name string
// }

// parseGoMod extracts the module name from go.mod
// func parseGoMod(path string) (ProjectData, error) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return ProjectData{}, err
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if strings.HasPrefix(line, "module ") {
// 			name := strings.TrimSpace(strings.TrimPrefix(line, "module "))
// 			return ProjectData{Name: name}, nil
// 		}
// 	}
// 	return ProjectData{}, nil
// }

func main() {
	root := templates.Root()

	http.Handle("/", templ.Handler(root))

	http.Handle("/assets/",
		http.StripPrefix("/assets",
			http.FileServer(http.Dir("assets"))))

	fmt.Println("Listening on :3000 (the proxy is on :7331)")
	http.ListenAndServe(":3000", nil)
	// data, err := parseGoMod("go.mod")
	// if err != nil {
	// 	log.Printf("Could not parse go.mod: %v", err)
	// }

	// tmpl := template.Must(template.ParseFiles("templates/index.html"))

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	tmpl.Execute(w, data)
	// })

	// log.Println("Server running at http://localhost:8080/")
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
