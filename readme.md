### Setup

Check out a [deployed app](http://34.65.220.116:8000/cache/top) or run locally via

`git clone https://github.com/getupandgo/gocache.git`

`make up`

### How to interact

Sample requests to app are available in form of postman collection in `mocks` folder, you can just import them. Requests marked with `prod` are pointing on deployed app.

### How to run tests

Tests depend on `docker` (for populating local redis instance) and `gomock`. Commands below do all needed stuff and running test suite

`make redis-dev`

`make test`