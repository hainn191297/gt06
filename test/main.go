package main

import (
	"fmt"
)

func encodeRawTimeZone(gmt float64, region string, language string) ([2]byte, error) {
	if gmt < 0 {
		gmt = -gmt
	}
	rawGmt := int(gmt * 100)

	rawValue := rawGmt << 4

	if region == "Western" {
		rawValue |= (1 << 4)
	}

	switch language {
	case "Chinese":
		rawValue |= 1 << 1 // China
	case "English":
		rawValue |= 1 // English
	case "":
	default:
		return [2]byte{}, fmt.Errorf("invalid language: %s", language)
	}

	return [2]byte{byte(rawValue >> 8), byte(rawValue & 0xFF)}, nil
}

func decodeRawTimeZone(data [2]byte) (float64, string, string, error) {
	rawValue := int(data[0])<<8 | int(data[1])

	gmtRaw := (rawValue >> 4) & 0xFFF
	gmt := float64(gmtRaw) / 100.0

	isEastern := (rawValue>>4)&0x1 == 0
	region := "Eastern"
	if !isEastern {
		region = "Western"
	}

	if region == "Western" {
		gmt = -gmt
	}

	languageBits := rawValue & 0x3
	language := ""
	switch languageBits {
	case 0x1:
		language = "English"
	case 0x2:
		language = "Chinese"
	case 0x3:
		return 0, "", "", fmt.Errorf("invalid language bits: %02b", languageBits)
	}

	return gmt, region, language, nil
}

func main() {
	data1, err1 := encodeRawTimeZone(12.45, "Western", "Chinese")
	if err1 == nil {
		fmt.Printf("Encoded (Western, Chinese): 0x%X 0x%X\n", data1[0], data1[1])
	}

	data2, err2 := encodeRawTimeZone(8.00, "Eastern", "English")
	if err2 == nil {
		fmt.Printf("Encoded (Eastern, English): 0x%X 0x%X\n", data2[0], data2[1])
	}

	data3, err3 := encodeRawTimeZone(5.50, "Eastern", "")
	if err3 == nil {
		fmt.Printf("Encoded (Eastern, No Language): 0x%X 0x%X\n", data3[0], data3[1])
	}

	fmt.Println("\nDecoding results:")
	gmt1, region1, language1, errDec1 := decodeRawTimeZone(data1)
	if errDec1 == nil {
		fmt.Printf("Decoded (Western, Chinese): GMT: %+v, Region: %s, Language: %s\n", gmt1, region1, language1)
	}

	gmt2, region2, language2, errDec2 := decodeRawTimeZone(data2)
	if errDec2 == nil {
		fmt.Printf("Decoded (Eastern, English): GMT: %+v, Region: %s, Language: %s\n", gmt2, region2, language2)
	}

	gmt3, region3, language3, errDec3 := decodeRawTimeZone(data3)
	if errDec3 == nil {
		fmt.Printf("Decoded (Eastern, No Language): GMT: %+v, Region: %s, Language: %s\n", gmt3, region3, language3)
	}
}
