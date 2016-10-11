package gomc

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/gorilla/context"
    //"github.com/dgrijalva/jwt-go"
    //"log"
    "encoding/json"
    "fmt"
    //"reflect"
    "strings"
    "github.com/garyburd/redigo/redis"
    "strconv"
)


type AppRequestInterface interface{
    ValidateRequestFunc(fn http.HandlerFunc) (http.HandlerFunc)
}

type AppRequest struct{}


type Route struct {
    Path string
    Handler func(http.ResponseWriter, *http.Request)
    Methods []string
    HeadersRegexp []string
    Headers []string
    ValidateRequest bool
    RateLimitRequest bool
    AppRequest AppRequestInterface
}

var Routes []Route
var Databases map[string]DatabaseConfig

func RegisterRoute(route Route){
    Routes = append(Routes, route)
}

func RegisterDatabases(dbs map[string]DatabaseConfig){
    Databases = dbs
}

const RequestValidUser string = "RequestValidUser"
const RequestOrganizationId string = "RequestOrganizationId"
const RequestApiKey string = "RequestApiKey"
const RequestRateLimitLimit string = "RequestRateLimitLimit"
const RequestRateLimitKey string = "RequestRateLimitKey"
const RequestRateLimitCurrent string = "RequestRateLimitCurrent"
const RequestRateLimitRemaining string = "RequestRateLimitRemaining"
const RequestRateLimitReset string = "RequestRateLimitReset"

func myLookupKey(key interface{}, r *http.Request) []uint8{

    apiKey := "iqcST9Au8h-KbKv2wKFCVEW2iEP8O30ln3V25ASNpX-sung-UTonNAZYMQUzpJjF"



    return []uint8(apiKey)
}

func ValidateRequestFunc(ari AppRequestInterface, fn http.HandlerFunc) (http.HandlerFunc){
    //fmt.Println("Hello1")
    return ari.ValidateRequestFunc(fn)
}

func (m AppRequest) ValidateRequestFunc(fn http.HandlerFunc) http.HandlerFunc {
    //fmt.Println("Hello")
    return func(w http.ResponseWriter, r *http.Request) {
        //fmt.Println("ValidateRequestFunc req kldahfgksdfhlkj")
        fn(w, r)
    }
}

func valReq(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        //fmt.Println("val req kldahfgksdfhlkj")
        /*
        //Default Headers
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("X-Rate-Limit-Limit", strconv.Itoa(context.Get(r, RequestRateLimitLimit).(int)))
        w.Header().Set("X-Rate-Limit-Remaining", strconv.Itoa(context.Get(r, RequestRateLimitRemaining).(int)))
        w.Header().Set("X-Rate-Limit-Reset", strconv.Itoa(context.Get(r, RequestRateLimitReset).(int)))

        //Return Error
        if(context.Get(r, RequestRateLimitRemaining).(int) > 0){
            fn(w, r)
        } else {

            repsonse := RequestErrorWrapper{
                Message : "API rate limit exceeded for " + context.Get(r, RequestRateLimitKey).(string) + ".",
            }
            w.WriteHeader(http.StatusForbidden)
            json.NewEncoder(w).Encode(repsonse)

        }

        //Clear Conext
        context.Clear(r)
        */
    }
}


func validateRequest(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        fn(w, r)
        /*
        tokenString := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Don't forget to validate the alg is what you expect:
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
            }

            return myLookupKey(token.Claims["sub"], r), nil
        })

        if err == nil && token.Valid && token != nil {
            context.Set(r, RequestValidUser, true)

            //Request
            fn(w, r)

        } else {

            w.Header().Set("Content-Type", "application/json")
            repsonse := RequestErrorWrapper{
                Message : "You are not authorized to perform this action",
            }
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(repsonse)
        }
        //Debug(token)
        //Debug(err)
        /*
        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            fmt.Println(claims["foo"], claims["nbf"])
        } else {
            fmt.Println(err)
        }
        */
        //fn(w, r)
        //Parse
        /*
        token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                //return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
            }
            return myLookupKey(token.Claims["sub"], r), nil
        })

        if err == nil && token.Valid && token != nil {
            context.Set(r, RequestValidUser, true)

            //Request
            fn(w, r)

        } else {

            w.Header().Set("Content-Type", "application/json")
            repsonse := RequestErrorWrapper{
                Message : "You are not authorized to perform this action",
            }
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(repsonse)
        }
        */
    }
}

