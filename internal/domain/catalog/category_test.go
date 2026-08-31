package catalog

import "testing"

func TestResolve(t *testing.T) {
	cases := []struct {
		name            string
		serviceCategory string
		industry        string
		wantSlug        string
	}{
		// Eyni sahenin ferqli yazilislari bir yere dusmelidir —
		// kateqoriya siyahisinin parcalanmamasi bundan asilidir.
		{"dis hekimi", "Diş Həkimi", "", "dentist"},
		{"asci herflerle", "dis hekimi", "", "dentist"},
		{"stomatoloq", "Stomatoloq", "", "dentist"},

		// "Dis klinikasi" hem dis, hem klinika sozunu dasiyir;
		// daha xususi olan qalib gelmelidir.
		{"dis klinikasi", "Diş klinikası", "", "dentist"},

		{"xestexana", "Xəstəxana", "", "hospital"},
		{"klinika", "Şəfa Klinikası", "", "hospital"},
		{"hekim", "Kardioloq", "", "doctor"},
		{"berber", "Bərbər", "", "barber"},
		{"sac", "Kişi saç ustası", "", "barber"},
		{"gozellik", "Gözəllik salonu", "", "beauty"},
		{"usta", "Santexnik usta", "", "master"},
		{"baytar", "Baytar həkimi", "", "vet"},

		// service_category bos olanda industry-e baxilir.
		{"industry ehtiyat", "", "healthcare", "doctor"},

		// Tanilmayan sahe itmemelidir.
		{"taninmayan", "Qeyri-adi xidmet", "", SlugOther},
		{"bos", "", "", SlugOther},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := Resolve(testCase.serviceCategory, testCase.industry)
			if got.Slug != testCase.wantSlug {
				t.Fatalf("Resolve(%q, %q) = %q, gozlenilen %q",
					testCase.serviceCategory, testCase.industry, got.Slug, testCase.wantSlug)
			}
			if got.Name == "" || got.Icon == "" {
				t.Fatalf("kateqoriyanin adi/ikonu bos: %+v", got)
			}
		})
	}
}

func TestBySlug(t *testing.T) {
	if _, ok := BySlug("dentist"); !ok {
		t.Fatal("dentist tapilmadi")
	}
	if _, ok := BySlug("  DENTIST "); !ok {
		t.Fatal("slug bosluq/boyuk herfe qarsi davamli olmalidir")
	}
	if _, ok := BySlug("yoxdur"); ok {
		t.Fatal("olmayan slug tapildi")
	}
}

func TestAllIsCopy(t *testing.T) {
	first := All()
	first[0].Name = "deyisdirildi"
	if All()[0].Name == "deyisdirildi" {
		t.Fatal("All() daxili siyahini asir")
	}
}
