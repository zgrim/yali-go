#### Lingua::YALI (Yet Another Language Identifier)
#### ported from Perl 
http://search.cpan.org/perldoc?Lingua%3A%3AYALI%3A%3ALanguageIdentifier

(expects use of go-bindata to embed the resources, otherwise is easily adapted for fs)

```sh
cp /from/perl/src/share/*.gz data/
```

```sh
go-bindata -nocompress -nometadata -nomemcopy -pkg yali -prefix data data/
```

```sh
go build
```

```go
y := yali.New("")
y.LoadAllMem()
```

or

```go
y := yali.New("/from/disk/path")
y.LoadAllFS()
```

use

```go
for i, lp := range y.IdentifyString("Foo bar has gone very baz") {
	fmt.Printf("#%d. Language %s = %.4f\n", i, lp.Lang, lp.Score)
}
```
