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

type square_meters struct {
	Covered string `json:"covered"`
	Total   string `json:"total"`
}

type real_state struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type house struct {
	Id                 string        `json:"id"`
	Scrape_date        string        `json:"scrape_date"`
	Publication_date   string        `json:"publication_date"`
	Location_info      location      `json:"location_info"`
	Price_info         price         `json:"price_info"`
	Square_meters_info square_meters `json:"square_meters_info"`
	Real_state_info    real_state    `json:"real_state_info"`
	Credit_suitable    bool          `json:"credital_suitable"`
	Link               string        `json:"link"`
}

func GetLocationInfo(raw_location string) location {
	splitted_location := sf.Split(raw_location, ",")

	if len(splitted_location) == 5 {
		return location{
			Zone:     sf.TrimSpace(splitted_location[4]),
			Locality: sf.TrimSpace(splitted_location[3]),
			Hood:     sf.TrimSpace(splitted_location[1]) + sf.TrimSpace(splitted_location[2]),
			Address:  sf.TrimSpace(splitted_location[0]),
		}
	} else if len(splitted_location) == 3 {
		return location{
			Zone:     sf.TrimSpace(splitted_location[2]),
			Locality: sf.TrimSpace(splitted_location[1]),
			Hood:     sf.TrimSpace("unknown"),
			Address:  sf.TrimSpace(splitted_location[0]),
		}
	}

	return location{
		Zone:     sf.TrimSpace(splitted_location[3]),
		Locality: sf.TrimSpace(splitted_location[2]),
		Hood:     sf.TrimSpace(splitted_location[1]),
		Address:  sf.TrimSpace(splitted_location[0]),
	}
}

func main() {
	houses := make([]house, 0)
	domains := [2]string{"inmuebles.mercadolibre.com.ar", "casa.mercadolibre.com.ar"}
	main_collector := colly.NewCollector(colly.AllowedDomains(domains[0], domains[1]))
	inner_collector := main_collector.Clone()

	// Find and visit all links
	main_collector.OnHTML("div.poly-card__content", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a", "href")

		hash := md5.New()
		hash.Write([]byte(link))

		houses = append(houses, house{
			Id:               hex.EncodeToString(hash.Sum(nil)),
			Scrape_date:      time.Now().Format(time.RFC3339),
			Publication_date: "unknown",
			Location_info:    GetLocationInfo(e.ChildText("span.poly-component__location")),
			Price_info: price{
				sf.Replace(e.ChildText("span.andes-money-amount__fraction"), ".", "", 1),
				e.ChildText("span.andes-money-amount__currency-symbol"),
			},
			Square_meters_info: square_meters{"unknown", "unknown"},
			Real_state_info:    real_state{"unknown", "unknown"},
			Credit_suitable:    false,
			Link:               link,
		})

		inner_collector.Visit(link)
	})

	main_collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	// TODO: ver como puedo acceder a los valores de la tabla. Pareciera ser que, de manera aleatoria,a veces puedo acceder a los valores y a veces no.
	// inner_collector.OnHTML("div.ui-pdp-container__col.col-1.ui-vip-core-container--content-left", func(e *colly.HTMLElement) {
	// 	link := e.Request.URL.String()
	// 	value := 0
	//
	// 	hash := md5.New()
	// 	hash.Write([]byte(link))
	//
	// 	i := slices.IndexFunc(houses, func(h house) bool {
	// 		return h.Id == hex.EncodeToString(hash.Sum(nil))
	// 	})
	//
	// 	if i != -1 {
	// 		e.ForEach("table>tbody>tr", func(_ int, e *colly.HTMLElement) {
	// 			if value == 0 {
	// 				fmt.Println("entro al 1.1 " + e.ChildText("td"))
	// 				houses[i].Square_meters_info.Total = e.ChildText("td")
	// 				if houses[i].Square_meters_info.Total == "unknown" {
	// 					fmt.Println("entro al 1.2 " + e.ChildText("td>span"))
	// 					houses[i].Square_meters_info.Total = e.ChildText("td>span")
	// 				}
	// 			}
	// 			if value == 1 {
	// 				fmt.Println(e.ChildText("td"))
	// 				fmt.Println("entro al 2.1 " + e.ChildText("td"))
	// 				houses[i].Square_meters_info.Covered = e.ChildText("td")
	// 				if houses[i].Square_meters_info.Covered == "unknown" {
	// 					fmt.Println("entro al 2.2 " + e.ChildText("td>span"))
	// 					houses[i].Square_meters_info.Covered = e.ChildText("td>span")
	// 				}
	// 			}
	// 			value++
	// 		})
	// 	}
	// })

	inner_collector.OnHTML("div#ui-pdp-main-container", func(e *colly.HTMLElement) {
		link := e.Request.URL.String()

		hash := md5.New()
		hash.Write([]byte(link))

		i := slices.IndexFunc(houses, func(h house) bool {
			return h.Id == hex.EncodeToString(hash.Sum(nil))
		})

		if i != -1 {
			publication_date := e.ChildText("div.ui-pdp-seller-validated > p.ui-pdp-color--GRAY.ui-pdp-size--XSMALL.ui-pdp-family--REGULAR.ui-pdp-seller-validated__title")

			if sf.Contains(publication_date, "identidad verificada") {
				publication_date = e.ChildText("p.ui-pdp-color--GRAY.ui-pdp-size--XSMALL.ui-pdp-family--REGULAR.ui-pdp-header__bottom-subtitle")
			}

			houses[i].Publication_date = publication_date
			houses[i].Real_state_info.Name = e.ChildText("div.ui-vip-profile-info__info-link")
		}
	})

	inner_collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	main_collector.Visit("https://inmuebles.mercadolibre.com.ar/casas/venta/bsas-gba-sur/la-plata")

	result, error := json.MarshalIndent(houses, "", "\t")

	if error != nil {
		fmt.Println(error)
	}

	fmt.Println(string(result))
}
