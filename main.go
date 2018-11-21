package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Package contains the details about the source cldr
type Package struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	PeerDependencies struct {
		CldrCore string `json:"cldr-core"`
	} `json:"peerDependencies"`
	Homepage    string `json:"homepage"`
	Author      string `json:"author"`
	Maintainers []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		URL   string `json:"url"`
	} `json:"maintainers"`
	Repository struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"repository"`
	Licenses []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"licenses"`
	Bugs string `json:"bugs"`
}

// Symbols contains the infor we want to extract for each locale
type Symbols struct {
	Decimal     string `json:"decimal"`
	Group       string `json:"group"`
	PercentSign string `json:"percentSign"`
}

// Locale has the details of the settings
type Locale struct {
	Identity struct {
		Version struct {
			CldrVersion string `json:"_cldrVersion"`
		} `json:"version"`
	} `json:"identity"`
	Numbers struct {
		MinimumGroupingDigits string  `json:"minimumGroupingDigits"`
		Symbols               Symbols `json:"symbols-numberSystem-latn"`
		DecimalFormats        struct {
			Standard string `json:"standard"`
		} `json:"decimalFormats-numberSystem-latn"`
		ScientificFormats struct {
			Standard string `json:"standard"`
		} `json:"scientificFormats-numberSystem-latn"`
		PercentFormats struct {
			Standard string `json:"standard"`
		} `json:"percentFormats-numberSystem-latn"`
	} `json:"numbers"`
}

// Numbers holds the info we want to generate a single file with details about each locale
type Numbers struct {
	Main map[string]Locale `json:"main"`
}

func main() {

	dir := flag.String("dir", "cldr-numbers-full", "Directory when cldr-numbers-full is located")

	flag.Parse()

	type Result struct {
		Package Package            `json:"package"`
		Symbols map[string]Symbols `json:"symbols"`
	}

	result := Result{}
	result.Symbols = map[string]Symbols{}

	err := filepath.Walk(*dir, func(filePath string, file os.FileInfo, err error) error {
		if file != nil {
			if file.Name() == "numbers.json" {
				bytes, err := ioutil.ReadFile(filePath)
				if err != nil {
					fmt.Printf("Unable to read file %s: %s", filePath, err)
				}
				numbers := &Numbers{}
				err = json.Unmarshal(bytes, numbers)
				if err != nil {
					fmt.Printf("Unable to parse file %s: %s", filePath, err)
				}
				for key, val := range numbers.Main {
					result.Symbols[key] = val.Numbers.Symbols
				}
			} else if file.Name() == "package.json" {
				bytes, err := ioutil.ReadFile(filePath)
				if err != nil {
					fmt.Printf("Unable to read file %s: %s", filePath, err)
				}
				pack := &Package{}
				err = json.Unmarshal(bytes, pack)
				if err != nil {
					fmt.Printf("Unable to parse file %s: %s", filePath, err)
				}
				result.Package = *pack
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Unable to iterate file: %s", err)
		return
	}

	if len(result.Symbols) == 0 {
		fmt.Print("No locales found, check path to cldr-numbers-full")
		return
	}

	if _, err := os.Stat("dist"); os.IsNotExist(err) {
		os.Mkdir("dist", 0644)
	}

	pretty, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Unable to marshal file: %s", err)
		return
	}

	err = ioutil.WriteFile("dist/cldr-numbers.json", pretty, 0644)
	if err != nil {
		fmt.Printf("Unable to write file: %s", err)
		return
	}

	min, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Unable to marshal file: %s", err)
		return
	}

	err = ioutil.WriteFile("dist/cldr-numbers.min.json", min, 0644)
	if err != nil {
		fmt.Printf("Unable to write file: %s", err)
		return
	}

	fmt.Println("Complete")

}
