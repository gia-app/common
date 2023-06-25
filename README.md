# Common

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/gia-app/common/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/gia-app/common/tree/main)

reusable clients across different microservices.

### Requests
Inspired by the [imroc/req](https://github.com/imroc/req) library but I did not like how they supported testing.
Also, I was only making use of a very small subset of the features so I decided to create a proprietary lightweight version
as well as define a `Mock` client to be used for unit testing.

### MongoDB

Example usage
```golang
serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
db_uri := fmt.Sprintf("mongodb+srv://%v:%v@%v/?retryWrites=true&w=majority", DBUSER, DBPASSWORD, DBHOST)
clientOptions := options.Client().
    ApplyURI(db_uri).
    SetServerAPIOptions(serverAPIOptions)
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
client, err := mongo.Connect(ctx, clientOptions)
if err != nil {
    log.Fatal(err)
}

var mc storage.Manager = storage.NewMongoClient(client, client.Database(DBNAME))
```
