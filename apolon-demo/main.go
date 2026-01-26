package main

import (
	"fmt"
	"log"

	"github.com/jkeresman01/apolon/apolon"
	"github.com/jkeresman01/apolon/apolon-shared"
)

// Don't care
const connectionString = "postgres://Josip:Pa55w.rd@localhost:5432/apolon_db?sslmode=disable"

func main() {
	db, err := apolon.Open(connectionString)
	if err != nil {
		log.Fatal("failed to connect:", err)
	}
	defer db.Close()

	fmt.Println("Auto Migration")
	if err := db.AutoMigrate(&Patient{}); err != nil {
		log.Fatal("migration failed:", err)
	}
	fmt.Println("Table 'patients' created successfully")

	fmt.Println("\nInsert Patients")
	milica1 := &Patient{Name: "Milica1", Age: 25}
	milica2 := &Patient{Name: "Milica2", Age: 30}
	milica3 := &Patient{Name: "Milica3", Age: 17}

	db.Add(milica1)
	db.Add(milica2)
	db.Add(milica3)

	affected, err := db.SaveChanges()
	if err != nil {
		log.Fatal("insert failed:", err)
	}
	fmt.Printf("Inserted %d patients\n", affected)
	fmt.Printf("  Milica1 ID: %d\n", milica1.ID)
	fmt.Printf("  Milica2 ID: %d\n", milica2.ID)
	fmt.Printf("  Milica3 ID: %d\n", milica3.ID)

	fmt.Println("\nList All Patients")
	patients, err := apolon.Set[Patient](db).
		OrderBy(PatientFields.Name.Asc()).
		AsNoTracking().
		ToSlice()
	if err != nil {
		log.Fatal("query failed:", err)
	}
	for _, p := range patients {
		fmt.Printf("  ID=%d, Name=%s, Age=%d\n", p.ID, p.Name, p.Age)
	}

	fmt.Println("\nQuery: Age > 18")
	adults, err := apolon.Set[Patient](db).
		Where(PatientFields.Age.Gt(18)).
		AsNoTracking().
		ToSlice()
	if err != nil {
		log.Fatal("query failed:", err)
	}
	fmt.Printf("Found %d adults\n", len(adults))
	for _, p := range adults {
		fmt.Printf("  %s (%d)\n", p.Name, p.Age)
	}

	fmt.Println("\nQuery: Name contains 'Milica'")
	matched, err := apolon.Set[Patient](db).
		Where(PatientFields.Name.Contains("Milica")).
		AsNoTracking().
		ToSlice()
	if err != nil {
		log.Fatal("query failed:", err)
	}
	for _, p := range matched {
		fmt.Printf("  %s\n", p.Name)
	}

	fmt.Println("\n=== Query: OR condition (Age < 18 OR Age > 25) ===")
	result, err := apolon.Set[Patient](db).
		Where(shared.Or(
			PatientFields.Age.Lt(18),
			PatientFields.Age.Gt(25),
		)).
		AsNoTracking().
		ToSlice()
	if err != nil {
		log.Fatal("query failed:", err)
	}
	for _, p := range result {
		fmt.Printf("  %s (%d)\n", p.Name, p.Age)
	}

	fmt.Println("\nUpdate with Change Tracking")
	patient, err := apolon.Set[Patient](db).
		Where(PatientFields.Name.Eq("Milica1")).
		First()
	if err != nil {
		log.Fatal("query failed:", err)
	}
	fmt.Printf("Before: %s, Age=%d\n", patient.Name, patient.Age)

	patient.Age = 26
	db.ChangeTracker.DetectChanges()

	entry := db.Entry(patient)
	fmt.Printf("Changed: %v\n", entry.GetChangedProperties())

	affected, err = db.SaveChanges()
	if err != nil {
		log.Fatal("update failed:", err)
	}
	fmt.Printf("After: %s, Age=%d (rows affected: %d)\n", patient.Name, patient.Age, affected)

	// fmt.Println("\nDelete")
	// toDelete, _ := apolon.Set[Patient](db).
	// 	Where(PatientFields.Name.Eq("Milica3")).
	// 	First()
	// if toDelete != nil {
	// 	db.Remove(toDelete)
	// 	affected, _ = db.SaveChanges()
	// 	fmt.Printf("Deleted Milica3 (rows affected: %d)\n", affected)
	// }

	fmt.Println("\nGenerated SQL Examples")
	sql, args := apolon.Set[Patient](db).
		Where(PatientFields.Age.Gt(18)).
		Where(PatientFields.Name.Contains("Milica")).
		OrderBy(PatientFields.Age.Desc()).
		Limit(10).
		ToSQL()
	fmt.Printf("SQL:  %s\n", sql)
	fmt.Printf("Args: %v\n", args)

	fmt.Println("\nFinal State")
	final, _ := apolon.Set[Patient](db).
		OrderBy(PatientFields.ID.Asc()).
		AsNoTracking().
		ToSlice()
	for _, p := range final {
		fmt.Printf("  ID=%d, Name=%s, Age=%d\n", p.ID, p.Name, p.Age)
	}
}
