package internal

import (
	"bytes"
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

func ProcessTemplates(targetDir string, data UAConfig) {
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		log.Fatal("Error creating yaml directory: ", err)
	}

	for _, file := range GetDeploySteps() {
		// go func(f string) {
			ProcessTemplate("templates/" + file, targetDir + "/" + filepath.Base(file), data)
			log.Println("Processing: " + file)
		// }(file)
	}

	log.Println("YAML files ready")
}

func ProcessTemplate(inputFile string, outputFile string, data UAConfig) {

	templateContent, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("failed to read template file: %v", err)
	}

	tmpl, err := template.New(inputFile).Funcs(template.FuncMap{
		"base64": func(s string) string {
			return base64.StdEncoding.EncodeToString([]byte(s))
		},
	}).Parse(string(templateContent))
	if err != nil {
		log.Fatalf("failed to parse: %v", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}

	err = os.WriteFile(outputFile, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}

	log.Printf("Template saved to %s\n", outputFile)
}
