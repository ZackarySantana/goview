package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/zackarysantana/goview/template"
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
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	// Create a channel to send data to the template.
	// 	data := make(chan templ.Component)
	// 	// Run a background process that will take 10 seconds to complete.
	// 	go func() {
	// 		// Always remember to close the channel.
	// 		wg := sync.WaitGroup{}
	// 		wg.Add(5)
	// 		defer func() {
	// 			wg.Wait()
	// 			close(data)
	// 		}()
	// 		sleepTimeSecs := []int{4, 2, 0, 1, 1}
	// 		for i := 1; i <= 5; i++ {
	// 			go func() {
	// 				defer wg.Done()
	// 				time.Sleep(time.Duration(sleepTimeSecs[i-1]) * time.Second)

	// 				select {
	// 				case <-r.Context().Done():
	// 					// Quit early if the client is no longer connected.
	// 					return
	// 				case <-time.After(time.Second):
	// 					// Send a new piece of data to the channel.
	// 					data <- templates.Slot(i)
	// 				}
	// 			}()
	// 		}
	// 	}()

	// 	// Pass the channel to the template.
	// 	component := templates.Root(data)

	// 	// Serve using the streaming mode of the handler.
	// 	templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
	// })

	// http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	// 	// Create a channel to send deferred component renders to the template.
	// 	data := make(chan templates.SlotContents)

	// 	// We know there are 3 slots, so start a WaitGround.
	// 	var wg sync.WaitGroup
	// 	wg.Add(3)

	// 	// Start the async processes.
	// 	// Sidebar.
	// 	go func() {
	// 		defer wg.Done()
	// 		time.Sleep(time.Second * 3)
	// 		data <- templates.SlotContents{
	// 			Name:     "a",
	// 			Contents: templates.A(),
	// 		}
	// 	}()

	// 	// Content.
	// 	go func() {
	// 		defer wg.Done()
	// 		time.Sleep(time.Second * 2)
	// 		data <- templates.SlotContents{
	// 			Name:     "b",
	// 			Contents: templates.B(),
	// 		}
	// 	}()

	// 	// Footer.
	// 	go func() {
	// 		defer wg.Done()
	// 		time.Sleep(time.Second * 1)
	// 		data <- templates.SlotContents{
	// 			Name:     "c",
	// 			Contents: templates.C(),
	// 		}
	// 	}()

	// 	// Close the channel when all processes are done.
	// 	go func() {
	// 		wg.Wait()
	// 		close(data)
	// 	}()

	// 	// Pass the channel to the template.
	// 	component := templates.Page(data)

	// 	// Serve using the streaming mode of the handler.
	// 	templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
	// })

	http.Handle("/", templ.Handler(template.Index()))

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
