package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/gocolly/colly"
)

var museumObjects = []MuseumObject{}

var urls = []string{}
var mainUrl = "https://colecciones.banrepcultural.org/page/coleccin-arqueolgica-de-los-museos-del-oro/6357a765e27d753f221c6160?pgn="

func main() {
	fmt.Println("WEB SCRAPING GOLD MUSEUM")
	fmt.Println("Getting object urls...")
	getUrls()
	writeUrlsToCSV()
	fmt.Println("Getting objects...")
	for _, url := range urls {
		getObjects(url)
	}
	fmt.Println("Got", len(museumObjects), "objects")
	writeObjectsToCSV()
	writeObjectsToJSON()
	// Creates separated CSV files for each property for debugging
	// checkObjectProperties()
	fmt.Println("End of script")
}

func getUrls() {
	// Visit all search pages and get links for each object's page
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
	for i := 0; i <= 12; i++ {
		c.OnHTML("div.card", func(e *colly.HTMLElement) {
			newUrl := e.ChildAttr("a", "href")
			if !contains(urls, newUrl) {
				urls = append(urls, e.ChildAttr("a", "href"))
			}
		})
		c.Visit(fmt.Sprintf("%s%d", mainUrl, i))
	}
	fmt.Println("Got", len(urls), "links")
}

func writeUrlsToCSV() {
	fmt.Println("Saving object urls in CSV...")
	file, err := os.Create("GoldMuseumUrls.csv")
	if err != nil {
		fmt.Println("Failed to create output CSV file", err)
	}
	defer file.Close()
	// initializing a file writer
	writer := csv.NewWriter(file)
	// writing the CSV headers
	headers := []string{
		"url",
	}
	writer.Write(headers)
	for _, url := range urls {
		// converting url as part of an array of strings
		record := []string{
			url,
		}
		// adding a CSV record to the output file
		writer.Write(record)
	}
	defer writer.Flush()
	fmt.Println("Finished writing to", file.Name())
}

