package main

import (
	"encoding/json"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pilagod/gorm-cursor-paginator/paginator"
)

// Product for product model
type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	// for gorm setup you can refer to: https://gorm.io/docs/#Quick-Start
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// reset product table
	db.Migrator().DropTable(&Product{})
	db.AutoMigrate(&Product{})

	// setup product with price 123
	db.Create(&Product{Code: "A", Price: 123})
	db.Create(&Product{Code: "B", Price: 123})
	db.Create(&Product{Code: "C", Price: 123})
	// setup product with price 456 for noise
	db.Create(&Product{Code: "D", Price: 456})
	db.Create(&Product{Code: "E", Price: 456})

	// for gorm query you can refer to: https://gorm.io/docs/query.html
	stmt := db.Where("price = ?", 123)

	// paginator comes in

	// page 1

	fmt.Println("===== page 1 - no cursor with limit 1 =====")

	p := paginator.New(&paginator.Config{
		Limit: 1,
	})

	var p1Products []Product

	result, p1Cursor, err := p.Paginate(stmt, &p1Products)
	// paginator error
	if err != nil {
		panic(err.Error())
	}
	// gorm error
	// https://gorm.io/docs/error_handling.html
	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("products:", toJSON(p1Products))
	fmt.Println("cursor:", toJSON(p1Cursor))

	// page 2

	fmt.Println("===== page 2 - use after cursor from page 1 =====")

	p = paginator.New(&paginator.Config{
		After: *p1Cursor.After,
	})

	var p2Products []Product

	result, p2Cursor, err := p.Paginate(stmt, &p2Products)
	if err != nil {
		panic(err.Error())
	}
	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("products:", toJSON(p2Products))
	fmt.Println("cursor:", toJSON(p2Cursor))

	// page 3

	fmt.Println("===== page 3 - use before cursor from page 2 =====")

	p = paginator.New(&paginator.Config{
		Before: *p2Cursor.Before,
	})

	var p3Products []Product

	result, p3Cursor, err := p.Paginate(stmt, &p3Products)
	if err != nil {
		panic(err.Error())
	}
	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("products:", toJSON(p3Products))
	fmt.Println("cursor:", toJSON(p3Cursor))
}

func toJSON(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic("parse json error")
	}
	return string(bytes)
}
