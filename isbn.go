package mela

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidISBN = errors.New("the given string does not have the right number of characters to be an ISBN")
var ErrInvalidISBN10 = errors.New("the given string has 10 digits, but is not a valid ISBN-10")
var ErrIncorrectISBN10 = errors.New("the given ISBN-10 has an incorrect digit")
var ErrInvalidISBN13 = errors.New("the given string has 13 digits, but is not a valid ISBN-13")
var ErrIncorrectISBN13 = errors.New("the given ISBN-13 has an incorrect digit")

func validateISBN(isbn10or13 string) (string, error) {
	isbn10or13 = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(isbn10or13, " ", ""), "-", ""))

	switch len(isbn10or13) {
	case 10:
		if err := validateISBN10(isbn10or13); err != nil {
			return "", err
		}

		return isbn10To13(isbn10or13), nil
	case 13:
		if err := validateISBN13(isbn10or13); err != nil {
			return "", err
		}

		return isbn10or13, nil
	default:
		return "", ErrInvalidISBN
	}
}

func validateISBN10(isbn10 string) error {
	check := isbn10CheckDigit(isbn10)
	if check == 0x0 {
		return ErrInvalidISBN10
	}

	if isbn10[9] != check {
		return ErrIncorrectISBN10
	}
	return nil
}

func isbn10CheckDigit(isbn10 string) byte {
	total := 0
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(isbn10[i]))
		if err != nil {
			return 0x0
		}

		total += digit * (10 - i)
	}

	checkInt := 11 - (total % 11)
	checkStr := strconv.Itoa(checkInt)
	switch checkStr {
	case "11":
		checkStr = "0"
	case "10":
		checkStr = "X"
	}

	return checkStr[0]
}

func isbn10To13(isbn10 string) string {
	isbn := "978" + isbn10[0:9]
	check := isbn13CheckDigit(isbn)

	return isbn + string(check)
}

func validateISBN13(isbn13 string) error {
	check := isbn13CheckDigit(isbn13)
	if check == 0x0 {
		return ErrInvalidISBN13
	}

	if isbn13[12] != check {
		return ErrIncorrectISBN13
	}
	return nil
}

func isbn13CheckDigit(isbn13 string) byte {
	total := 0
	for i := 0; i < 12; i++ {
		digit, err := strconv.Atoi(string(isbn13[i]))
		if err != nil {
			return 0x0
		}

		mult := (i%2)*2 + 1
		total += digit * mult
	}

	checkInt := 10 - (total % 10)
	checkStr := strconv.Itoa(checkInt)
	if checkStr == "10" {
		return '0'
	}

	return checkStr[0]
}
