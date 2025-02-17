package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	sf "strings"
	"time"

	"github.com/gocolly/colly"
)

type location struct {
	Zone     string `json:"zone"`
	Locality string `json:"locality"`
	Hood     string `json:"hood"`
	Address  string `json:"address"`
}

type price struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type squareMeters struct {
	Covered string `json:"covered"`
	Total   string `json:"total"`
}

type realState struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type house struct {
	Id               string       `json:"id"`
	ScrapeDate       string       `json:"scrape_date"`
	PublicationDate  string       `json:"publication_date"`
	LocationInfo     location     `json:"location_info"`
	PriceInfo        price        `json:"price_info"`
	SquareMetersInfo squareMeters `json:"square_meters_info"`
	RealStateInfo    realState    `json:"real_state_info"`
	CreditSuitable   bool         `json:"credital_suitable"`
	Link             string       `json:"link"`
}

func GetLocationInfo(rawLocation string) location {
	splittedLocation := sf.Split(rawLocation, ",")

	if len(splittedLocation) == 5 {
		return location{
			Zone:     sf.TrimSpace(splittedLocation[4]),
			Locality: sf.TrimSpace(splittedLocation[3]),
			Hood:     sf.TrimSpace(splittedLocation[1]) + sf.TrimSpace(splittedLocation[2]),
			Address:  sf.TrimSpace(splittedLocation[0]),
		}
	} else if len(splittedLocation) == 3 {
		return location{
			Zone:     sf.TrimSpace(splittedLocation[2]),
			Locality: sf.TrimSpace(splittedLocation[1]),
			Hood:     sf.TrimSpace("unknown"),
			Address:  sf.TrimSpace(splittedLocation[0]),
		}
	}

	return location{
		Zone:     sf.TrimSpace(splittedLocation[3]),
		Locality: sf.TrimSpace(splittedLocation[2]),
		Hood:     sf.TrimSpace(splittedLocation[1]),
		Address:  sf.TrimSpace(splittedLocation[0]),
	}
}

func main() {
	houses := make([]house, 0)
	domains := [2]string{"inmuebles.mercadolibre.com.ar", "casa.mercadolibre.com.ar"}
	mainCollector := colly.NewCollector(colly.AllowedDomains(domains[0], domains[1]))
	innerCollector := mainCollector.Clone()

	// Find and visit all links
	mainCollector.OnHTML("div.poly-card__content", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a", "href")

		hash := md5.New()
		hash.Write([]byte(link))

		houses = append(houses, house{
			Id:              hex.EncodeToString(hash.Sum(nil)),
			ScrapeDate:      time.Now().Format(time.RFC3339),
			PublicationDate: "unknown",
			LocationInfo:    GetLocationInfo(e.ChildText("span.poly-component__location")),
			PriceInfo: price{
				sf.Replace(e.ChildText("span.andes-money-amount__fraction"), ".", "", 1),
				e.ChildText("span.andes-money-amount__currency-symbol"),
			},
			SquareMetersInfo: squareMeters{"unknown", "unknown"},
			RealStateInfo:    realState{"unknown", "unknown"},
			CreditSuitable:   false,
			Link:             link,
		})

		innerCollector.Visit(link)
	})

	mainCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	// TODO: ver como puedo acceder a los valores de la tabla. Pareciera ser que, de manera aleatoria,a veces puedo acceder a los valores y a veces no.
	innerCollector.OnHTML("div#ui-pdp-main-container", func(e *colly.HTMLElement) {
		link := e.Request.URL.String()
		value := 0

		hash := md5.New()
		hash.Write([]byte(link))

		i := slices.IndexFunc(houses, func(h house) bool {
			return h.Id == hex.EncodeToString(hash.Sum(nil))
		})

		if i != -1 {
			publicationDate := e.ChildText("div.ui-pdp-seller-validated > p.ui-pdp-color--GRAY.ui-pdp-size--XSMALL.ui-pdp-family--REGULAR.ui-pdp-seller-validated__title")

			if sf.Contains(publicationDate, "identidad verificada") {
				publicationDate = e.ChildText("p.ui-pdp-color--GRAY.ui-pdp-size--XSMALL.ui-pdp-family--REGULAR.ui-pdp-header__bottom-subtitle")
			}

			houses[i].PublicationDate = publicationDate
			houses[i].RealStateInfo.Name = e.ChildText("div.ui-vip-profile-info__info-link")
		}
	})

	innerCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	mainCollector.Visit("https://inmuebles.mercadolibre.com.ar/casas/venta/bsas-gba-sur/la-plata")

	result, error := json.MarshalIndent(houses, "", "\t")
	
	if error != nil {
		fmt.Println(error)
	}

	fmt.Println(string(result))
}
