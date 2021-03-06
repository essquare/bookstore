# Bookstore

Demo repository implementing a simple bookstore, which has only two entities: users and books.
The books are owned by users.
The authentication is implemented as custom endpoint, which issues JWT-Tokens.

The following endpoints are currently implemented:

- [POST] /authenticate
- [GET] /users
- [POST] /users
- [PUT] /users/{userID:[0-9]+}
- [DELETE] /users/{userID:[0-9]+}
- [GET] /users/{userID:[0-9]+}
- [GET] /users/{userID:[0-9]+}/books
- [POST] /users/{userID:[0-9]+}/books
- [PUT] /users/{userID:[0-9]+}/books/{bookID:[0-9]+}
- [DELETE] /users/{userID:[0-9]+}/books/{bookID:[0-9]+}

The last two do not need any authentication:

- [GET] /books - list all books
- [GET] /books/{bookID:[0-9]+} - get information about a book

## Building, Running, Testing

You'll need make and a recent go version with modules support.
Just building from source:

```bash
make
```

to migrate/create the database:

```bash
make migrate
```

to create the admin user:

```bash
ADMIN_USERNAME=admin ADMIN_PASSWORD=test123 make create_admin
```

to start the application:

```bash
make run
```

to run the tests:

```bash
make test
```

## Acknowledgments & Credits

- [List of contributors](https://github.com/essquare/bookstore/graphs/contributors)
- the structure of the project and some code parts are borrowed from [`miniflux`](https://github.com/miniflux/v2)
- distributed under Apache 2.0 License