func getObjects(url string) {

	object := MuseumObject{}
	object.PageUrl = url
	optionalData := []string{}
	dataCounter := 0
	dataCounter2 := 0
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

	c.OnRequest(func(r *colly.Request) {
		// fmt.Println("Visiting Museo del Oro at:", url)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		// fmt.Println("Response received!")
	})

	// Images
	c.OnHTML("img.m-auto", func(e *colly.HTMLElement) {
		img := ObjectImg{}
		img.Url = strings.TrimSpace(e.Attr("src"))
		img.AltText = strings.TrimSpace(e.Attr("alt"))
		object.Imgs = append(object.Imgs, img)
	})

	// Images author
	c.OnHTML("p", func(e *colly.HTMLElement) {
		object.ImgsAuthor = strings.TrimSpace(e.Text)
	})

	// Big collection
	c.OnHTML("h3.fs-2", func(e *colly.HTMLElement) {
		object.BigCollection = strings.TrimSpace(e.Text)

	})

	// Name
	c.OnHTML("h2.fs-1", func(e *colly.HTMLElement) {
		object.Name = strings.TrimSpace(e.Text)

	})

	// Region, Period
	c.OnHTML("h3.fs-3", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		char := text[0]
		if !unicode.IsDigit(rune(char)) {
			object.Region = text
		}
		if unicode.IsDigit(rune(char)) {
			object.Period = text
		}
	})

	// Collection, Origin, Current location
	c.OnHTML("a.col-md-8", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if dataCounter == 0 {
			object.Collection = text
			dataCounter++
		} else if dataCounter == 1 {
			originString := strings.ReplaceAll(text, "(", "")
			originString = strings.ReplaceAll(originString, ")", "")
			originArr := strings.Split(originString, "/")
			if len(originArr) == 3 {
				object.OriginCity = strings.TrimSpace(originArr[0])
				object.OriginDepartment = originArr[2]
				object.OriginCountry = originArr[1]
			} else if len(originArr) == 2 {
				if strings.Contains(originArr[0], "Bog") {
					object.OriginCity = strings.TrimSpace(originArr[0])
					object.OriginDepartment = "Cundinamarca"
					object.OriginCountry = originArr[1]
				} else {
					object.OriginDepartment = strings.TrimSpace(originArr[0])
					object.OriginCountry = originArr[1]
				}
			} else {
				object.OriginCountry = strings.TrimSpace(originArr[0])
			}
			dataCounter++
		} else if dataCounter == 2 {
			object.CurrentLocation = strings.TrimSpace(e.Text)
			dataCounter++
		}
	})

	// Material, Function, Volume data, Height, Width, Length, Register number, Technique, Description
	c.OnHTML("div.col-md-8", func(e *colly.HTMLElement) {
		if dataCounter2 == 0 {
			innerHTML, _ := e.DOM.Html()
			innerHTML = strings.ReplaceAll(innerHTML, "<br>", " ")
			innerHTML = strings.ReplaceAll(innerHTML, "<br/>", " ")
			object.Material = strings.TrimSpace(innerHTML)
			dataCounter2++
		} else if dataCounter2 == 1 {
			object.Function = strings.TrimSpace(e.Text)
			dataCounter2++
		} else if dataCounter2 == 2 {
			// Get inner HTML of the div
			innerHTML2, _ := e.DOM.Html()
			innerHTML2 = strings.ReplaceAll(innerHTML2, "<br>", " ")
			innerHTML2 = strings.ReplaceAll(innerHTML2, "<br/>", " ")
			text := strings.TrimSpace(innerHTML2)
			optionalData = append(optionalData, text)
		}
		for _, data := range optionalData {
			volData := strings.Contains(data, "alto/largo")
			regNum := unicode.IsDigit(rune(data[len(data)-1]))
			tech := len(data) < 90 && !volData && !strings.Contains(data, "#")
			desc := len(data) >= 20 && !volData
			if volData {
				splitData := strings.Split(data, " ")
				// checkVolumeData(splitData)
				h := strings.ReplaceAll(splitData[2], ",", ".")
				w := strings.ReplaceAll(splitData[6], ",", ".")
				object.Height = h
				object.Width = w
				if len(splitData) > 10 {
					l := strings.ReplaceAll(splitData[10], ",", ".")
					object.Length = l
				}
			} else if regNum && !volData && !desc {
				object.CatalogueId = data
			} else if tech && !volData && !regNum {
				object.Technique = data
				object.Technique = strings.ReplaceAll(object.Technique, "\n", " ")
			} else if desc && !volData && !regNum && !tech {
				object.Description = strings.ReplaceAll(data, "\n", " ")
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		// fmt.Println("Finished visit!")
	})

	c.Visit(url)

	object.ScrapingDate = time.Now().String()
	museumObjects = append(museumObjects, object)
	fmt.Println("Got", object.Name, object.CatalogueId)
}

// func checkVolumeData(splitData []string) {
// 	// Check data received in volumeData to save it properly in object properties
// 	if len(splitData) > 6 {
// 		fmt.Println(splitData[0], splitData[3], splitData[6])
// 	} else {
// 		fmt.Println(splitData[0], splitData[3])
// 	}
// }

func writeObjectsToCSV() {
	fmt.Println("Saving objects data in CSV...")
	file2, err := os.Create("GoldMuseumObjects.csv")
	if err != nil {
		fmt.Println("Failed to create output CSV file", err)
	}
	defer file2.Close()
	// initializing a file writer
	writer2 := csv.NewWriter(file2)
	// writing the CSV headers
	headers2 := []string{
		"CatalogueId",
		"Name",
		"PageUrl",
		"BigCollection",
		"Collection",
		"Region",
		"Period",
		"OriginCountry",
		"OriginDepartment",
		"OriginCity",
		"CurrentLocation",
		"Material",
		"Function",
		"Heigth",
		"Widht",
		"Length",
		"Technique",
		"Description",
		"Images",
		"ImagesAuthor",
		"ScrapingDate",
	}
	writer2.Write(headers2)
	for _, obj := range museumObjects {
		imagesString := ""
		for i, img := range obj.Imgs {
			if i == len(obj.Imgs)-1 {
				imagesString += img.Url + " " + img.AltText
				break
			}
			imagesString += img.Url + " " + img.AltText + " - "
		}
		// Object properties as list of strings
		record := []string{
			obj.CatalogueId,
			obj.PageUrl,
			obj.Name,
			obj.BigCollection,
			obj.Collection,
			obj.Region,
			obj.Period,
			obj.OriginCountry,
			obj.OriginDepartment,
			obj.OriginCity,
			obj.CurrentLocation,
			obj.Material,
			obj.Function,
			obj.Height,
			obj.Width,
			obj.Length,
			obj.Technique,
			obj.Description,
			imagesString,
			obj.ImgsAuthor,
			obj.ScrapingDate,
		}
		writer2.Write(record)
	}
	defer writer2.Flush()
	fmt.Println("Finished writing to", file2.Name())
}

func writeObjectsToJSON() {
	fmt.Println("Saving objects in JSON...")
	data := JSONResult{
		museumObjects,
	}
	file, err := os.Create("GoldMuseumObjects.json")
	if err != nil {
		fmt.Println("Failed to create output CSV file", err)
		return
	}
	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Error creating JSON objects", err)
		return
	} else {
		// writer := os.WriteFile("GoldMuseumObjects.json", content, 0644)
		_, err := file.Write(content)
		if err != nil {
			fmt.Println("Error writing to JSON file", err)
			return
		} else {
			fmt.Println("Finished writing to", file.Name())
		}
	}
	defer file.Close()
}

