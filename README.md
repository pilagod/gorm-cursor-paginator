gorm-cursor-paginator
[![Build Status](https://travis-ci.org/pilagod/gorm-cursor-paginator.svg?branch=master)](https://travis-ci.org/pilagod/gorm-cursor-paginator)
[![Coverage Status](https://coveralls.io/repos/github/pilagod/gorm-cursor-paginator/badge.svg?branch=master)](https://coveralls.io/github/pilagod/gorm-cursor-paginator?branch=master)
=====================

A paginator doing cursor-based pagination based on [GORM](https://github.com/jinzhu/gorm)

Usage by Example
----------------

Assume there is an optional query struct for paging:

```go
type PagingQuery struct {
    AfterCursor     *string
    BeforeCursor    *string
    Limit           *int
    Order           *string
}
```

and a GORM model:

```go
type Model struct {
    ID          uint
    CreatedAt   time.Time
}
```

You can simply build up a cursor paginator from the PagingQuery like:

```go
import (
    paginator "github.com/pilagod/gorm-cursor-paginator"
)

func InitModelPaginatorFrom(q PagingQuery) {
    p := paginator.New()

    p.SetKeys("CreatedAt", "ID") // [defualt: "ID"] (order of keys matters)

    if q.AfterCursor != nil {
        p.SetAfterCursor(*q.AfterCursor) // [default: ""]
    }

    if q.BeforeCursor != nil {
        p.SetBeforeCursor(*q.BeforeCursor) // [default: ""]
    }

    if q.Limit != nil {
        p.SetLimit(*q.Limit) // [default: 10]
    }

    if q.Order != nil {
        if *q.Order == "ascending" {
            p.SetOrder(paginator.ASC) // [default: paginator.DESC]
        }
    }
}
```

Then you can start to do pagination easily with GORM:

```go
func Find(db *gorm.DB, q PagingQuery) ([]Model, paginator.Cursors, error) {
    var models []Model

    stmt := db.Where(/* ... other filters ... */)
    stmt = db.Or(/* ... more other filters ... */)

    // init paginator for Model
    p := InitModelPaginatorFrom(q)

    // use GORM-like syntax to do pagination
    result := p.Paginate(stmt, &models)

    if result.Error != nil {
        // ...
    }
    // get cursors for next iteration
    cursors := p.GetNextCursors()

    return models, cursors, nil
}
```

After pagination, you can call `GetNextCursors()`, which returns a `Cursors` struct, to get cursors for next iteration:

```go
type Cursors struct {
    AfterCursor     string
    BeforeCursor    string
}
```

That's all ! Enjoy your paging in the GORM world :tada: