# project structure

* `design/` contains **Goa** DSL
* `api/` contains **Goa**'s generated example implementation
  * via `goa example github.com/bmedicke/pom/design -o api`
  * `gen/` contains **Goa**'s generated code
    * via `goa gen github.com/bmedicke/pom/design -o api`
      * should be rerun every time the design changes
    * `pom/` service endpoints and interface and service client
    * `http/` HTTP transport layer including *OpenAPI 3.0* spec
* `hooks/` contains default hook script

```sh
# from main project folder:
goa example github.com/bmedicke/pom/apidesign -o api # run once.
goa gen github.com/bmedicke/pom/apidesign -o api # repeat after apidesign change.

cd api/cmd/pom
go build
cd ../pom-cli
go build
```

---

```sh
goa gen github.com/bmedicke/pom/api/design -o api
goa example github.com/bmedicke/pom/api/design -o api
```
