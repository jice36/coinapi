//curl -H 'Accept: application/json' -H "Authorization: Bearer {eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjY4ODM5NjMsImxvZ2luIjoibWFuIiwicGFzc3dvcmQiOiIxMjM0In0.0isCgND3bBHv-xPixZ1UAgGwTz64i5bTdCP6X8XPSAE}" http://127.0.0.1:8008/
//curl  -H 'Content-Type: application/json' --data '{"login":"man", "password":"1234"}' http://127.0.0.1:8008/login/
//curl  -H 'Content-Type: application/json' --data '{"id":"btc", "interval": 2}' http://127.0.0.1:8008/add/
//curl  -H 'Content-Type: application/json' --data '{"id":"btc"}' http://127.0.0.1:8008/delete/
//curl  -H 'Content-Type: application/json' --data '{"id":"btc}' http://127.0.0.1:8008/info/
package main

import (
    "fmt"
    "strings"
    "encoding/json"
    "log"
    "context"
    "time"
    "math/rand"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/buaazp/fasthttprouter"
    "github.com/valyala/fasthttp"
    "github.com/dgrijalva/jwt-go"
)

var TokenPassword = "paassswrod"
//структура для учётной записи пользователя
type User struct {
	Login string `json:"login"`
	Password string `json:"password"`
}
//структура для добавления валюты
type Currency struct{
  Id string
  Interval int
}
//структура для бд
type dbcurrency struct{
  Id string
}

func DeserializeJsonUser(ctx *fasthttp.RequestCtx) User{
  var user User
  err := json.Unmarshal(ctx.PostBody(), &user)
  if err != nil {
    log.Println(err)
  }
  return user
}

func DeserializeJsonCurrencyAdd(ctx *fasthttp.RequestCtx) Currency{
  var cur Currency
  err := json.Unmarshal(ctx.PostBody(), &cur)
  if err != nil {
    log.Println(err)
  }
  return cur
}

func DeserializeJsonCurrency(ctx *fasthttp.RequestCtx) dbcurrency{
  var cur dbcurrency
  err := json.Unmarshal(ctx.PostBody(), &cur)
  if err != nil {
    log.Println(err)
  }
  return cur
}

func main(){
  router := fasthttprouter.New()
  router.POST("/login/", LoginUser)
  router.POST("/add/", Auth(AddCurrency))
  router.POST("/delete/", Auth(DeleteCurrency))
  router.GET("/list/", Auth(ListCurrency))
  router.POST("/info/", Auth(CurrencyInfo))


  panic(fasthttp.ListenAndServe("127.0.0.1:8008", router.Handler))
}
//middleware
func Auth(next fasthttp.RequestHandler) fasthttp.RequestHandler{
  return func(ctx *fasthttp.RequestCtx) {

  tokenHeader := string(ctx.Request.Header.Peek("Authorization"))

  if tokenHeader == "" {
    ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
    return
  }

  splitted := strings.Split(tokenHeader, " ")
  if len(splitted) != 2 {
  ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
    return
  }

  tokenPart := splitted[1]


  token, _ := jwt.Parse(tokenPart, func(token *jwt.Token) (interface{}, error) {
      return []byte(TokenPassword), nil
  })

  if !token.Valid {
    ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
    return
  }
  next(ctx)
}
}
//создание токена
func CreateToken(login string, password string)(string, error){
  Claims := jwt.MapClaims{}
  Claims["login"] = login
  Claims["password"] = password
  Claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
  t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)
  token, err := t.SignedString([]byte(TokenPassword))
  if err != nil {
    return "", err
  }
  return token, nil
}
//добавление валюты
func AddCurrency(ctx *fasthttp.RequestCtx) {
  cur := DeserializeJsonCurrencyAdd(ctx)
  AddCurrencyInDD(cur.Id, cur.Interval)
}
//удаление валюты
func DeleteCurrency(ctx *fasthttp.RequestCtx) {
  cur := DeserializeJsonCurrency(ctx)
  DeleteCurrecncyInDB(cur.Id)
}
//список валюты
func ListCurrency(ctx *fasthttp.RequestCtx) {
  listcurrency(ctx)
}

func CurrencyInfo(ctx *fasthttp.RequestCtx) {
  cur := DeserializeJsonCurrency(ctx)
  message := map[string]interface{}{
    "id": cur.Id,
    "price":  string(rand.Int()),
  }

  data, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

  fmt.Fprintf(ctx, string(data))
}

func DeleteCurrecncyInDB(id string){
  client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
  if err != nil {
      log.Fatal(err)
  }

  err = client.Connect(context.TODO())
  if err != nil {
      log.Fatal(err)
  }

  err = client.Ping(context.TODO(), nil)
  if err != nil {
      log.Fatal(err)
  }

  collection := client.Database("currency").Collection("currency")

  filter := dbcurrency{id}
  _ , err = collection.DeleteMany(context.TODO(), filter)
  if err != nil {
    log.Fatal(err)
  }
}

func AddCurrencyInDD(id string, interval int){
  client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
  if err != nil {
      log.Fatal(err)
  }

  err = client.Connect(context.TODO())
  if err != nil {
      log.Fatal(err)
  }

  err = client.Ping(context.TODO(), nil)
  if err != nil {
      log.Fatal(err)
  }

  collection := client.Database("currency").Collection("currency")

  data := dbcurrency{id}

  insertResult, err := collection.InsertOne(context.TODO(), data)
  if err != nil {
      log.Fatal(err)
  }

  log.Println(insertResult.InsertedID)

}

func listcurrency(ctx *fasthttp.RequestCtx){
  client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
  if err != nil {
      log.Fatal(err)
  }

  err = client.Connect(context.TODO())
  if err != nil {
      log.Fatal(err)
  }

  err = client.Ping(context.TODO(), nil)
  if err != nil {
      log.Fatal(err)
  }

  collection := client.Database("currency").Collection("currency")

  options := options.Find()

  filter := bson.M{}

  var results []*dbcurrency

  cur, err := collection.Find(context.TODO(), filter, options)
  if err != nil {
    log.Fatal(err)
  }

  for cur.Next(context.TODO()) {
    var elem dbcurrency
    err := cur.Decode(&elem)
    if err != nil {
        log.Fatal(err)
    }
    results = append(results, &elem)
  }

  if err := cur.Err(); err != nil {
    log.Fatal(err)
  }

  cur.Close(context.TODO())

  log.Println(results)
}

func LoginUser(ctx *fasthttp.RequestCtx){
  user := DeserializeJsonUser(ctx)
  SearchUserInDB(user)
  token,err :=  CreateToken(user.Login, user.Password)
  if err != nil {
       log.Println("Не удалось получить токен ")
   }
  fmt.Fprintf(ctx, token)
}

func SearchUserInDB(user User){
  client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
  if err != nil {
    log.Fatal(err)
  }

  err = client.Connect(context.TODO())
  if err != nil {
    log.Fatal(err)
  }

  err = client.Ping(context.TODO(), nil)
  if err != nil {
    log.Fatal(err)
  }

  collection := client.Database("user").Collection("user")

  var result User
  filter := User{user.Login, user.Password}
  err = collection.FindOne(context.TODO(), filter).Decode(&result)
  if err != nil {
      log.Fatal(err)
  }

  log.Println(result)

}
