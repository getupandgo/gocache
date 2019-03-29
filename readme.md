**Setup:**

Check out a deployed app on 

http://34.65.220.116:8000/cache/top

or run locally via

`git clone`

`make up`


**How to interact:**

Sample requests to `gocache` are avialible in form of postman collection in `mocks` folder

**How to run tests:**

Tests depend on docker (for populating dev-redis) and gomock. Commands below do all needed stuff and launching suite:

`make redis-dev`

`make test`