// func checkObjectProperties() {
// 	fmt.Println("Saving objects properties for checking data state...")
// 	mo := MuseumObject{}
// 	o := reflect.ValueOf(&mo).Elem()
// 	objProps := []string{}
// 	for i := 0; i < o.NumField(); i++ {
// 		varName := o.Type().Field(i).Name
// 		objProps = append(objProps, varName)
// 	}
// 	for _, prop := range objProps {
// 		fileName := fmt.Sprintf("%s.csv", prop)
// 		file2, err := os.Create(fileName)
// 		if err != nil {
// 			fmt.Println("Failed to create output CSV file", err)
// 		}
// 		defer file2.Close()
// 		// initializing a file writer
// 		writer2 := csv.NewWriter(file2)
// 		// writing the CSV headers
// 		headers2 := []string{
// 			prop,
// 		}
// 		writer2.Write(headers2)
// 		for _, obj := range museumObjects {
// 			var record []string
// 			if prop == "imagesString" {
// 				imagesString := ""
// 				for _, img := range obj.Imgs {
// 					imagesString += "*" + img.Url + "-" + img.AltText + "*"
// 				}
// 				record = []string{
// 					imagesString,
// 				}
// 			} else {
// 				record = []string{
// 					obj.getProperty(prop),
// 				}
// 			}
// 			writer2.Write(record)
// 		}
// 		defer writer2.Flush()
// 		fmt.Println("Finished writing to", file2.Name())
// 	}

// }

// func (m *MuseumObject) getProperty(propName string) string {
// 	return reflect.ValueOf(m).Elem().FieldByName(propName).String()
// }

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

type MuseumObject struct {
	CatalogueId      string      `json:"catalogueId"`
	PageUrl          string      `json:"pageUrl"`
	Name             string      `json:"name"`
	BigCollection    string      `json:"bigCollection"`
	Collection       string      `json:"collection"`
	Region           string      `json:"region"`
	Period           string      `json:"period"`
	OriginCountry    string      `json:"originCountry"`
	OriginDepartment string      `json:"originDepartment"`
	OriginCity       string      `json:"originCity"`
	CurrentLocation  string      `json:"currentLocation"`
	Material         string      `json:"material"`
	Function         string      `json:"function"`
	Height           string      `json:"height"`
	Width            string      `json:"width"`
	Length           string      `json:"length"`
	Technique        string      `json:"technique"`
	Description      string      `json:"description"`
	Imgs             []ObjectImg `json:"imgs"`
	ImgsAuthor       string      `json:"imgsAuthor"`
	ScrapingDate     string      `json:"scrapingDate"`
}

// func (m *MuseumObject) printObject() {
// 	fmt.Printf("\n%+v\n\n", m)
// 	fmt.Println("Catalogue id:", m.CatalogueId)
// 	fmt.Println("Url:", m.PageUrl)
// 	fmt.Println("Name:", m.Name)
// 	fmt.Println("Big collection:", m.BigCollection)
// 	fmt.Println("Collection:", m.Collection)
// 	fmt.Println("Region:", m.Region)
// 	fmt.Println("Period:", m.Period)
// 	fmt.Println("Origin country:", m.OriginCountry)
// 	fmt.Println("Origin department:", m.OriginDepartment)
// 	fmt.Println("Origin city:", m.OriginCity)
// 	fmt.Println("Current location:", m.CurrentLocation)
// 	fmt.Println("Material:", m.Material)
// 	fmt.Println("Function:", m.Function)
// 	fmt.Println("Height:", m.Height)
// 	fmt.Println("Width:", m.Width)
// 	fmt.Println("Length:", m.Length)
// 	fmt.Println("Technique:", m.Technique)
// 	fmt.Println("Description:", m.Description)
// 	fmt.Println("Images:", m.Imgs)
// 	fmt.Println("Images author:", m.ImgsAuthor)
// 	fmt.Println("Scraping date:", m.ScrapingDate)
// }

type ObjectImg struct {
	Url     string `json:"url"`
	AltText string `json:"altText"`
}

type JSONResult struct {
	Data []MuseumObject `json:"data"`
}
