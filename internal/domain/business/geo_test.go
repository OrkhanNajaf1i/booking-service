package business

import (
	"math"
	"testing"
)

func ptr(value float64) *float64 { return &value }

func TestHaversineKm(t *testing.T) {
	// Baki merkezi → Baki Hava Limani, teqriben 20 km.
	got := HaversineKm(40.4093, 49.8671, 40.4675, 50.0467)
	if got < 15 || got > 25 {
		t.Fatalf("mesafe = %.2f km, 15–25 km gozlenilirdi", got)
	}

	// Eyni noqte sifir vermelidir.
	if same := HaversineKm(40.4093, 49.8671, 40.4093, 49.8671); math.Abs(same) > 1e-9 {
		t.Fatalf("eyni noqte ucun mesafe %v", same)
	}
}

func TestNearestDistanceKm(t *testing.T) {
	target := &BookableBusiness{Locations: []LocationSummary{
		{Name: "Uzaq", Latitude: ptr(40.6000), Longitude: ptr(50.2000)},
		{Name: "Yaxin", Latitude: ptr(40.4100), Longitude: ptr(49.8680)},
		{Name: "Koordinatsiz"},
	}}

	distance, ok := target.NearestDistanceKm(40.4093, 49.8671)
	if !ok {
		t.Fatal("koordinatli filial var, mesafe hesablanmali idi")
	}
	if distance > 1 {
		t.Fatalf("en yaxin filial secilmeyib: %.2f km", distance)
	}

	// Koordinati olmayan biznes mesafe suzgecinden kecmemelidir.
	unknown := &BookableBusiness{Locations: []LocationSummary{{Name: "Filial"}}}
	if _, ok := unknown.NearestDistanceKm(40.4, 49.8); ok {
		t.Fatal("koordinatsiz biznes ucun mesafe qaytarildi")
	}

	// Filiali olmayan biznes de eynidir.
	empty := &BookableBusiness{}
	if _, ok := empty.NearestDistanceKm(40.4, 49.8); ok {
		t.Fatal("filialsiz biznes ucun mesafe qaytarildi")
	}
}