func rateLimit(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        //IP
        clientIp := strings.Split(r.RemoteAddr,":")[0]

        //Redis Connect
        conn, err := redis.Dial("tcp", Config.Databases[Config.RateLimitDataUseDatabaseConfig].Host + ":" + Config.Databases[Config.RateLimitDataUseDatabaseConfig].Port)
        if err != nil {
            //fmt.Println(err)
        }
        defer conn.Close()

        //Valid User
        context.Set(r, RequestRateLimitLimit, Config.LimitNonUser)
        context.Set(r, RequestRateLimitKey, clientIp)
        if(context.Get(r, RequestValidUser).(bool)) {
            context.Set(r, RequestRateLimitLimit, Config.LimitUser)
            context.Set(r, RequestRateLimitKey, context.Get(r, RequestOrganizationId).(string))
        }

        rateLimitKey := context.Get(r, RequestRateLimitKey).(string)


        //Hours
        keyExists, _ := redis.Int(conn.Do("EXISTS", rateLimitKey + ":hour"))
        if keyExists == 0 {
            conn.Do("SET", rateLimitKey + ":hour", context.Get(r, RequestRateLimitLimit))
            conn.Do("EXPIRE", rateLimitKey + ":hour", 86400)
        }
        rateLimitCurrent, _ := redis.Int(conn.Do("GET", rateLimitKey + ":hour"))
        rateLimitRemaining := 0
        if(rateLimitCurrent > 0){
            rateLimitRemaining, _ = redis.Int(conn.Do("DECR", rateLimitKey + ":hour"))
        }
        rateLimitReset, _ := redis.Int(conn.Do("TTL", rateLimitKey + ":hour"))

        context.Set(r, RequestRateLimitCurrent, rateLimitCurrent)
        context.Set(r, RequestRateLimitRemaining, rateLimitRemaining)
        context.Set(r, RequestRateLimitReset, rateLimitReset)

        fn(w, r)
    }
}

func requestReset(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        context.Set(r, RequestOrganizationId, "")
        context.Set(r, RequestRateLimitKey, "")
        context.Set(r, RequestValidUser, false)
        context.Set(r, RequestRateLimitLimit, 0)
        context.Set(r, RequestRateLimitCurrent, 0)
        context.Set(r, RequestRateLimitRemaining, 1)
        context.Set(r, RequestRateLimitReset, 0)
        context.Set(r, RequestApiKey, 0)

        fn(w, r)
    }
}


func setResponse(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        //fmt.Println("set resp kldahfgksdfhlkj")

        //Default Headers
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("X-Rate-Limit-Limit", strconv.Itoa(context.Get(r, RequestRateLimitLimit).(int)))
        w.Header().Set("X-Rate-Limit-Remaining", strconv.Itoa(context.Get(r, RequestRateLimitRemaining).(int)))
        w.Header().Set("X-Rate-Limit-Reset", strconv.Itoa(context.Get(r, RequestRateLimitReset).(int)))

        //Return Error
        if(context.Get(r, RequestRateLimitRemaining).(int) > 0){
            fn(w, r)
        } else {

            repsonse := RequestErrorWrapper{
                Message : "API rate limit exceeded for " + context.Get(r, RequestRateLimitKey).(string) + ".",
            }
            w.WriteHeader(http.StatusForbidden)
            json.NewEncoder(w).Encode(repsonse)

        }

        //Clear Conext
        context.Clear(r)
    }
}

func TestRun(){
    fmt.Println("TEST RUN")
}

func Run(port string){

    //Databases
    if len(Databases) > 0 {
        for key, value := range Databases {
            if value.Type == "mongodb"{
                err, session := CreateMongoSession(value)
                if err != nil {

                } else {
                    value.MongoSession = session
                }
            }
            Databases[key] = value
        }
    }

    //Router
    router := mux.NewRouter().StrictSlash(true)
    if len(Routes) > 0 {
        for i := 0; i < len(Routes); i++ {
            handler := Routes[i].Handler
            /*
            if Routes[i].ValidateRequest {
                fmt.Println("YEP")
                //handler = ValidateRequestFunc(Routes[i].AppRequest, handler)
                //handler = ValidateRequestFunc(Routes[i].AppRequest, handler)
                handler = valReq(handler)
            }
            */

            handler = setResponse(handler)
            if Routes[i].RateLimitRequest {
                handler = rateLimit(handler)
            }
            if Routes[i].ValidateRequest {
                handler = ValidateRequestFunc(Routes[i].AppRequest, handler)
            }
            handler = requestReset(handler)
            router.HandleFunc(Routes[i].Path, handler).Methods(Routes[i].Methods...).HeadersRegexp(Routes[i].HeadersRegexp...).Headers(Routes[i].Headers...)
        }
    }
    fmt.Println("Running On " + port)
    fmt.Println(http.ListenAndServe(":" + port, router))

}