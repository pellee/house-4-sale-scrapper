package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	sf "strings"

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
	Id                 string        `json:"id"` // TODO: buscar la manera de convertir el link de la propiedad en un hash.
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
			Id:            hex.EncodeToString(hash.Sum(nil)),
			Location_info: GetLocationInfo(e.ChildText("span.poly-component__location")),
			Price_info: price{
				sf.Replace(e.ChildText("span.andes-money-amount__fraction"), ".", "", 1),
				e.ChildText("span.andes-money-a375mount__currency-symbol"),
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

	inner_collector.OnHTML("div#ui-pdp-main-container", func(e *colly.HTMLElement) {
		link := e.Request.URL.String()

		hash := md5.New()
		hash.Write([]byte(link))

		i := slices.IndexFunc(houses, func(h house) bool {
			return h.Id == hex.EncodeToString(hash.Sum(nil))
		})

		if i != -1 {
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
