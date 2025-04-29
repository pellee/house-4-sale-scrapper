package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	sf "strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

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
	ScrapeDate       string       `json:"scrapeDate"`
	PageScraped      string       `json:"pageScraped"`
	PublicationDate  string       `json:"publicationDate"`
	LocationInfo     string       `json:"locationInfo"`
	PriceInfo        price        `json:"priceInfo"`
	SquareMetersInfo squareMeters `json:"squareMetersInfo"`
	RealStateInfo    realState    `json:"realStateInfo"`
	CreditSuitable   bool         `json:"creditalSuitable"`
	Link             string       `json:"link"`
	PropertyType     string       `json:"propertyType"`
	Bedrooms         string       `json:"bedrooms"`
	Bathrooms        string       `json:"bathrooms"`
}

func GetId(link string) string {
	hash := md5.New()
	hash.Write([]byte(link))

	return hex.EncodeToString(hash.Sum(nil))
}

func main() {
	houses := make([]house, 0)

	domains := [2]string{"inmuebles.mercadolibre.com.ar", "casa.mercadolibre.com.ar"}

	mainCollector := colly.NewCollector(colly.AllowedDomains(domains[0], domains[1]), colly.Async(true))
	mainCollector.Limit(&colly.LimitRule{RandomDelay: 5 * time.Second, Parallelism: 2})
	mainCollector.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})

	innerCollector := mainCollector.Clone()

	mainCollector.OnHTML("div.poly-card__content", func(e *colly.HTMLElement) {
		link := sf.Split(e.ChildAttr("a", "href"), "#")[0]

		houses = append(houses, house{
			Id:              GetId(link),
			PageScraped:     "Mercado Libre",
			PropertyType:    "Casa",
			ScrapeDate:      time.Now().Format(time.RFC3339),
			PublicationDate: "unknown",
			LocationInfo:    e.ChildText("span.poly-component__location"),
			PriceInfo: price{
				sf.Replace(e.ChildText("span.andes-money-amount__fraction"), ".", "", 1),
				e.ChildText("span.andes-money-amount__currency-symbol"),
			},
			SquareMetersInfo: squareMeters{sf.Split(e.ChildText("div.poly-component__attributes-list > ul > li:nth-child(3)"), " ")[0], "unknown"},
			RealStateInfo:    realState{"unknown", "unknown"},
			CreditSuitable:   false,
			Link:             link,
			Bedrooms:         "unknown",
			Bathrooms:        "unknown",
		})

		innerCollector.Visit(link)
	})

	mainCollector.OnHTML("[title=Siguiente]", func(h *colly.HTMLElement) {
		mainCollector.Visit(h.Attr("href"))
	})

	innerCollector.OnHTML("div#ui-pdp-main-container", func(e *colly.HTMLElement) {
		i := slices.IndexFunc(houses, func(h house) bool {
			return h.Id == GetId(e.Request.URL.String())
		})

		if i != -1 {
			publicationDate := e.ChildText("div.ui-pdp-seller-validated > p.ui-pdp-color--GRAY.ui-pdp-size--XSMALL.ui-pdp-family--REGULAR.ui-pdp-seller-validated__title")

			if sf.Contains(publicationDate, "identidad verificada") {
				publicationDate = e.ChildText("p.ui-pdp-color--GRAY.ui-pdp-size--XSMALL.ui-pdp-family--REGULAR.ui-pdp-header__bottom-subtitle")
			}

			bedrooms := e.ChildText("div.ui-vpp-striped-specs:nth-child(1) > div:nth-child(1) > table:nth-child(2) > tbody > tr:nth-child(3) > td")
			bathrooms := e.ChildText("div.ui-vpp-striped-specs:nth-child(1) > div:nth-child(1) > table:nth-child(2) > tbody > tr:nth-child(4) > td")

			if len(sf.TrimSpace(bedrooms)) != 0 {
				houses[i].Bedrooms = bedrooms
			}

			if len(sf.TrimSpace(bathrooms)) != 0 {
				houses[i].Bathrooms = bathrooms
			}

			houses[i].PublicationDate = publicationDate
			houses[i].RealStateInfo.Name = e.ChildText("div.ui-vip-profile-info__info-link")
			houses[i].SquareMetersInfo.Total = sf.Split(e.ChildText("div.ui-pdp-highlighted-specs-res__icon-label:nth-child(1) > span:nth-child(2)"), " ")[0]
		}
	})

	mainCollector.OnRequest(func(r *colly.Request) {
		extensions.RandomUserAgent(mainCollector)
		extensions.Referer(mainCollector)
		fmt.Println("Visiting", r.URL)
	})

	innerCollector.OnRequest(func(r *colly.Request) {
		extensions.RandomUserAgent(innerCollector)
		extensions.Referer(innerCollector)
		fmt.Println("Visiting", r.URL)
	})

	mainCollector.Visit("https://inmuebles.mercadolibre.com.ar/casas/venta/apto-credito/bsas-gba-sur/la-plata")

	mainCollector.Wait()
	innerCollector.Wait()

	result, error := json.MarshalIndent(houses, "", "\t")

	if error != nil {
		fmt.Println(error)
	}

	fmt.Println(string(result))
	fmt.Println(len(result))
}
