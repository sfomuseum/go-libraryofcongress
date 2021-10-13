# go-libraryofcongress

Go package providing tools for working with Library of Congress data.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-libraryofcongress.svg)](https://pkg.go.dev/github.com/sfomuseum/go-libraryofcongress)

## Tools

```
$> make cli
go build -mod vendor -o bin/parse-lcsh cmd/parse-lcsh/main.go
```

### parse-lcsh

`parse-lcsh` is a command-line tool to parse the Library of Congress Subject Headings (`lcsh.both.ndjson`) file and output CSV-encoded subject heading ID and (English) label data.

For example:

```
$> ./bin/parse-lcsh /usr/local/data/loc/lcsh.both.ndjson | less
id,label
sh98007138,Sports tournaments
sh85133899,Tennis--Tournaments
sh85133890,Tennis
sh91004781,Federation Cup
sh99005024,History
sh2009114899,Anarchism--Italy--History--20th century
sh2002012476,20th century
sh85004812,Anarchism
sh2008122899,Kitchens--Planning
sh85072576,Kitchens
sh2002006228,Planning
sh88001899,"Humorous poetry, Russian"
sh85116005,Russian poetry
sh85116022,Russian wit and humor
sh2008123899,Integrated circuits--Amateurs' manuals
sh99001292,Amateurs' manuals
sh85067117,Integrated circuits
sh85065604,Indians of South America--Ecuador--Antiquities
sh85040894,Ecuador--Antiquities
sh2005006899,Valdivian culture
... and so on
```

_Note: Subject headings with empty labels are ignored._

## See also

* https://id.loc.gov/index.html
* https://id.loc.gov/download/