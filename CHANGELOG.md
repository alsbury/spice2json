# Changelog

## 0.3.2

* Bump dependencies

## 0.3.1

* Fix regression causing comments not to be included in output

## 0.3.0

* Add ability to read schema from stdin
* Add ability to read from spicedb rest endpoint
* Add ability to read from spicedb grpc endpoint
* Add -v option to print version and exit

## 0.2.0

* Add representation for relation types with caveats
* Add permission user set data with unions, intersections, and 
* Add basic representation of caveats
* Don't output empty fields
* Output to stdout if no output file specified
* Make namespace optional
* Delete comment start and end marks from comments
* Exit with non-zero return code on errors
* Upgrade to latest spicedb and go versions

## 0.1.0

* Update to spicedb 1.16.1
* Add support for allowed relations
* CI workflow based on Github actions

## 0.0.2

* Add readme and usage docs

## 0.0.1


Basic output of definitions with relations, permissions