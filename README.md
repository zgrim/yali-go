YALI port http://search.cpan.org/perldoc?Lingua%3A%3AYALI%3A%3ALanguageIdentifier

Lingua YALI (Yet Another Language Identifier)

cp /from/perl/src/share/*.gz data/

go-bindata -nocompress -nometadata -nomemcopy -pkg yali -prefix data data/

go build

y := yali.New("")
y.LoadAllMem()

or

y := yali.New("/from/disk/path")
y.LoadAllFS()

use

for i, lp := range y.IdentifyString("Foo bar has gone very baz") {
	fmt.Printf("#%d. Language %s = %.4f\n", i, lp.Lang, lp.Score)
}
