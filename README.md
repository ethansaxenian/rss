# RSS

A minimal RSS feed reader inspired by [Miniflux](https://github.com/miniflux/v2).

Built using Go, Templ, HTMX, and TailwindCSS.


### Requirements

- [Go](https://go.dev/)
- [Mise](https://mise.jdx.dev/)

### Running the app

Run `mise run dev` to run the development environment.
This includes hot-reloading of the app when `*.go`, `*.templ`, or `queries/*.sql` files change.

`mise run run` will run the app without hot-reloading.

`mise run build` will build the application binary (`./bin/main`).

Run `mise tasks` to view the full list of tasks.


