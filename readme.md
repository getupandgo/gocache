**Setup:**

Check out deployed version on 

http://34.65.220.116:8000/cache/top

or

`git clone`

`make up`


**How to interact:**

Sample requests to `gocache` are avialible in form of postman collection in `mocks` folder

**How to run tests:**

Tests depends on docker (for populating dev redis) and gomock. Commands below do all needed stuff and launching suite:

`make redis-dev`

`make test`