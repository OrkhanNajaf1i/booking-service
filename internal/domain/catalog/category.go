// File: internal/domain/catalog/category.go
//
// Kesf ekraninin kateqoriya taksonomiyasi.
//
// Problem: `businesses.service_category` serbest metndir — sahib ozu
// yazir. "Dis Hekimi", "dis hekimi", "Stomatoloq" ucu de eyni sahedir,
// amma metn kimi ferqlidir. Kateqoriyani birbasa hemin metnden yigsaq,
// biznes sayi artdiqca siyahi eyni seyin onlarla variantina bolunecek.
//
// Ona gore serbest metn burada sabit kateqoriyaya cevrilir. Sahibin
// yazdigi metn saxlanilir (kartda alt basliq kimi gorunur), amma
// qruplasdirma yalniz slug uzre gedir.
package catalog

import "strings"

// Category – kesf ekranindaki sabit sahe.
type Category struct {
	// Slug API-de ve filtrde islenen deyismez acardir.
	Slug string `json:"slug"`
	// Name musteriye gorunen ad.
	Name string `json:"name"`
	// Icon mobil/veb terefdeki ikon acaridir — her iki tetbiq eyni
	// adlandirmani islesin deye serverden gelir.
	Icon string `json:"icon"`

	// keywords – serbest metni bu kateqoriyaya baglayan sozler.
	// ASCII transliterasiya ile yazilir (bax: normalize).
	keywords []string
}

// SlugOther – tanilmayan sahe.
const SlugOther = "other"

// categories – tanima siralamasi. Sira ONEMLIDIR: daha xususi sahe
// evvel gelmelidir. "Dis klinikasi" hem "dis", hem "klinika" sozunu
// dasiyir; dogru cavab dis hekimidir, ona gore dentist yuxaridadir.
var categories = []Category{
	{
		Slug: "dentist", Name: "Diş həkimi", Icon: "dentist",
		keywords: []string{"dis", "stomatoloq", "stomatologiya", "dentist", "ortodont", "implant"},
	},
	{
		Slug: "hospital", Name: "Xəstəxana və klinika", Icon: "hospital",
		keywords: []string{"xestexana", "hospital", "klinika", "clinic", "tibb merkezi", "medical center"},
	},
	{
		Slug: "vet", Name: "Baytar", Icon: "vet",
		keywords: []string{"baytar", "veterinar", "vet", "heyvan"},
	},
	{
		Slug: "doctor", Name: "Həkim", Icon: "doctor",
		keywords: []string{
			"hekim", "doctor", "terapevt", "kardioloq", "pediatr", "nevroloq",
			"ginekoloq", "uroloq", "dermatoloq", "endokrinoloq", "oftalmoloq",
			"lor", "psixoloq", "psixiatr", "cerrah", "travmatoloq", "healthcare",
		},
	},
	{
		Slug: "lab", Name: "Laboratoriya", Icon: "lab",
		keywords: []string{"laboratoriya", "analiz", "lab"},
	},
	{
		Slug: "barber", Name: "Bərbər", Icon: "barber",
		keywords: []string{"berber", "barber", "sac", "kisi salonu", "hairdresser"},
	},
	{
		Slug: "beauty", Name: "Gözəllik salonu", Icon: "beauty",
		keywords: []string{
			"gozellik", "beauty", "salon", "kosmetoloq", "kosmetologiya",
			"manikur", "pedikur", "makiyaj", "kirpik", "qas", "epilyasiya",
		},
	},
	{
		Slug: "spa", Name: "SPA və masaj", Icon: "spa",
		keywords: []string{"spa", "masaj", "massage", "hamam"},
	},
	{
		Slug: "fitness", Name: "İdman və fitnes", Icon: "fitness",
		keywords: []string{"fitnes", "fitness", "idman", "gym", "zal", "yoga", "pilates", "treyner", "mesqci"},
	},
	{
		Slug: "education", Name: "Təhsil və repetitor", Icon: "education",
		keywords: []string{"repetitor", "muellim", "kurs", "tehsil", "dersler", "hazirliq", "education"},
	},
	{
		Slug: "photo", Name: "Foto və video", Icon: "photo",
		keywords: []string{"foto", "photo", "video", "studiya"},
	},
	{
		Slug: "nutrition", Name: "Dietoloq və qidalanma", Icon: "nutrition",
		keywords: []string{"dietoloq", "qidalanma", "nutrition", "pehriz"},
	},
	{
		Slug: "physio", Name: "Fizioterapiya və reabilitasiya", Icon: "physio",
		keywords: []string{"fizioterapiya", "reabilitasiya", "physio", "manual terapiya"},
	},
	{
		Slug: "tattoo", Name: "Tatu və pirsinq", Icon: "tattoo",
		keywords: []string{"tatu", "tattoo", "pirsinq", "piercing"},
	},
	{
		Slug: "auto", Name: "Avtoservis", Icon: "auto",
		keywords: []string{"avtoservis", "avtoyuma", "sinalanma", "diaqnostika", "car service"},
	},
	{
		Slug: "master", Name: "Usta və təmir", Icon: "master",
		keywords: []string{
			"usta", "temir", "repair", "santexnik", "elektrik", "kondisioner",
			"mebel", "qaynaq", "kombi",
		},
	},
	{
		Slug: "cleaning", Name: "Təmizlik xidməti", Icon: "cleaning",
		keywords: []string{"temizlik", "cleaning", "yigisdirma", "kimyevi temizleme"},
	},
	{
		Slug: "legal", Name: "Hüquq və konsultasiya", Icon: "legal",
		keywords: []string{"huquq", "vekil", "notarius", "legal", "konsultasiya", "mesletci"},
	},
	{
		Slug: "event", Name: "Tədbir və şənlik", Icon: "event",
		keywords: []string{"tedbir", "senlik", "toy", "event", "aparici", "dj"},
	},
	{
		Slug: SlugOther, Name: "Digər", Icon: "other",
		keywords: nil, // tanilmayanlarin yigildigi yer
	},
}

