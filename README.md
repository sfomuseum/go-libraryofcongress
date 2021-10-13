# go-libraryofcongress

Go package providing tools for working with Library of Congress data.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-libraryofcongress.svg)](https://pkg.go.dev/github.com/sfomuseum/go-libraryofcongress)

## Tools

```
$> make cli
go build -mod vendor -o bin/parse-lcsh cmd/parse-lcsh/main.go
```

### parse-lcnaf

`parse-lcname is a command-line tool to parse the Library of Congress `lcname.both.ndjson` (or `lcname.both.ndjson.zip`) file and output CSV-encoded subject heading ID and (English) label data.

For example:

```
> ./bin/parse-lcnaf ~/Downloads/lcnaf.both.ndjson.zip > lcnaf.csv

Time passes...
More time passes...
Time keeps on ticking ticking in to the future...
```

#### Notes

* Subject headings with empty labels are ignored.
* It is assumed that you have downloaded and uncompressed the [lcsh.both.ndjson](https://id.loc.gov/download) file from the Library of Congress' servers. Future releases may support fetching this file directly.
* This tool will work with the compressed and uncompressed version of `lcnaf.both.ndjson`. Keep in mind that compressed file is already 7GB and expands to an uncompressed 55GB.

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

#### Notes

* Subject headings with empty labels are ignored.
* It is assumed that you have downloaded and uncompressed the [lcsh.both.ndjson](https://id.loc.gov/download) file from the Library of Congress' servers. Future releases may support fetching this file directly.
* This tool will work with the compressed and uncompressed version of `lcsh.both.ndjson`.

## See also

* https://id.loc.gov/index.html
* https://id.loc.gov/download/