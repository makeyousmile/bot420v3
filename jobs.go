package main

var cityNames []string = []string{
	"Барановичи",
	"Бобруйск",
	"Борисов",
	"Браслав",
	"Брест",
	"Витебск",
	"Гомель",
	"Городок",
	"Гродно",
	"Дзержинск",
	"Жлобин",
	"Жодино",
	"Заславль",
	"Кобрин",
	"Лепель",
	"Лида",
	"Минск",
	"Могилев",
	"Мозырь",
	"Молодечно",
	"Новополоцк",
	"Орша",
	"Осиповичи",
	"Островец",
	"Отправка по Беларуси",
	"Пинск",
	"Полоцк",
	"Речица",
	"Светлогорск",
	"Слоним",
	"Слуцк",
	"Смолевичи",
	"Солигорск",
}
var cityValues []string = []string{
	"538",
	"508",
	"518",
	"688",
	"432",
	"138",
	"431",
	"892",
	"145",
	"1079",
	"915",
	"140",
	"687",
	"1372",
	"1325",
	"999",
	"144",
	"410",
	"809",
	"519",
	"626",
	"430",
	"1221",
	"1248",
	"812",
	"593",
	"625",
	"738",
	"835",
	"964",
	"437",
	"1349",
	"433",
}
var catNames []string = []string{
	"Марихуана",
	"Cтимуляторы",
	"Эйфоретики",
	"Психоделики",
	"Энтеогены",
	"Экстази",
	"Диссоциативы",
	"Опиаты",
	"Химические	реактивы/Конструкторы",
	"Аптека",
	"Обнал BTC",
	"SSH, VPN",
	"Цифровые товары",
	"Документы",
	"Карты, SIM",
	"Дизайн и графика",
	"Наружная реклама",
	"Фальшивые деньги",
	"Приборы и оборудование",
	"Анаболики/Стероиды",
	"Партнёрство и Франшиза",
	"Работа",
	"Другое",
	"Каннабиноиды",
}
var catValues []string = []string{
	"3",
	"1",
	"2",
	"6",
	"67",
	"27",
	"5",
	"26",
	"7",
	"28",
	"89",
	"63",
	"29",
	"30",
	"31",
	"54",
	"66",
	"61",
	"64",
	"65",
	"68",
	"52",
	"8",
	"4",
}

type Links struct {
	i          int
	CityNames  []string
	CityValues []string
	CatNames   []string
	CatValues  []string
	Jobs       []string
	values     []Values
}
type Values struct {
	cityValue string
	cityNames string
	catValue  string
	catNames  string
}

func NewLinks() *Links {
	links := new(Links)
	links.CityValues = cityValues
	links.CityNames = cityNames
	links.CatValues = catValues
	links.CatNames = catNames
	links.i = 0

	lenOfValues := len(catValues) * len(cityValues)
	jobs := make([]string, lenOfValues)
	values := make([]Values, lenOfValues)
	for i := 0; i < len(catValues); i++ {
		for j := 0; j < len(cityValues); j++ {
			cat := catValues[i]
			city := cityValues[j]
			job := hydraProxy + "catalog/" + cat + "?query=&region_id=" + city + "&subregion_id=0&price%5Bmin%5D=&price%5Bmax%5D=&unit=g&weight%5Bmin%5D=&weight%5Bmax%5D=&type=momental"
			jobs[links.i] = job
			values[links.i].catValue = catValues[i]
			values[links.i].catNames = catNames[i]
			values[links.i].cityValue = cityValues[j]
			values[links.i].cityNames = cityNames[j]

			links.i++
			//log.Print(job)
		}
	}
	links.Jobs = jobs
	links.values = values

	return links
}

func (s *Links) getJob() string {
	if s.i == len(s.Jobs) {
		s.i = 0
	}
	link := s.Jobs[s.i]
	s.i++
	return link
}

func (s Links) getJobs() []string {

	return s.Jobs
}

func (s Links) getValues() (string, string) {
	cityValue := cityValues[s.i]
	catValue := catValues[s.i]
	return cityValue, catValue
}

func (s Links) getNames() (string, string) {

	return s.values[s.i-1].cityNames, s.values[s.i-1].catNames
}
