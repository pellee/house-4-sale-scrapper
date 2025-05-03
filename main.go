package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
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

	locationCssSelector := "span.poly-component__location"
	moneyCssSelector := "span.andes-money-amount__fraction"
	currencyCssSelector := "span.andes-money-amount__currency-symbol"
	squareMetersCoveredCssSelector := "div.poly-component__attributes-list > ul > li:nth-child(3)"
	squareMetersTotalCssSelector := "div.ui-pdp-highlighted-specs-res__icon-label:nth-child(1) > span:nth-child(2)"
	publicationDateCssSelector := "div.ui-pdp-seller-validated > p.ui-pdp-color--GRAY.ui-pdp-size--XSMALL.ui-pdp-family--REGULAR.ui-pdp-seller-validated__title"
	publicationDateAlternativeCssSelector := "p.ui-pdp-color--GRAY.ui-pdp-size--XSMALL.ui-pdp-family--REGULAR.ui-pdp-header__bottom-subtitle"
	bedroomsCssSelector := "div.ui-vpp-striped-specs:nth-child(1) > div:nth-child(1) > table:nth-child(2) > tbody > tr:nth-child(3) > td"
	bathroomsCssSelector := "div.ui-vpp-striped-specs:nth-child(1) > div:nth-child(1) > table:nth-child(2) > tbody > tr:nth-child(4) > td"
	realStateNameCssSelector := "div.ui-vip-profile-info__info-link"

	pageUrl := "https://inmuebles.mercadolibre.com.ar/casas/venta/apto-credito/bsas-gba-sur/la-plata/"
	priceRangeUrl := "_PriceRange_{1}USD-{2}USD"
	nthElement := 1
	secondPartPageUrl := "_Desde_{n}"

	builder := sf.Builder{}

	var minRangePrice, maxRangePrice int32

	fmt.Println("Enter the price ranges for the search: ")
	_, err := fmt.Scanln(&minRangePrice, &maxRangePrice)
	if err != nil {
		log.Fatal(err)
	}

	if maxRangePrice < minRangePrice {
		auxRangePrice := minRangePrice
		minRangePrice = maxRangePrice
		maxRangePrice = auxRangePrice
	}

	priceRangeUrl = sf.Replace(priceRangeUrl, "{1}", strconv.Itoa(int(minRangePrice)), 1)
	priceRangeUrl = sf.Replace(priceRangeUrl, "{2}", strconv.Itoa(int(maxRangePrice)), 1)


	builder.WriteString(pageUrl)
	builder.WriteString(priceRangeUrl)

	mainCollector := colly.NewCollector(colly.AllowedDomains(domains[0], domains[1]), colly.Async(true))
	mainCollector.Limit(&colly.LimitRule{RandomDelay: 5 * time.Second, Parallelism: 3})
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
			LocationInfo:    e.ChildText(locationCssSelector),
			PriceInfo: price{
				sf.Replace(e.ChildText(moneyCssSelector), ".", "", 1),
				e.ChildText(currencyCssSelector),
			},
			SquareMetersInfo: squareMeters{sf.Split(e.ChildText(squareMetersCoveredCssSelector), " ")[0], "unknown"},
			RealStateInfo:    realState{"unknown", "unknown"},
			Link:             link,
			Bedrooms:         "unknown",
			Bathrooms:        "unknown",
		})

		innerCollector.Visit(link)
	})

	mainCollector.OnHTML("li.andes-pagination__button.andes-pagination__button--next", func(h *colly.HTMLElement) {
		if !h.DOM.HasClass("andes-pagination__button--disabled") {
			buider := sf.Builder{}
			nthElement += 48
			nextPageUrl := sf.Replace(secondPartPageUrl, "{n}", strconv.Itoa(nthElement), 1)

			buider.WriteString(pageUrl)
			buider.WriteString(nextPageUrl)
			buider.WriteString(priceRangeUrl)
			mainCollector.Visit(buider.String())
		}
	})

	innerCollector.OnHTML("div#ui-pdp-main-container", func(e *colly.HTMLElement) {
		i := slices.IndexFunc(houses, func(h house) bool {
			return h.Id == GetId(e.Request.URL.String())
		})

		if i != -1 {
			publicationDate := e.ChildText(publicationDateCssSelector)

			if sf.Contains(publicationDate, "identidad verificada") {
				publicationDate = e.ChildText(publicationDateAlternativeCssSelector)
			}

			bedrooms := e.ChildText(bedroomsCssSelector)
			bathrooms := e.ChildText(bathroomsCssSelector)

			if len(sf.TrimSpace(bedrooms)) != 0 {
				houses[i].Bedrooms = bedrooms
			}

			if len(sf.TrimSpace(bathrooms)) != 0 {
				houses[i].Bathrooms = bathrooms
			}

			houses[i].PublicationDate = publicationDate
			houses[i].RealStateInfo.Name = e.ChildText(realStateNameCssSelector)
			houses[i].SquareMetersInfo.Total = sf.Split(e.ChildText(squareMetersTotalCssSelector), " ")[0]
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
	})

	mainCollector.Visit(builder.String())

	mainCollector.Wait()
	innerCollector.Wait()

	result, error := json.MarshalIndent(houses, "", "\t")

	if error != nil {
		fmt.Println(error)
	}

	fmt.Println(string(result))
	fmt.Println(len(houses))
}
