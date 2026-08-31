// File: internal/domain/business/geo.go
//
// "Yaxinliqdakilar" ucun mesafe hesabi.
//
// PostGIS qosulmayib — bir marketplace-in bu merhelesinde lazim deyil:
// bron edile bilen biznes sayi azdir ve suzgec yaddasda islenir. Sayi
// arta bilen yere catanda burani PostGIS-in `ST_DWithin`-i ile evez
// etmek olar; API muqavilesi deyismir.
package business

import "math"

// earthRadiusKm – orta Yer radiusu.
const earthRadiusKm = 6371.0

// NearestDistanceKm – istifadecinin noqtesinden biznesin EN YAXIN
// filialina qeder mesafe.
//
// Biznesin bir nece filiali ola bilir; musteri ucun onemli olan hansisa
// birinin yaxin olmasidir. Koordinati olan filial yoxdursa ikinci deyer
// false qayidir — bele biznes mesafe suzgecinden kecmemelidir, cunki
// harada oldugu bilinmir.
func (b *BookableBusiness) NearestDistanceKm(lat, lng float64) (float64, bool) {
	nearest := math.Inf(1)
	found := false

	for _, location := range b.Locations {
		if !location.HasCoordinates() {
			continue
		}
		distance := HaversineKm(lat, lng, *location.Latitude, *location.Longitude)
		if distance < nearest {
			nearest = distance
			found = true
		}
	}

	if !found {
		return 0, false
	}
	return nearest, true
}

// HaversineKm – iki koordinat arasindaki boyuk cevre mesafesi (km).
func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)

	return earthRadiusKm * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
