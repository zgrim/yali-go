package yali

import (
	"fmt"
	"testing"
)

type Test struct {
	expect string
	buf    string
}

var Tests = []Test{

	{
		expect: "eng",
		buf:    `Hurricane Ismael was a weak Pacific hurricane that killed over one hundred people in northern Mexico in September of the 1995 Pacific hurricane season. It developed from a persistent area of deep convection on September 12, and steadily strengthened as it moved to the north-northwest. Ismael attained hurricane status on September 14 while located 210 miles (340 km) off the coast of Mexico.`,
	},

	{
		expect: "deu",
		buf:    `Über den Verfasser des Evangeliums, dessen die katholische und protestantische Kirche am 25. April gedenkt, ist wenig bekannt. Man geht davon aus, dass er sein Werk ungefähr in der Zeit des Großen Jüdischen Krieges, also um 70 n. Chr. geschrieben hat. Das Markusevangelium stellt dabei das erste Werk einer neuen literarischen Gattung, der Evangelien, dar, die später von anderen Autoren nachgeahmt wurde.`,
	},

	{
		expect: "fra",
		buf:    `L’étude de l’équilibre général a été reprise par Kenneth Arrow et Gérard Debreu qui établiront de façon rigoureuse les conditions d’existence et de stabilité de cet équilibre, parmi lesquelles`,
	},

	{
		expect: "spa",
		buf:    `Miguel Ángel (llamado Michelangelo), fue uno de los artistas más reconocidos por sus esculturas, pinturas y arquitectura. Realizó su labor artística durante más de setenta años entre Florencia y Roma, que era donde vivían sus grandes mecenas, la familia Médicis de Florencia, y los diferentes papas romanos. Fue el primer artista occidental del que se publicaron dos biografías en vida`,
	},

	{
		expect: "rus",
		buf:    `канадская панк-группа из города Аякс, Онтарио, основана в 1996 году. Текущий состав группы: Дерик Уибли (вокалист, гитара, клавишные), Джейсон МакКэслин (бас-гитара, бэк-вокал), Стив Джоз (ударные, бэк-вокал). С момента подписания контракта с лейблом Island Records в 1999 году группа выпустила 5 студийных альбомов, один концертный альбом, два концертных DVD и более 15 синглов. Суммарные продажи альбомов составили более 10 миллионов копий`,
	},

	{
		expect: "ron",
		buf:    `Atacul de noapte a fost o bătălie între Vlad Ţepeş, domnul Valahiei, şi sultanul Mehmed al II-lea al Imperiului Otoman, desfăşurată în apropierea cetăţii de scaun a Ţării Româneşti, Târgovişte, în noaptea de 17 iunie 1462. Campania a avut ca rezultat o victorie decisivă a valahilor. Conflictul a pornit iniţial de la refuzul lui Vlad de a plăti tribut otomanilor şi s-a amplificat după ce Vlad a invadat Bulgaria şi a tras în ţeapă peste 23.000 de turci şi bulgari. Mehmed a ridicat o armată uriaşă cu obiectivul de a cuceri Valahia şi a o anexa la imperiul său.`,
	},

	{
		expect: "heb",
		buf:    `בעיתונים ובכתבי עת שונים, והפך עד מהרה לתצלום המסמל את הניצחון במלחמה עבור הציבור האמריקני. בשנת 1945 הוא זכה בפרס פוליצר לתצלום הטוב ביותר, והיה לתצלום הראשון שזכה בפרס באותה שנה שבה פורסם לראשונה. התצלום היווה בסיס לשני סרטי קולנוע, חולות איוו ג'ימה וגיבורי הדגל, והיה מקור השראה לעוד מספר תצלומים אחרים.`,
	},
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestKWList(t *testing.T) {

	y := New("")
	y.LoadAllMem()

	for _, mytest := range Tests {

		result := y.IdentifyString(mytest.buf)
		if result[0].Lang != mytest.expect {
			l := min(10, len(mytest.buf))
			t.Errorf("FAILED test for lang %s buf %s...", mytest.expect, mytest.buf[0:l])
		}

		fmt.Printf("test [%s] score [%.4f]\n", mytest.expect, result[0].Score)
	}
}
