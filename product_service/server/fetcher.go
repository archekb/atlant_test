package main

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"net/http"
	nURL "net/url"
	"strconv"
	"strings"
)

// PriceCoderInterface - encode and decode price into int64, it's more safety for save into db
type PriceCoderInterface interface {
	Encode(string) (int64, error)
}

// ProductSaverInterface - save product to anyware
type ProductSaverInterface interface {
	Save(string, int64) error
}

// Fetcher - contains parametrs of csv parser, price coder and data saver
type Fetcher struct {
	Comma           rune
	Comment         rune
	FieldsPerRecord int
	PriceCoder      PriceCoderInterface
	ProductSaver    ProductSaverInterface
}

// NewFetcher - make new Fetcher with current params:
// coma - csv separator;
// comment - prefix of csv comment;
// fieldsPerRecord - count of fields per line;
// priceCoder - interface for convert price to db format;
// productSaver - interface for saving product
func NewFetcher(comma, comment rune, fieldsPerRecord int, priceCoder PriceCoderInterface, productSaver ProductSaverInterface) *Fetcher {
	return &Fetcher{Comma: comma, Comment: comment, FieldsPerRecord: fieldsPerRecord, PriceCoder: priceCoder, ProductSaver: productSaver}
}

// Fetch - open and read url
func (f *Fetcher) Fetch(url string) error {
	if !f.IsURL(url) {
		return errors.New("URL is not valid")
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Incorrect answer code " + strconv.Itoa(resp.StatusCode))
	}

	reader := csv.NewReader(resp.Body)
	reader.FieldsPerRecord = f.FieldsPerRecord
	reader.Comment = f.Comment
	reader.Comma = f.Comma

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if errors.Is(err, csv.ErrFieldCount) {
			log.Println("Field count more than 2, ignore", line)
			continue
		}

		if err != nil {
			return err
		}

		product, price, err := f.lineParcer(line)
		if err != nil {
			log.Println("Can't parse line", line, "becuse error", err)
			continue
		}

		err = f.ProductSaver.Save(product, price)
		if err != nil {
			log.Println("Can't save product", product, price, "becuse error", err)
			continue
		}
	}

	return nil
}

// IsURL - check valid url
func (f *Fetcher) IsURL(url string) bool {
	u, err := nURL.Parse(url)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (f *Fetcher) lineParcer(line []string) (string, int64, error) {
	if len(line) < f.FieldsPerRecord {
		return "", 0, errors.New("Some thing went wrong")
	}

	price, err := f.PriceCoder.Encode(strings.TrimSpace(line[1]))
	if err != nil {
		return "", 0, err
	}

	return strings.TrimSpace(line[0]), price, nil
}
