package sql2xml

import "encoding/xml"

const (
	// USDRate is using to convert prices in USD to UAH
	USDRate = 28.1
	// ImgURLPrefix is a prefix for actual image URL
	ImgURLPrefix = "https://URL/image/cache/"
	// MySQLDSN stores DSN to connect to MySQL DB
	MySQLDSN = "username:password@protocol(address)/dbname"
)

// CategoryStruct describes structure of category XML element
type CategoryStruct struct {
	XMLName      xml.Name `xml:"category"`
	CategoryID   int      `xml:"id,attr"`
	ParentID     int      `xml:"parentId,attr,omitempty"`
	CategoryName string   `xml:",chardata"`
}

// DescriptionStruct describes structure of description XML element
type DescriptionStruct struct {
	XMLName xml.Name `xml:"description"`
	CDATA   string   `xml:",cdata"`
}

// ItemParamStruct describes structure of item's params XML elements
type ItemParamStruct struct {
	XMLName xml.Name `xml:"param"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

// ItemStruct describes structure of item XML element
type ItemStruct struct {
	XMLName         xml.Name `xml:"item"`
	ID              int      `xml:"id,attr"`
	SellingType     string   `xml:"selling_type,attr"`
	Name            string   `xml:"name"`
	CategoryID      int      `xml:"categoryId"`
	PriceUAH        float64  `xml:"priceuah"`
	Image           string   `xml:"image"`
	Vendor          string   `xml:"vendor"`
	ItemParamsArray []ItemParamStruct
	VendorCode      string             `xml:"vendorCode"`
	Description     *DescriptionStruct `xml:"description"`
	Available       string             `xml:"available"`
	Keywords        string             `xml:"keywords"`
}

// XMLStruct describes top level XML element
type XMLStruct struct {
	XMLName       xml.Name `xml:"catalog"`
	CategoryArray []CategoryStruct
	ItemArray     []ItemStruct
}
