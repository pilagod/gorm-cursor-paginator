gorm-cursor-paginator
[![Build Status](https://travis-ci.org/pilagod/gorm-cursor-paginator.svg?branch=master)](https://travis-ci.org/pilagod/gorm-cursor-paginator)
[![Coverage Status](https://coveralls.io/repos/github/pilagod/gorm-cursor-paginator/badge.svg?branch=master&kill_cache=1)](https://coveralls.io/github/pilagod/gorm-cursor-paginator?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/pilagod/gorm-cursor-paginator)](https://goreportcard.com/report/github.com/pilagod/gorm-cursor-paginator)
=====================

A paginator doing cursor-based pagination based on [GORM](https://github.com/go-gorm/gorm)

> This doc is for v2, which uses [GORM v2](https://github.com/go-gorm/gorm). If you are using [GORM v1](https://github.com/jinzhu/gorm), please checkout [v1 doc](https://github.com/pilagod/gorm-cursor-paginator/tree/v1).

Features
--------

- Query extendable.
- Multiple paging keys.
- Paging rule customization (e.g., order, SQL representation) for each key.
- GORM `column` tag supported.
- Error handling enhancement.
- Exporting `cursor` module for advanced usage.

Installation
------------

```sh
go get -u github.com/pilagod/gorm-cursor-paginator/v2
```

Usage By Example
----------------

Given a `User` model:

```go
type User struct {
    ID          int
    JoinedAt    time.Time `gorm:"column:created_at"`
}
```

We need to construct a `paginator.Paginator` based on fields of `User` struct. First we import `paginator`:

```go
import (
   "github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)
```

Then we can start configuring `paginator.Paginator`, here are some useful patterns:

```go
// configure paginator with paginator.Config and paginator.Option
func UserPaginator(
    cursor paginator.Cursor, 
    order *paginator.Order,
    limit *int,
) *paginator.Paginator {
    opts := []paginator.Option{
        &paginator.Config{
            // keys should be ordered by ordering priority
            Keys: []string{"ID", "JoinedAt"}, // default: []string{"ID"}
            Limit: 5, // default: 10
            Order: paginator.ASC, // default: DESC
        },
    }
    if limit != nil {
        opts = append(opts, paginator.WithLimit(*limit))
    }
    if order != nil {
        opts = append(opts, paginator.WithOrder(*order))
    }
    if cursor.After != nil {
        opts = append(opts, paginator.WithAfter(*cursor.After))
    }
    if cursor.Before != nil {
        opts = append(opts, paginator.WithBefore(*cursor.Before))
    }
    return paginator.New(opts...)
}

// configure paginator with setters
func UserPaginator(
    cursor paginator.Cursor,
    order *paginator.Order, 
    limit *int,
) *paginator.Paginator {
    p := paginator.New(
        paginator.WithKeys("ID", "JoinedAt"),
        paginator.WithLimit(5),
        paginator.WithOrder(paginator.ASC),
    )
    if order != nil {
        p.SetOrder(*order)
    }
    if limit != nil {
        p.SetLimit(*limit)
    }
    if cursor.After != nil {
        p.SetAfter(*cursor.After)
    }
    if cursor.Before != nil {
        p.SetBefore(*cursor.Before)
    }
    return p
}
```

If you need fine grained setting for each key, you can use `paginator.Rule`:

> `SQLRepr` is especially useful when you have `JOIN` or table alias in your SQL query. If `SQLRepr` is not specified, paginator will use the table name from paginated model, plus table key derived by below rules to form the SQL query:
>
> 1. Search GORM tag `column` on struct field.
> 2. If tag not found, convert struct field name to snake case.
>

```go
func UserPaginator(/* ... */) {
    opts := []paginator.Option{
        &paginator.Config{
            Rules: []paginator.Rule{
                {
                    Key: "ID",
                },
                {
                    Key: "JoinedAt",
                    Order: paginator.ASC,
                    SQLRepr: "users.created_at",
                },
            },
            Limit: 5,
            Order: paginator.DESC, // outer order will apply to keys without order specified, in this example is the key "ID".
        },
    }
    // ...
    return paginator.New(opts...)
}
```

After setup, you can start paginating with GORM:

```go
func FindUsers(db *gorm.DB, query Query) ([]User, paginator.Cursor, error) {
    var users []User

    // extend query before paginating
    stmt := db.
        Select(/* fields */).
        Joins(/* joins */).
        Where(/* queries */)

    // find users with pagination
    result, cursor, err := UserPaginator(/* config */).Paginate(stmt, &users)

    // this is paginator error, e.g., invalid cursor
    if err != nil {
        return nil, paginator.Cursor{}, err
    }

    // this is gorm error
    if result.Error != nil {
        return nil, paginator.Cursor{}, result.Error
    }

    return users, cursor, nil
}
```

The second value returned from `paginator.Paginator.Paginate` is a `paginator.Cursor`, which is a re-exported struct from `cursor.Cursor`:

```go
type Cursor struct {
    After  *string `json:"after" query:"after"`
    Before *string `json:"before" query:"before"`
}
```

That's all! Enjoy paginating in the GORM world. :tada:

> For more paginating examples, please checkout [exmaple/main.go](https://github.com/pilagod/gorm-cursor-paginator/blob/master/example/main.go) and [paginator/paginator_paginate_test.go](https://github.com/pilagod/gorm-cursor-paginator/blob/master/paginator/paginator_paginate_test.go)
>
> For manually encoding/decoding cursor exmaples, please checkout [cursor/encoding_test.go](https://github.com/pilagod/gorm-cursor-paginator/blob/master/cursor/encoding_test.go)

Known Issues
------------

1. Please make sure you're not paginating by nullable fields. Nullable values would occur [NULLS { FIRST | LAST } problems](https://learnsql.com/blog/how-to-order-rows-with-nulls/). Current workaround recommended is to select only non-null fields for paginating, or filter null values beforehand:

    ```go
    stmt = db.Where("nullable_field IS NOT NULL")
    ```

License
-------

© Cyan Ho (pilagod), 2018-NOW

Released under the [MIT License](https://github.com/pilagod/gorm-cursor-paginator/blob/master/LICENSE)
