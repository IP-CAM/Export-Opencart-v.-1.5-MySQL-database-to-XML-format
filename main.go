package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"log"
	"math"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	sql2xml "github.com/nal/opencart-sql2xml/go"
)

func main() {
	db, err := sql.Open("mysql", sql2xml.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Print("Connected to db...")

	// Get top categories
	rows, err := db.Query(`
		SELECT c.category_id, cd.name, IF(sort_order>0, sort_order, 999) weight FROM s_category c
		JOIN s_category_description cd ON c.category_id = cd.category_id
		WHERE c.status = 1 AND c.parent_id = 0 ORDER BY weight ASC`)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Fill top categories
	var xmlStruct sql2xml.XMLStruct
	var topCategory sql2xml.CategoryStruct
	var categoryWeight int

	for rows.Next() {
		err := rows.Scan(&topCategory.CategoryID, &topCategory.CategoryName, &categoryWeight)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		xmlStruct.CategoryArray = append(xmlStruct.CategoryArray, topCategory)

		// Get subcategories
		rows, err := db.Query(`
		SELECT c.category_id, cd.name, IF(sort_order>0, sort_order, 999) weight FROM s_category c
		JOIN s_category_description cd ON c.category_id = cd.category_id
		WHERE c.status = 1 AND c.parent_id = ? ORDER BY weight ASC`, topCategory.CategoryID)

		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		defer rows.Close()

		// Fill sub categories
		var subCategory sql2xml.CategoryStruct

		for rows.Next() {
			err := rows.Scan(&subCategory.CategoryID, &subCategory.CategoryName, &categoryWeight)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			xmlStruct.CategoryArray = append(xmlStruct.CategoryArray, subCategory)
		}

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Get products (items in term of output xml)
	rows, err = db.Query(`SELECT p.product_id id, pd.name, pc.category_id, p.price, p.image, man.name vendor,
	p.sku vendorCode, pd.description, IF(p.stock_status_id = 7, 'true', 'false') available,
	pd.meta_keyword keywords
	FROM s_product p
	JOIN s_product_to_category pc   ON p.product_id         = pc.product_id
	JOIN s_manufacturer man         ON p.manufacturer_id    = man.manufacturer_id
	JOIN s_product_description pd   ON p.product_id         = pd.product_id
	WHERE p.status = 1`)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Get product categories.
	// Product can be in several categories. De-dup by using the first one.
	uniqItems := make(map[int]int)

	for rows.Next() {
		var item sql2xml.ItemStruct
		var description sql2xml.DescriptionStruct

		err := rows.Scan(&item.ID, &item.Name, &item.CategoryID,
			&item.PriceUAH, &item.Image, &item.Vendor, &item.VendorCode,
			&description.CDATA, &item.Available, &item.Keywords)

		if err != nil {
			log.Fatal(err)
		}

		if uniqItems[item.ID] == 1 {
			continue
		}
		uniqItems[item.ID] = 1

		// In description.CDATA we have escaped html code.
		// It has <style>..</style> section. Lets strip it.
		description.CDATA = html.UnescapeString(description.CDATA)
		htmlStyleIndex := strings.Index(description.CDATA, "</style>")
		if htmlStyleIndex > 0 {
			htmlStyleContent := description.CDATA[0 : htmlStyleIndex+8] // 8 is len("</style>")
			description.CDATA = strings.TrimPrefix(description.CDATA, htmlStyleContent)
		}

		item.SellingType = "r"
		item.PriceUAH = math.Round(item.PriceUAH * sql2xml.USDRate)
		item.Image = sql2xml.ImgURLPrefix + item.Image
		item.Image = strings.Replace(item.Image, ".jpg", "-500x500.jpg", 1)
		item.Image = strings.Replace(item.Image, ".JPG", "-500x500.jpg", 1)
		item.Description = &description

		// Get product custom params. Params is one-to-many relationship.
		// First we group custom params in SQL query, then store in two arrays.
		itemParams, itemParamsErr := db.Query(`SELECT GROUP_CONCAT(pad.name SEPARATOR ';;;') name, GROUP_CONCAT(pa.text SEPARATOR ';;;') value
		FROM s_product p
		LEFT JOIN s_product_attribute pa     ON p.product_id         = pa.product_id
		LEFT JOIN s_attribute_description pad ON pa.attribute_id     = pad.attribute_id
		WHERE p.product_id = ? GROUP BY p.product_id`, item.ID)

		if itemParamsErr != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		defer itemParams.Close()

		for itemParams.Next() {
			var paramName, paramValue sql.NullString
			var itemParam sql2xml.ItemParamStruct

			err := itemParams.Scan(&paramName, &paramValue)

			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			if paramName.Valid {
				itemParam.Name = paramName.String
			} else {
				continue
			}

			if paramValue.Valid {
				itemParam.Value = paramValue.String
			} else {
				continue
			}

			sql2xml.GenerateItemParams(item, itemParam)
		}
		xmlStruct.ItemArray = append(xmlStruct.ItemArray, item)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Started xml.MarshalIndent")
	output, err := xml.MarshalIndent(xmlStruct, "    ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	log.Println("Finished xml.MarshalIndent")

	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)

	log.Print("Finished...")
}
