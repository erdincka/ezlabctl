package internal

import (
	"embed"
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templatesFS embed.FS

func ProcessTemplates(targetDir string, data UAConfig) {
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		log.Fatal("Error creating yaml directory: ", err)
	}

	// log.Println(templatesFS.ReadDir("templates"))
	funcMap := template.FuncMap{
		"base64": func(s string) string {
			return base64.StdEncoding.EncodeToString([]byte(s))
		},
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.yaml")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	// Iterate over each template in the set
	for _, templateName := range tmpl.Templates() {
		// Build the output file path
		outputFilePath := filepath.Join(targetDir, filepath.Base(templateName.Name()))

		// Create the output file
		outFile, err := os.Create(outputFilePath)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
		}
		defer outFile.Close()

		// Execute the template with the provided data
		err = tmpl.ExecuteTemplate(outFile, templateName.Name(), data)
		if err != nil {
			log.Fatalf("failed to execute template: %v", err)
		}

		log.Println("Processed template:", templateName.Name())
	}

	log.Println("YAML files ready")
}
