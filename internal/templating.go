package internal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

// func ProcessTemplate(inputFile string, outputFile string, data map[string]interface{}) {
func ProcessTemplate(inputFile string, outputFile string, data UADeployConfig) {

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

func GetMaprConfig() DFConfig {
	// Define a regex pattern to match IPv4 addresses
	ipv4Pattern := `(\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b)`
	// Define regex patterns for accesskey and secretkey
	accessKeyPattern := `"accesskey":"([A-Z0-9]+)"`
	secretKeyPattern := `"secretkey":"([A-Z0-9]+)"`

	// Compile the regular expressions
	ipV4re, err := regexp.Compile(ipv4Pattern)
	if err != nil {
		log.Fatalf("failed to compile regex: %v", err)
	}
	accessKeyRe, err := regexp.Compile(accessKeyPattern)
	if err != nil {
		log.Fatalf("failed to compile accesskey regex: %v", err)
	}
	secretKeyRe, err := regexp.Compile(secretKeyPattern)
	if err != nil {
		log.Fatalf("failed to compile secretkey regex: %v", err)
	}

	dfConfig := DFConfig{}

	fileContent := ReadFile("/tmp/cldb_nodes.json")
	dfConfig.CldbNodes = ipV4re.FindString(string(fileContent))

	fileContent = ReadFile("/tmp/rest_nodes.json")
	dfConfig.RestNodes = ipV4re.FindString(string(fileContent))

	fileContent = ReadFile("/tmp/s3_nodes.json")
	dfConfig.S3Nodes = ipV4re.FindString(string(fileContent))

	fileContent = ReadFile("/tmp/s3_keys.json")
	accessKeyMatch := accessKeyRe.FindStringSubmatch(string(fileContent))
	if len(accessKeyMatch) < 2 {
		log.Fatalf("accesskey not found")
	}
	dfConfig.AccessKey = base64.StdEncoding.EncodeToString([]byte(accessKeyMatch[1]))

	secretKeyMatch := secretKeyRe.FindStringSubmatch(string(fileContent))
	if len(secretKeyMatch) < 2 {
		log.Fatalf("secretkey not found")
	}
	dfConfig.SecretKey = base64.StdEncoding.EncodeToString([]byte(secretKeyMatch[1]))

	return dfConfig
}
