package internal

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

// func ProcessTemplate(inputFile string, outputFile string, data map[string]interface{}) {
func ProcessTemplate(inputFile string, outputFile string, data UADeployConfig) {

	templateContent, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("failed to read template file: %v", err)
	}

	tmpl, err := template.New(inputFile).Parse(string(templateContent))
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

	fmt.Printf("Template saved to %s\n", outputFile)
}

// func ValuesFromYamlFile(dataFile string) (map[string]interface{}, error) {
//     data, err := os.Open(dataFile)
//     if err != nil {
//         return nil, errors.New(err.Error())
//     }
//     defer data.Close()
//     var values map[string]interface{}
//     return values, nil
// }

// func Parse(templateFile, dataFile, outputFile string) error {
//     tmpl, err := template.ParseFiles(templateFile)
//     if err != nil {
//         return errors.New(err.Error())
//     }
//     values, err := ValuesFromYamlFile(dataFile)
//     if err != nil {
//         return err
//     }
//     output, err := os.Create(outputFile)
//     if err != nil {
//         return errors.New(err.Error())
//     }
//     defer output.Close()

//     err = tmpl.Execute(output, values)
//     if err != nil {
//         return errors.New(err.Error())
//     }

// 	return nil
// }

// returns all files with .yaml extension in ./templates folder
func GetTemplateFiles() ([]string, error) {
	templateDir := "./templates"
	files, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			fileList = append(fileList, templateDir + "/" + file.Name())
		}
	}

	return fileList, nil
}