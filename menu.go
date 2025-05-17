package main

import (
	"fmt"
	"log"
)

type Menu struct {
	propertyTypeSelected string
	propertyLocation     PropertyLocation
	priceRanges          PriceRange
}
type PropertyLocation struct {
	zone     string
	locality string
}
type PriceRange struct {
	min int32
	max int32
}

func MenuSelectionPropertyLocation() PropertyLocation {
	var zoneSelection, localitySelection int32

	localities := [4][4]string{{"capital-federeal", "parque-patricios", "san-nicolas", "san-cristobal"}, {"bsas-gba-sur", "la-plata", "lanus", "lomas-de-zamora"}, {"bsas-gba-oeste", "moron", "ituzaingo", "castelar"}, {"bsas-gba-norte", "san-isidro", "san-fernando", "vicente-lopez"}}

	fmt.Println("Enter the zone you like to search.\n1. CABA\n2. GBA Sur\n3. GBA Oeste\n4. GBA Norte")
	_, err := fmt.Scanln(&zoneSelection)
	if err != nil {
		log.Fatal(err)
	}

	switch zoneSelection {
	case 1:
		fmt.Println("Enter the locality you like to search.\n1. Parque Patricios\n2. San Nicolas\n3. San Cristobal")
	case 2:
		fmt.Println("Enter the locality you like to search.\n1. La Plata\n2. Lanus\n3. Lomas De Zamora")
	case 3:
		fmt.Println("Enter the locality you like to search.\n1. Moron\n2. Ituzaingo\n3. Castelar")
	case 4:
		fmt.Println("Enter the locality you like to search.\n1. San Isidro\n2. San Fernando\n3. Vicente Lopez")
	}
	_, err = fmt.Scanln(&localitySelection)
	if err != nil {
		log.Fatal(err)
	}

	zoneSelection -= 1
	return PropertyLocation{
		zone:     localities[zoneSelection][0],
		locality: localities[zoneSelection][localitySelection],
	}
}

func MenuSelectionPropertyType() string {
	var propertyTypeSelection int32
	propertyTypes := [3]string{"casas", "departamentos", "ph"}

	fmt.Println("Enter the property type you like to search.\n1. Casas\n2. Departamentos \n3. PH")
	_, err := fmt.Scanln(&propertyTypeSelection)
	if err != nil {
		log.Fatal(err)
	}
	return propertyTypes[propertyTypeSelection-1]
}

func MenuSelectionPropertyPriceRange() PriceRange {
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
	return PriceRange{
		min: minRangePrice,
		max: maxRangePrice,
	}
}

func MenuSelection() Menu {
	return Menu{
		propertyTypeSelected: MenuSelectionPropertyType(),
		priceRanges:          MenuSelectionPropertyPriceRange(),
		propertyLocation:     MenuSelectionPropertyLocation(),
	}
}
