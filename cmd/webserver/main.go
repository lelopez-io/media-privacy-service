package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lelopez-io/media-scrubber-service/internal/mediaprocessor"
)

type Session struct {
	ID           string
	FileCounter  int
	LastAccessed time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.Mutex
}

var sessionManager *SessionManager

func main() {
	cleanWorkdir := flag.Bool("clean", false, "Clean the workdir before starting the server")
	flag.Parse()

	if *cleanWorkdir {
		err := cleanWorkDir()
		if err != nil {
			log.Fatalf("Failed to clean workdir: %v", err)
		}
		fmt.Println("Workdir cleaned successfully.")
	}

	sessionManager = &SessionManager{
		sessions: make(map[string]*Session),
	}

	go sessionManager.cleanupSessions()

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/download/", handleDownload)

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func cleanWorkDir() error {
	workdir := filepath.Join("workdir", "web")
	err := os.RemoveAll(workdir)
	if err != nil {
		return fmt.Errorf("failed to remove workdir: %v", err)
	}
	return os.MkdirAll(workdir, os.ModePerm)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
    filename := r.URL.Path[len("/download/"):]
    filePath := filepath.Join("workdir", "web", filename)

    // Check if file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    // Serve the file
    http.ServeFile(w, r, filePath)
}

func (sm *SessionManager) getSession(w http.ResponseWriter, r *http.Request) *Session {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	cookie, err := r.Cookie("session_id")
	if err != nil {
		sessionID := uuid.New().String()
		session := &Session{ID: sessionID, FileCounter: 0, LastAccessed: time.Now()}
		sm.sessions[sessionID] = session

		http.SetCookie(w, &http.Cookie{
			Name:    "session_id",
			Value:   sessionID,
			Expires: time.Now().Add(24 * time.Hour),
		})

		return session
	}

	session, found := sm.sessions[cookie.Value]
	if !found {
		session = &Session{ID: cookie.Value, FileCounter: 0, LastAccessed: time.Now()}
		sm.sessions[cookie.Value] = session
	}

	session.LastAccessed = time.Now()
	return session
}

func (sm *SessionManager) cleanupSessions() {
	for {
		time.Sleep(1 * time.Hour)
		sm.mutex.Lock()
		for id, session := range sm.sessions {
			if time.Since(session.LastAccessed) > 24*time.Hour {
				delete(sm.sessions, id)
			}
		}
		sm.mutex.Unlock()
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse the multipart form data
    err := r.ParseMultipartForm(32 << 20) // 32 MB max memory
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Get the session
    session := sessionManager.getSession(w, r)

    // Get the files from the request
    files := r.MultipartForm.File["file-input"]
    
    type ProcessedFile struct {
        Index    int
        Filename string
        Error    string
    }
    
    processedFiles := make([]ProcessedFile, len(files))
    var wg sync.WaitGroup

    // Pre-assign file counters
    fileCounters := make([]int, len(files))
    for i := range files {
        session.FileCounter++
        fileCounters[i] = session.FileCounter
    }

    // Create a buffered channel to limit concurrency
    semaphore := make(chan struct{}, 5) // Adjust this number based on your needs

    for i, fileHeader := range files {
        wg.Add(1)
        go func(i int, fileHeader *multipart.FileHeader, fileCounter int) {
            defer wg.Done()
            semaphore <- struct{}{} // Acquire semaphore
            defer func() { <-semaphore }() // Release semaphore

            file, err := fileHeader.Open()
            if err != nil {
                processedFiles[i] = ProcessedFile{Index: i, Error: fmt.Sprintf("Error opening file %s: %v", fileHeader.Filename, err)}
                return
            }
            defer file.Close()

            // Read the file content
            fileContent, err := ioutil.ReadAll(file)
            if err != nil {
                processedFiles[i] = ProcessedFile{Index: i, Error: fmt.Sprintf("Error reading file %s: %v", fileHeader.Filename, err)}
                return
            }

            // Calculate hash of the file content
            hash := sha256.Sum256(fileContent)
            hashString := hex.EncodeToString(hash[:])

            // Create session directory
            sessionDir := filepath.Join("workdir", "web", session.ID)
            err = os.MkdirAll(sessionDir, os.ModePerm)
            if err != nil {
                processedFiles[i] = ProcessedFile{Index: i, Error: fmt.Sprintf("Error creating session directory: %v", err)}
                return
            }

            // Create hash directory
            hashDir := filepath.Join(sessionDir, hashString)
            err = os.MkdirAll(filepath.Join(hashDir, "input"), os.ModePerm)
            err = os.MkdirAll(filepath.Join(hashDir, "output"), os.ModePerm)
            if err != nil {
                processedFiles[i] = ProcessedFile{Index: i, Error: fmt.Sprintf("Error creating hash directory: %v", err)}
                return
            }

            inputPath := filepath.Join(hashDir, "input", fileHeader.Filename)
            outputDir := filepath.Join(hashDir, "output")

            // Check if any file exists in the output directory
            outputFiles, err := ioutil.ReadDir(outputDir)
            if err == nil && len(outputFiles) > 0 {
                // File already processed
                processedFiles[i] = ProcessedFile{Index: i, Filename: filepath.Join(session.ID, hashString, "output", outputFiles[0].Name())}
                return
            }

            // Write the file content to the input file
            err = ioutil.WriteFile(inputPath, fileContent, 0644)
            if err != nil {
                processedFiles[i] = ProcessedFile{Index: i, Error: fmt.Sprintf("Error writing input file for %s: %v", fileHeader.Filename, err)}
                return
            }

            // Process the file
            outputFilename := mediaprocessor.GenerateOrderedFilename(fileCounter, filepath.Ext(fileHeader.Filename))
            outputPath := filepath.Join(outputDir, outputFilename)
            err = mediaprocessor.ProcessLocalMediaFile(inputPath, outputPath)
            if err != nil {
                processedFiles[i] = ProcessedFile{Index: i, Error: fmt.Sprintf("Error processing file %s: %v", fileHeader.Filename, err)}
                return
            }

            processedFiles[i] = ProcessedFile{Index: i, Filename: filepath.Join(session.ID, hashString, "output", outputFilename)}
        }(i, fileHeader, fileCounters[i])
    }

    wg.Wait()

    // Return the processed files information and any errors
    w.Header().Set("Content-Type", "text/html")
    for _, pf := range processedFiles {
        if pf.Error != "" {
            fmt.Fprintf(w, "<li class='text-red-500 py-2'>%s</li>", pf.Error)
        } else {
            downloadPath := filepath.Join("/download", pf.Filename)
            fmt.Fprintf(w, "<li class='flex justify-between items-center py-2'>"+
                "<span>File processed: %s</span>"+
                "<a href='%s' download class='text-blue-500 hover:text-blue-700'>"+
                "<svg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round' class='feather feather-download'><path d='M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4'></path><polyline points='7 10 12 15 17 10'></polyline><line x1='12' y1='15' x2='12' y2='3'></line></svg>"+
                "</a></li>", filepath.Base(pf.Filename), downloadPath)
        }
    }

    // Log errors to console for debugging
    for _, pf := range processedFiles {
        if pf.Error != "" {
            fmt.Println(pf.Error)
        }
    }
}