// bySlug – tez axtaris ucun indeks.
var bySlug = func() map[string]Category {
	index := make(map[string]Category, len(categories))
	for _, item := range categories {
		index[item.Slug] = item
	}
	return index
}()

// All – butun kateqoriyalar tanima sirasi ile.
func All() []Category {
	out := make([]Category, len(categories))
	copy(out, categories)
	return out
}

// Selectable – sahibin secebileceyi kateqoriyalar.
//
// "Diger" burada YOXDUR. O, pese deyil — yalniz kohne, kateqoriyasi
// secilmemis setirlerin dusdugu yerdir. Secim siyahisinda gorunse
// adamlar orani secib gedecek ve kesf ekrani yene qruplasmayacaq.
func Selectable() []Category {
	out := make([]Category, 0, len(categories)-1)
	for _, item := range categories {
		if item.Slug == SlugOther {
			continue
		}
		out = append(out, item)
	}
	return out
}

// IsSelectable – slug sahibin secebileceyi kateqoriyadirmi.
func IsSelectable(slug string) bool {
	category, ok := BySlug(slug)
	return ok && category.Slug != SlugOther
}

// BySlug – slug uzre kateqoriya. Tapilmasa ikinci deyer false olur.
func BySlug(slug string) (Category, bool) {
	item, ok := bySlug[strings.ToLower(strings.TrimSpace(slug))]
	return item, ok
}

// ResolveWith – sahibin acik secimini serbest metnin tehmininden
// ustun tutur.
//
// slug verilibse ve taninirsa hemin kateqoriya qaytarilir: sahib
// "Hekim" secibse, "Dis-cene cerrahi" yazsa da hekim olaraq qalir.
// slug bos olan kohne setirler ucun acar soz tanimasina qayidilir.
func ResolveWith(slug, serviceCategory, industry string) Category {
	if category, ok := BySlug(slug); ok {
		return category
	}
	return Resolve(serviceCategory, industry)
}

// Resolve – biznesin serbest metnini sabit kateqoriyaya baglayir.
//
// service_category daha deqiqdir ("Dis Hekimi"), industry daha genisdir
// ("healthcare") — ona gore evvelce birincisine baxilir. Hec biri
// uygun gelmese "Diger" qaytarilir: biznes kesf ekranindan itmemelidir.
func Resolve(serviceCategory, industry string) Category {
	for _, raw := range []string{serviceCategory, industry} {
		text := normalize(raw)
		if text == "" {
			continue
		}
		for _, category := range categories {
			for _, keyword := range category.keywords {
				if strings.Contains(text, keyword) {
					return category
				}
			}
		}
	}
	return bySlug[SlugOther]
}

// normalize – muqayise ucun metni sadelesdirir: kicik herf + Azerbaycan
// herflerinin ASCII qarsiligi. Belelikle "Diş" ile "Dis" eyni sayilir.
func normalize(text string) string {
	var builder strings.Builder
	builder.Grow(len(text))

	for _, symbol := range strings.ToLower(strings.TrimSpace(text)) {
		switch symbol {
		case 'ə':
			builder.WriteByte('e')
		case 'ı':
			builder.WriteByte('i')
		case 'ş':
			builder.WriteByte('s')
		case 'ç':
			builder.WriteByte('c')
		case 'ğ':
			builder.WriteByte('g')
		case 'ö':
			builder.WriteByte('o')
		case 'ü':
			builder.WriteByte('u')
		default:
			builder.WriteRune(symbol)
		}
	}
	return builder.String()
}
