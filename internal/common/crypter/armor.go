package crypter

import (
	"fmt"
	"strings"
)

const armorHeader = "-----BEGIN ENCRYPTED CONTENT-----"
const armorFooter = "-----END ENCRYPTED CONTENT-----"
const armorLength = 32

// Armor takes the input string and puts it into a defined string representation
func Armor(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty input supplied")
	}

	armoredText := armorHeader
	armoredText += "\n"

	inputLength := len(input)
	if inputLength <= armorLength {
		armoredText += input
		armoredText += "\n"
		armoredText += armorFooter

		return armoredText, nil
	}

	// divide the input by the armorLength
	// and put each part into a separate line
	lines := inputLength / armorLength
	if inputLength%armorLength != 0 {
		lines += 1
	}

	index := 0
	inputLine := ""
	remainder := ""
	for i := range int(lines) {
		index = i * armorLength
		remainder = input[index:]
		if len(remainder) < armorLength {
			inputLine = input[index:]
		} else {
			inputLine = input[index : index+armorLength]
		}
		armoredText += inputLine
		armoredText += "\n"
	}

	armoredText += armorFooter
	return armoredText, nil
}

// DeArmor takes the armor string representation and removes the header/footer
func DeArmor(armorInput string) (string, error) {
	if armorInput == "" {
		return "", fmt.Errorf("empty armorInput supplied")
	}
	if !strings.Contains(armorInput, armorHeader) || !strings.Contains(armorInput, armorFooter) {
		return "", fmt.Errorf("provided armorInput is invalid")
	}

	armorInput = strings.ReplaceAll(armorInput, "\n", "")
	armorInput = strings.ReplaceAll(armorInput, armorHeader, "")
	armorInput = strings.ReplaceAll(armorInput, armorFooter, "")

	return armorInput, nil
}
