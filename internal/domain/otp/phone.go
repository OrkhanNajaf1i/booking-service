// File: internal/domain/otp/phone.go
package otp

import (
	"fmt"
	"strings"
)

// NormalizePhone – istifadecinin yazdigini E.164 formasina salir.
//
// Adam nomresini istediyi kimi yazir: "050 111 22 33", "+994 50 111
// 22 33", "(050) 111-22-33", "994501112233". Hamisi eyni setre
// dusmelidir — eks halda hemin adam her defe yeni hesab yaradar.
//
// Hazirda yalniz Azerbaycan nomreleri qebul olunur: istifadeci kutlesi
// buradadir ve yanlis olkeye SMS gondermek pula basa gelir. Basqa olke
// lazim olanda burani genislendirmek kifayetdir.
func NormalizePhone(raw string) (string, error) {
	digits := onlyDigits(raw)

	switch {
	case digits == "":
		return "", ErrInvalidPhone

	// 994 50 111 22 33
	case len(digits) == 12 && strings.HasPrefix(digits, "994"):
		digits = digits[3:]

	// 0 50 111 22 33
	case len(digits) == 10 && strings.HasPrefix(digits, "0"):
		digits = digits[1:]

	// 50 111 22 33
	case len(digits) == 9:
		// oldugu kimi qalir

	default:
		return "", ErrInvalidPhone
	}

	if !isKnownOperator(digits[:2]) {
		return "", ErrInvalidPhone
	}

	return "+994" + digits, nil
}

// MaskPhone – jurnalda ve cavabda gosterilen forma.
//
// Tam nomre jurnala dusmemelidir; eyni zamanda istifadeci kodun hansi
// nomreye getdiyini gormelidir.
func MaskPhone(e164 string) string {
	// Gozlenilen forma: +994XXXXXXXXX (13 simvol)
	if len(e164) != 13 {
		return "***"
	}
	// +994 50 *** ** 33 — operator kodu ve son iki reqem gorunur ki,
	// adam oz nomresini taniyasin; ortasi gizlenir.
	return fmt.Sprintf("+994 %s *** ** %s", e164[4:6], e164[11:])
}

func onlyDigits(raw string) string {
	var builder strings.Builder
	builder.Grow(len(raw))

	for _, symbol := range raw {
		if symbol >= '0' && symbol <= '9' {
			builder.WriteRune(symbol)
		}
	}
	return builder.String()
}

// azerbaycan mobil operator kodlari.
var operatorCodes = map[string]struct{}{
	"50": {}, "51": {}, // Azercell
	"55": {}, "99": {}, // Bakcell
	"70": {}, "77": {}, // Nar
	"60": {}, // Naxtel / virtual
	"10": {}, // Azerfon xidmet nomreleri
}

func isKnownOperator(prefix string) bool {
	_, ok := operatorCodes[prefix]
	return ok
}
