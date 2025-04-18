package text_test

import (
	"testing"

	"golang.binggl.net/monorepo/pkg/text"
)

const validEnc = "LzIwMjNfMDZfMDMvRVhBTVBMRV9ET0NVTUVOVC5QREY="
const validDec = "/2023_06_03/EXAMPLE_DOCUMENT.PDF"

const validEncUmlaute = "LzIwMjVfMDRfMTgvVVNJTkdfVU1MQVVURW/MiG/MiG/MiG/MiG/MiC5wZGY="
const validDecUmlaute = "/2025_04_18/USING_UMLAUTEööööö.pdf"
const validEncUmlauteURLEnc = "LzIwMjVfMDRfMTgvVVNJTkdfVU1MQVVURW%2FMiG%2FMiG%2FMiG%2FMiG%2FMiC5wZGY%3D"

func TestSafePathURLEncodingDecoding(t *testing.T) {
	// enc/dec with umlaute and resulting "special characters"
	decoded := text.DecBase64SafePath(validEncUmlauteURLEnc)
	if decoded != validDecUmlaute {
		t.Errorf("could not decode the safe-path encoded base64")
	}

	encoded := text.EncBase64SafePath(validDecUmlaute)
	if encoded != validEncUmlauteURLEnc {
		t.Errorf("could not encoded base64 with safe-path")
	}
}

func TestEncodingDecoding(t *testing.T) {
	decoded := text.DecBase64(validEnc)
	if decoded != validDec {
		t.Errorf("could not decode base64")
	}

	encoded := text.EncBase64(validDec)
	if encoded != validEnc {
		t.Errorf("could not encoded base64")
	}

	decoded = text.DecBase64(validEncUmlaute)
	if decoded != validDecUmlaute {
		t.Errorf("could not decode base64")
	}
}
