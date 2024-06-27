package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const (
	ATAP_URL = "https://atap.co/malaysia/en/professionals/"
)

func main() {
	atapCompanyInfos := []*AtapCompanyInfo{}
	// for i := 1; i <= 405; i++ {
	// 	url := ATAP_URL
	// 	if i > 1 {
	// 		url = fmt.Sprintf("%s?page=%d", ATAP_URL, i)
	// 	}
	// 	infos := getLinks(url)
	// 	atapCompanyInfos = append(atapCompanyInfos, infos...)
	// 	time.Sleep(1 * time.Second)
	// }

	// write to json
	// file, err := os.Create("atap.json")
	// if err != nil {
	// 	log.Fatal("Cannot create file", err)
	// }
	// defer file.Close()
	// b, err := json.MarshalIndent(atapCompanyInfos, "", "  ")
	// if err != nil {
	// 	log.Fatal("Cannot marshal data", err)
	// }
	// file.Write(b)

	// read from json
	file, err := os.Open("atap.json")
	if err != nil {
		log.Fatal("Cannot open file", err)
	}
	defer file.Close()
	atapCompanyInfos = []*AtapCompanyInfo{}
	err = json.NewDecoder(file).Decode(&atapCompanyInfos)
	if err != nil {
		log.Fatal("Cannot decode file", err)
	}

	ataps := []*AtapCompany{}
	for i, info := range atapCompanyInfos {
		fmt.Printf("Processing %d/%d: %s\n", i+1, len(atapCompanyInfos), info.CompanyName)
		contact := getContactDetail(info.Link)
		atap := &AtapCompany{
			CompanyInfo:    info,
			ContactDetails: contact,
		}
		ataps = append(ataps, atap)
		time.Sleep(500 * time.Millisecond)
	}
	writeAtapToCSV(ataps)
}

type AtapCompany struct {
	CompanyInfo    *AtapCompanyInfo
	ContactDetails *AtapContactDetail
}

type AtapContactDetail struct {
	ContactName string   `selector:"span.contact-name"`
	Address     string   `selector:"span.contact-address"`
	Telephones  []string `selector:"span.contact-telephone"`
}

type AtapCompanyInfo struct {
	CompanyName string
	Categories  []string
	Location    string
	Link        string
}

func writeAtapToCSV(data []*AtapCompany) {
	// Write to writeAtapToCSV
	file, err := os.Create("atap.csv")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Company Name", "Categories", "Location", "Contact Name", "Address", "Tel1", "Tel2", "Link"})
	for _, atap := range data {
		telephones := make([]string, 2)
		if len(atap.ContactDetails.Telephones) == 1 {
			telephones[0] = atap.ContactDetails.Telephones[0]
			telephones[1] = ""
		} else if len(atap.ContactDetails.Telephones) == 2 {
			telephones = atap.ContactDetails.Telephones
		}

		atap.ContactDetails.Telephones = telephones

		writer.Write([]string{
			atap.CompanyInfo.CompanyName,
			strings.Join(atap.CompanyInfo.Categories, ","),
			atap.CompanyInfo.Location,
			atap.ContactDetails.ContactName,
			atap.ContactDetails.Address,
			atap.ContactDetails.Telephones[0],
			atap.ContactDetails.Telephones[1],
			atap.CompanyInfo.Link,
		})
	}
}

func getLinks(url string) []*AtapCompanyInfo {
	c := colly.NewCollector()

	results := []*AtapCompanyInfo{}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div.vendor-card", func(_ int, el *colly.HTMLElement) {
			data := &AtapCompanyInfo{}
			data.CompanyName = strings.TrimSpace(el.ChildText("div.vendor-desc > a"))
			data.Link = strings.TrimSpace(el.ChildAttr("div.vendor-desc > a", "href"))
			el.ForEach("div.cat-link.vendor-expertise a", func(_ int, el *colly.HTMLElement) {
				data.Categories = append(data.Categories, strings.TrimSpace(el.Text))
			})
			fmt.Println(el.ChildText("div.cat-link.vendor-expertise > a"))
			data.Location = strings.TrimSpace(el.ChildText(`div[class="cat-link"] a`))
			results = append(results, data)
		})
	})

	c.Visit(url)

	for _, r := range results {
		fmt.Printf("%+v\n", r)
	}

	return results
}

func getContactDetail(url string) *AtapContactDetail {
	c := colly.NewCollector()
	data := &AtapContactDetail{}

	c.OnHTML("div.vendor-contact", func(e *colly.HTMLElement) {
		data.ContactName = strings.TrimSpace(e.ChildText("span.contact-name"))
		data.Address = strings.TrimSpace(e.ChildText("span.contact-address"))

		e.ForEach("span.contact-telephone", func(_ int, el *colly.HTMLElement) {
			telephone := el.Text
			last4Digits := el.ChildAttr("a", "data-mask")
			data.Telephones = append(
				data.Telephones,
				strings.TrimSpace(strings.ReplaceAll(telephone, "xxxx", last4Digits)),
			)

		})
	})

	c.Visit(url)
	return data
}
