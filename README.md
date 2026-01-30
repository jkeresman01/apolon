<div align="center">

  <h1>apolon</h1>

  <h4>A type-safe ORM for Go inspired by Entity Framework</h4>
  <h6><i>Change tracking, fluent queries, and auto-migration with compile-time safety.</i></h6>

[![Go](https://img.shields.io/badge/Go-00ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1.svg?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)

</div>

## About

> [!NOTE]
> Still working on this, requires approximtlly additional 3 years of free time to finish

Apolon is trying to bring Entity Framework-style patterns to Go: change tracking, unit of work, and type-safe queries.

## Example

```go
type Patient struct {
    ID   int    `apolon:"id,pk"`
    Name string `apolon:"name,notnull"`
    Age  int    `apolon:"age"`
}

func main() {
    db, _ := apolon.Open("postgres://...")
    defer db.Close()

    // Auto-create tables
    db.AutoMigrate(&Patient{})

    // Type-safe queries with generated fields
    patients, _ := apolon.Set[Patient](db).
        Where(PatientFields.Age.Gt(18)).
        Where(PatientFields.Name.Contains("Smith")).
        OrderBy(PatientFields.Name.Asc()).
        ToSlice()

    // Change tracking
    patients[0].Name = "Updated"
    db.SaveChanges() // Only updates changed fields
}
```

## Design Decisions && comparison to other ORMs

### GORM is cool ... but has few problems

<h6><i>GORM is the most popular Go ORM (cool), but it relies heavily on strings and reflection at runtime: </i></h6>

```go
// GORM - strings everywhere, no compile-time safety - very uncool
db.Where("age > ?", 18).Where("name LIKE ?", "%Smith%").Find(&patients)

// Typo in column name? Runtime error.
db.Where("agee > ?", 18).Find(&patients) // Silent failure or runtime party
```

```go
// Apolon - compile-time type safety
apolon.Set[Patient](db).
    Where(PatientFields.Age.Gt(18)).
    Where(PatientFields.Name.Contains("Smith")).
    ToSlice()

// Typo? Compiler catches it immediately.
PatientFields.Agee.Gt(18) // Compile error: PatientFields has no field Agee
```

<h6><i>Generated field accessors catch errors at compile time, not in production.</i></h6>


### Ent is cool ... but is a different beast

<h6><i>Ent is a powerful entity framework backed by Meta. It takes a schema-first approach — you define fields in a DSL, and Ent generates everything (structs, builders, predicates) from that:</i></h6>

```go
// Ent - schema defines the model, structs are generated
func (Patient) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").NotEmpty().MaxLen(100),
        field.Int("age").Positive(),
    }
}

client.Patient.Create().SetName("John").SetAge(30).Save(ctx)

// Want to update two fields? Two separate calls.
client.Patient.UpdateOneID(id).SetName("Jane").Save(ctx)
client.Patient.UpdateOneID(id).SetAge(31).Save(ctx)
```

```go
// Apolon - your struct is the schema, changes are tracked
p, _ := apolon.Set[Patient](db).Find(id)

p.Name = "Robert MeDiro"
p.Age = 31
db.SaveChanges() // Both changes flushed in one transaction
```

<h6><i>Ent is stateless and operation-oriented — each write is an explicit builder
call with no memory of what came before. Apolon tracks your entities in memory
and flushes all changes at once via `SaveChanges()`, so you work with plain Go
structs instead of builder chains.</i></h6>

<h6><i>Ent also owns your types — it generates the structs, the predicates, and the
mutation API from its schema DSL. Apolon flips this: you own your structs, and
the ORM reads them via tags.</i></h6>

### SQLBoiler is cool ... but starts from the other end

<h6><i>SQLBoiler is a database-first ORM — it introspects your existing database
schema and generates fully type-safe models and query helpers from it:</i></h6>

```go
// SQLBoiler - database-first: schema lives in PostgreSQL, code is generated from it
// 1. You create the table in SQL (or a migration tool)
// 2. Run `sqlboiler psql` to generate Go code
// 3. Use the generated models and query mods

patients, _ := models.Patients(
    qm.Where("age > ?", 18),
    qm.Where("name LIKE ?", "%Smith%"),
    qm.OrderBy("name ASC"),
).All(ctx, db)

// Type-safe models, but queries still use raw strings for conditions.
// Typo in "agee > ?" — no compile error, runtime surprise.
```

```go
// Apolon - code-first(at least for now): your struct IS the schema
type Patient struct {
    ID   int    `apolon:"id,pk"`
    Name string `apolon:"name"`
    Age  int    `apolon:"age"`
}

// Table auto-created from struct tags, conditions are typed
apolon.Set[Patient](db).
    Where(PatientFields.Age.Gt(18)).
    Where(PatientFields.Name.Contains("Smith")).
    OrderBy(PatientFields.Name.Asc()).
    ToSlice()
```

<h6><i>SQLBoiler generates the models for you from the database — you don't write
structs, it does. Apolon goes the other way: you write plain Go structs with
tags, and the ORM derives the schema from them.</i></h6>

```go
// SQLBoiler - explicit updates, you specify every column
patient, _ := models.FindPatient(ctx, db, id)
patient.Name = "Robert MeDiro"
patient.Age = 81
patient.Update(ctx, db, boil.Whitelist("name", "age")) // Must list changed columns

// Apolon - change tracking figures it out
p, _ := apolon.Set[Patient](db).Find(id)
p.Name = "Robert MeDiro"
p.Age = 81
db.SaveChanges() // Detects and flushes only changed fields
```

<h6><i>If you're familiar with Entity Framework, Apolon should feel natural.
The main difference is using generated field types instead of lambda
expressions (Go doesn't have those - so we do a bit of magic).</i></h6>

## Getting Started

### Prerequisites

> [!IMPORTANT]
> Apolon currently only supports PostgreSQL.

### Installation

```bash
go get github.com/jkeresman01/apolon
```

### Generating Field Accessors

<h6><i>Apolon requires generating typed field accessors for your models. Add the `go:generate` directive to your model file:</i></h6>

```go
//go:generate go run github.com/jkeresman01/apolon/apolon-cli generate -i . -o .

type Patient struct {
    ID   int    `apolon:"id,pk"`
    Name string `apolon:"name"`
    Age  int    `apolon:"age"`
}
```

Then run:

```bash
go generate ./...
```

<h6><i>This creates a `model_fields.go` file with typed accessors like `PatientFields.Age`, `PatientFields.Name`, etc. that you can use in queries.</i></h6>

<h6><i>You can also run the generator manually:</i></h6>

```bash
go run github.com/jkeresman01/apolon/apolon-cli generate -i ./models -o ./models
```

### Resources ###

https://entgo.io/docs/schema-fields

https://gorm.io/docs/

