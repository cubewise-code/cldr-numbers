package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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
	Decimal     string `json:"decimal,omitempty"`
	Group       string `json:"group,omitempty"`
	PercentSign string `json:"percentSign,omitempty"`
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

// Result is the format that is written to disk
type Result struct {
	Package Package            `json:"package"`
	Symbols map[string]Symbols `json:"symbols"`
}

func main() {

	dir := flag.String("dir", "cldr-numbers-full", "Directory when cldr-numbers-full is located")

	type Defaults struct {
		Decimal     string
		Group       string
		PercentSign string
	}
	defaults := Defaults{
		Decimal:     ".",
		Group:       ",",
		PercentSign: "%",
	}

	flag.Parse()

	full := Result{}
	full.Symbols = map[string]Symbols{}

	condensed := Result{}
	condensed.Symbols = map[string]Symbols{}

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
					full.Symbols[key] = val.Numbers.Symbols
					count := 0
					if val.Numbers.Symbols.Decimal == defaults.Decimal {
						val.Numbers.Symbols.Decimal = ""
						count++
					}
					if val.Numbers.Symbols.Group == defaults.Group {
						val.Numbers.Symbols.Group = ""
						count++
					}
					if val.Numbers.Symbols.PercentSign == defaults.PercentSign {
						val.Numbers.Symbols.PercentSign = ""
						count++
					}
					if count != 3 {
						// Don't add if everything was omited
						condensed.Symbols[key] = val.Numbers.Symbols
					}
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
				full.Package = *pack
				condensed.Package = *pack
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Unable to iterate file: %s", err)
		return
	}

	if len(full.Symbols) == 0 {
		fmt.Print("No locales found, check path to cldr-numbers-full")
		return
	}

	err = writeToFile("cldr-numbers", full, false)
	if err != nil {
		fmt.Print(err)
		return
	}

	err = writeToFile("cldr-numbers-condensed", condensed, false)
	if err != nil {
		fmt.Print(err)
		return
	}

	err = writeToFile("cldr-numbers", full, true)
	if err != nil {
		fmt.Print(err)
		return
	}

	err = writeToFile("cldr-numbers-condensed", condensed, true)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Println("Complete")

}

func writeToFile(name string, result Result, javaScript bool) error {
	if _, err := os.Stat("dist"); os.IsNotExist(err) {
		os.Mkdir("dist", 0644)
	}

	extension := "json"
	if javaScript {
		extension = "js"
	}

	pretty, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to marshal file: %s", err)
	}

	min, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Unable to marshal file: %s", err)
	}

	if javaScript {
		re := regexp.MustCompile(`("([A-Za-z_]+)"):`)
		sPretty := string(pretty)
		sPretty = re.ReplaceAllString(sPretty, `$2:`)
		pretty = []byte(sPretty)
		sMin := string(min)
		sMin = re.ReplaceAllString(sMin, `$2:`)
		min = []byte(sMin)
	}

	err = ioutil.WriteFile("dist/"+name+"."+extension, pretty, 0644)
	if err != nil {
		return fmt.Errorf("Unable to write file: %s", err)
	}

	err = ioutil.WriteFile("dist/"+name+".min."+extension, min, 0644)
	if err != nil {
		return fmt.Errorf("Unable to write file: %s", err)
	}

	return nil
}
